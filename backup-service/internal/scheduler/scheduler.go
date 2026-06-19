package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"fcstask-backend/backup-service/internal/backup"
	"fcstask-backend/backup-service/internal/lock"
	"fcstask-backend/backup-service/internal/logging"
)

type Scheduler struct {
	schedule   string
	runInitial bool
	backuper   *backup.Backuper
	logger     *logging.Logger
	fileLock   *lock.FileLock
	cmdTimeout time.Duration

	cron    *cron.Cron
	jobMu   sync.Mutex
	running bool
	ctx     context.Context
	cancel  context.CancelFunc

	onResult func(error)
}

func (s *Scheduler) OnResult(f func(error)) { s.onResult = f }

func New(schedule string, runInitial bool, backuper *backup.Backuper, fileLock *lock.FileLock, cmdTimeout time.Duration, logger *logging.Logger) *Scheduler {
	return &Scheduler{
		schedule:   schedule,
		runInitial: runInitial,
		backuper:   backuper,
		fileLock:   fileLock,
		cmdTimeout: cmdTimeout,
		logger:     logger,
	}
}

func (s *Scheduler) Start(parent context.Context) error {
	s.ctx, s.cancel = context.WithCancel(parent)

	s.cron = cron.New()
	if _, err := s.cron.AddFunc(s.schedule, s.runJob); err != nil {
		s.cancel()
		return err
	}
	s.cron.Start()
	s.running = true
	s.logger.Info("Scheduler started with schedule %q", s.schedule)
	s.logNextRun()

	if s.runInitial {
		go s.runJob()
	}
	return nil
}

func (s *Scheduler) runJob() {

	if !s.jobMu.TryLock() {
		s.logger.Warn("Previous backup still running; skipping this scheduled trigger")
		return
	}
	defer s.jobMu.Unlock()

	if err := s.fileLock.TryAcquire(); err != nil {
		if errors.Is(err, lock.ErrLocked) {
			s.logger.Warn("Another process holds the backup lock; skipping run")
			return
		}
		s.logger.Error("Failed to acquire backup lock: %v", err)
		return
	}
	defer func() {
		if err := s.fileLock.Release(); err != nil {
			s.logger.Warn("Failed to release backup lock: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(s.ctx, s.cmdTimeout)
	defer cancel()

	start := time.Now()
	s.logger.Info("Starting scheduled backup")
	if _, err := s.backuper.CreateBackup(ctx); err != nil {
		s.logger.Error("Scheduled backup failed: %v", err)
		if s.onResult != nil {
			s.onResult(err)
		}
		return
	}
	s.logger.Info("Scheduled backup finished in %s", time.Since(start).Round(time.Second))
	if s.onResult != nil {
		s.onResult(nil)
	}
	s.logNextRun()
}

func (s *Scheduler) logNextRun() {
	if s.cron == nil {
		return
	}
	entries := s.cron.Entries()
	if len(entries) > 0 {
		s.logger.Info("Next backup scheduled at %s", entries[0].Next.Format(time.RFC3339))
	}
}

func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	if s.cancel != nil {
		s.cancel()
	}
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.running = false
	s.logger.Info("Scheduler stopped")
}
