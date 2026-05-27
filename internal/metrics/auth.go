package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type AuthOutcome string

const (
	AuthOutcomeSuccess          AuthOutcome = "success"
	AuthOutcomeInvalidCreds     AuthOutcome = "invalid_credentials"
	AuthOutcomeUserAlreadyExist AuthOutcome = "user_already_exists"
	AuthOutcomeInvalidInput     AuthOutcome = "invalid_input"
	AuthOutcomeInternalError    AuthOutcome = "internal_error"
)

type AuthMetrics struct {
	SignupsTotal  *prometheus.CounterVec
	SignInsTotal  *prometheus.CounterVec
	SignOutsTotal prometheus.Counter
}

func newAuthMetrics(reg prometheus.Registerer) *AuthMetrics {
	factory := promauto.With(reg)

	return &AuthMetrics{
		SignupsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "auth",
				Name:      "signups_total",
				Help:      "Total number of user sign-up attempts.",
			},
			[]string{"outcome"},
		),
		SignInsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "auth",
				Name:      "signins_total",
				Help:      "Total number of user sign-in attempts.",
			},
			[]string{"outcome"},
		),
		SignOutsTotal: factory.NewCounter(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "auth",
				Name:      "signouts_total",
				Help:      "Total number of successful sign-outs.",
			},
		),
	}
}

func (m *AuthMetrics) IncSignup(outcome AuthOutcome) {
	m.SignupsTotal.WithLabelValues(string(outcome)).Inc()
}

func (m *AuthMetrics) IncSignIn(outcome AuthOutcome) {
	m.SignInsTotal.WithLabelValues(string(outcome)).Inc()
}

func (m *AuthMetrics) IncSignOut() {
	m.SignOutsTotal.Inc()
}
