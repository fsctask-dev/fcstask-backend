package health_monitor

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
)

type Config struct {
	Schedule string
	Timeout  time.Duration
}

func LoadConfig() Config {
	cf := Config{
		Schedule: "*/5 * * * *",
		Timeout:  5 * time.Second,
	}

	if sch := os.Getenv("HEALTH_SCHEDULE"); sch != "" {
		cf.Schedule = sch
	}

	if timeout := os.Getenv("HEALTH_TIMEOUT"); timeout != "" {
		if tout, err := time.ParseDuration(timeout); err == nil {
			cf.Timeout = tout
		}
	}

	return cf
}

type Monitor struct {
	url      string
	schedule string
	timeout  time.Duration
	logger   *log.Logger
	cron     *cron.Cron
}

func NewMonitor(url string) (*Monitor, error) {
	cf := LoadConfig()

	url = strings.TrimSuffix(url, "/")
	m := &Monitor{
		url:      url,
		schedule: cf.Schedule,
		timeout:  cf.Timeout,
		logger:   log.New(os.Stdout, "[health] ", log.LstdFlags),
		cron:     cron.New(),
	}

	return m, nil
}

func (m *Monitor) SetInterval(n int) {
	m.schedule = fmt.Sprintf("*/%d * * * *", n)
}

func (m *Monitor) SetTimeout(t time.Duration) {
	m.timeout = t
}

func (m *Monitor) CheckHealth() (bool, int, time.Duration, error) {
	start := time.Now()

	client := &http.Client{Timeout: m.timeout}
	resp, err := client.Get(m.url + "/health")
	duration := time.Since(start)
	if err != nil {
		return false, -1, duration, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return false, resp.StatusCode, duration, nil
	}

	return true, resp.StatusCode, duration, nil
}

func (m *Monitor) checkResult() {
	ok, code, duration, err := m.CheckHealth()
	if err != nil {
		m.logger.Printf("ERROR %v", err)
		return
	}
	if ok {
		m.logger.Printf("OK %v %v", code, duration)
		return
	}
	m.logger.Printf("ERROR Status Code: %v (%v)", code, duration)
}

func (m *Monitor) MonitorHealth() {
	_, err := m.cron.AddFunc(m.schedule, func() {
		m.checkResult()
	})
	if err != nil {
		m.logger.Printf("ERROR (cron failed) %v", err)
		return
	}

	m.cron.Start()
	m.logger.Println("Health monitoring started")
	go m.checkResult()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	exitSig := <-sigChan
	m.logger.Printf("exit sig %v", exitSig)

	m.cron.Stop()
	m.logger.Println("Stopped")
}
