package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PasswordResetMetrics tracks the password-reset flow. The "outcome" label is
// the service error code (success, not_found, unauthorized, internal_error, …).
type PasswordResetMetrics struct {
	RequestsTotal *prometheus.CounterVec
	ResendsTotal  *prometheus.CounterVec
	ConfirmsTotal *prometheus.CounterVec
}

func newPasswordResetMetrics(reg prometheus.Registerer) *PasswordResetMetrics {
	factory := promauto.With(reg)

	return &PasswordResetMetrics{
		RequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "password_reset",
				Name:      "requests_total",
				Help:      "Total number of password-reset requests.",
			},
			[]string{"outcome"},
		),
		ResendsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "password_reset",
				Name:      "resends_total",
				Help:      "Total number of password-reset code resends.",
			},
			[]string{"outcome"},
		),
		ConfirmsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "password_reset",
				Name:      "confirms_total",
				Help:      "Total number of password-reset confirmation attempts.",
			},
			[]string{"outcome"},
		),
	}
}

func (m *PasswordResetMetrics) IncRequest(outcome string) {
	if m == nil {
		return
	}
	m.RequestsTotal.WithLabelValues(outcome).Inc()
}

func (m *PasswordResetMetrics) IncResend(outcome string) {
	if m == nil {
		return
	}
	m.ResendsTotal.WithLabelValues(outcome).Inc()
}

func (m *PasswordResetMetrics) IncConfirm(outcome string) {
	if m == nil {
		return
	}
	m.ConfirmsTotal.WithLabelValues(outcome).Inc()
}
