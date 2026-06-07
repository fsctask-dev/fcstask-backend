package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type CheckerAction string

const (
	CheckerActionSubmitGrade CheckerAction = "submit_grade"
)

type CheckerMetrics struct {
	ActionTotal *prometheus.CounterVec
}

func newCheckerMetrics(reg prometheus.Registerer) *CheckerMetrics {
	factory := promauto.With(reg)
	return &CheckerMetrics{
		ActionTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "checker",
				Name:      "action_total",
				Help:      "Total number of checker actions.",
			},
			[]string{"action", "outcome"},
		),
	}
}

func (m *CheckerMetrics) IncAction(action CheckerAction, outcome AdminOutcome) {
	if m == nil {
		return
	}
	m.ActionTotal.WithLabelValues(string(action), string(outcome)).Inc()
}

type LatePolicyAction string

const (
	LatePolicyActionCreateOrUpdate LatePolicyAction = "create_or_update"
)

type LatePolicyMetrics struct {
	ActionTotal *prometheus.CounterVec
}

func newLatePolicyMetrics(reg prometheus.Registerer) *LatePolicyMetrics {
	factory := promauto.With(reg)
	return &LatePolicyMetrics{
		ActionTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "late_policy",
				Name:      "action_total",
			},
			[]string{"action", "outcome"},
		),
	}
}

func (m *LatePolicyMetrics) IncAction(action LatePolicyAction, outcome AdminOutcome) {
	if m == nil {
		return
	}
	m.ActionTotal.WithLabelValues(string(action), string(outcome)).Inc()
}
