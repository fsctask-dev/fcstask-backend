package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type DBRole string

const (
	DBRolePrimary DBRole = "primary"
	DBRoleReplica DBRole = "replica"
)

type DBMetrics struct {
	PoolOpen            *prometheus.GaugeVec
	PoolInUse           *prometheus.GaugeVec
	PoolIdle            *prometheus.GaugeVec
	WaitCount           *prometheus.GaugeVec
	WaitDurationSeconds *prometheus.GaugeVec
	MaxIdleClosed       *prometheus.GaugeVec
	MaxLifetimeClosed   *prometheus.GaugeVec
}

func newDBMetrics(reg prometheus.Registerer) *DBMetrics {
	factory := promauto.With(reg)
	labels := []string{"role", "replica_index"}

	return &DBMetrics{
		PoolOpen: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace, Subsystem: "db", Name: "pool_open",
			Help: "Number of established connections in the pool.",
		}, labels),
		PoolInUse: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace, Subsystem: "db", Name: "pool_in_use",
			Help: "Number of connections currently in use.",
		}, labels),
		PoolIdle: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace, Subsystem: "db", Name: "pool_idle",
			Help: "Number of idle connections in the pool.",
		}, labels),
		WaitCount: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace, Subsystem: "db", Name: "pool_wait_count_total",
			Help: "Total number of connections waited for.",
		}, labels),
		WaitDurationSeconds: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace, Subsystem: "db", Name: "pool_wait_duration_seconds_total",
			Help: "Total time blocked waiting for a new connection, in seconds.",
		}, labels),
		MaxIdleClosed: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace, Subsystem: "db", Name: "pool_max_idle_closed_total",
			Help: "Total connections closed due to SetMaxIdleConns.",
		}, labels),
		MaxLifetimeClosed: factory.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: Namespace, Subsystem: "db", Name: "pool_max_lifetime_closed_total",
			Help: "Total connections closed due to SetConnMaxLifetime.",
		}, labels),
	}
}
