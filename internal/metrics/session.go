package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SessionRevokeReason string

const (
	SessionRevokeReasonSignOut SessionRevokeReason = "signout"
	SessionRevokeReasonExpired SessionRevokeReason = "expired"
	SessionRevokeReasonAdmin   SessionRevokeReason = "admin_revoke"
)

type SessionMetrics struct {
	CreatedTotal        prometheus.Counter
	RevokedTotal        *prometheus.CounterVec
	CleanupDeletedTotal prometheus.Counter
	CleanupErrorsTotal  prometheus.Counter
}

func newSessionMetrics(reg prometheus.Registerer) *SessionMetrics {
	factory := promauto.With(reg)

	return &SessionMetrics{
		CreatedTotal: factory.NewCounter(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "session",
				Name:      "created_total",
				Help:      "Total number of sessions created.",
			},
		),
		RevokedTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "session",
				Name:      "revoked_total",
				Help:      "Total number of sessions revoked.",
			},
			[]string{"reason"},
		),
		CleanupDeletedTotal: factory.NewCounter(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "session",
				Name:      "cleanup_deleted_total",
				Help:      "Total number of outdated sessions removed by background cleanup.",
			},
		),
		CleanupErrorsTotal: factory.NewCounter(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "session",
				Name:      "cleanup_errors_total",
				Help:      "Total number of background session cleanup errors.",
			},
		),
	}
}

func (m *SessionMetrics) IncCreated() {
	if m == nil {
		return
	}
	m.CreatedTotal.Inc()
}

func (m *SessionMetrics) IncRevoked(reason SessionRevokeReason) {
	if m == nil {
		return
	}
	m.RevokedTotal.WithLabelValues(string(reason)).Inc()
}

func (m *SessionMetrics) AddCleanupDeleted(n int64) {
	if m == nil {
		return
	}
	m.CleanupDeletedTotal.Add(float64(n))
}

func (m *SessionMetrics) IncCleanupError() {
	if m == nil {
		return
	}
	m.CleanupErrorsTotal.Inc()
}
