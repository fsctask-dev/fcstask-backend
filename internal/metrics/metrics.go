package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Registry prometheus.Gatherer
	reg      *prometheus.Registry

	HTTP       *HTTPMetrics
	Auth       *AuthMetrics
	Session    *SessionMetrics
	Course     *CourseMetrics
	Admin      *AdminMetrics
	DB         *DBMetrics
	Checker    *CheckerMetrics
	LatePolicy *LatePolicyMetrics
}

func New() *Metrics {
	reg := NewRegistry()

	return &Metrics{
		Registry: reg,
		reg:      reg,

		HTTP:       newHTTPMetrics(reg),
		Auth:       newAuthMetrics(reg),
		Session:    newSessionMetrics(reg),
		Course:     newCourseMetrics(reg),
		Admin:      newAdminMetrics(reg),
		DB:         newDBMetrics(reg),
		Checker:    newCheckerMetrics(reg),
		LatePolicy: newLatePolicyMetrics(reg),
	}
}

func (m *Metrics) Registerer() prometheus.Registerer { return m.reg }
