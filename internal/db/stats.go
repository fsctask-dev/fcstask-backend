package db

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"fcstask-backend/internal/metrics"
)

func (c *Client) RunStatsCollector(ctx context.Context, m *metrics.DBMetrics, interval time.Duration) {
	if m == nil {
		return
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}

	snapshot := func() {
		if c.master != nil {
			if sqlDB, err := c.master.DB(); err == nil {
				publish(m, metrics.DBRolePrimary, 0, sqlDB.Stats())
			}
		}
		for i, r := range c.replicas {
			if r == nil {
				continue
			}
			if sqlDB, err := r.DB(); err == nil {
				publish(m, metrics.DBRoleReplica, i, sqlDB.Stats())
			}
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	snapshot()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot()
		}
	}
}

func publish(m *metrics.DBMetrics, role metrics.DBRole, replicaIdx int, s sql.DBStats) {
	r := string(role)
	idx := strconv.Itoa(replicaIdx)

	m.PoolOpen.WithLabelValues(r, idx).Set(float64(s.OpenConnections))
	m.PoolInUse.WithLabelValues(r, idx).Set(float64(s.InUse))
	m.PoolIdle.WithLabelValues(r, idx).Set(float64(s.Idle))
	m.WaitCount.WithLabelValues(r, idx).Set(float64(s.WaitCount))
	m.WaitDurationSeconds.WithLabelValues(r, idx).Set(s.WaitDuration.Seconds())
	m.MaxIdleClosed.WithLabelValues(r, idx).Set(float64(s.MaxIdleClosed))
	m.MaxLifetimeClosed.WithLabelValues(r, idx).Set(float64(s.MaxLifetimeClosed))
}
