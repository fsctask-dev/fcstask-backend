package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type JoinOutcome string

const (
	JoinOutcomeSuccess        JoinOutcome = "success"
	JoinOutcomeAlreadyMember  JoinOutcome = "already_member"
	JoinOutcomeCourseNotFound JoinOutcome = "course_not_found"
	JoinOutcomeForbidden      JoinOutcome = "forbidden"
)

type CourseMetrics struct {
	JoinTotal          *prometheus.CounterVec
	GradeRecordedTotal *prometheus.CounterVec
}

func newCourseMetrics(reg prometheus.Registerer) *CourseMetrics {
	factory := promauto.With(reg)

	return &CourseMetrics{
		JoinTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "course",
				Name:      "join_total",
				Help:      "Total number of attempts to join a course.",
			},
			[]string{"outcome"},
		),
		GradeRecordedTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "grade",
				Name:      "recorded_total",
				Help:      "Total number of grade records persisted.",
			},
			[]string{"is_passed"},
		),
	}
}

func (m *CourseMetrics) IncJoin(outcome JoinOutcome) {
	if m == nil {
		return
	}
	m.JoinTotal.WithLabelValues(string(outcome)).Inc()
}

func (m *CourseMetrics) IncGradeRecorded(isPassed bool) {
	if m == nil {
		return
	}
	v := "false"
	if isPassed {
		v = "true"
	}
	m.GradeRecordedTotal.WithLabelValues(v).Inc()
}
