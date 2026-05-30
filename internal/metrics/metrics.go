package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Registry prometheus.Gatherer
	reg      *prometheus.Registry

	HTTP    *HTTPMetrics
	Auth    *AuthMetrics
	Session *SessionMetrics
	Course  *CourseMetrics
	Admin   *AdminMetrics
	DB      *DBMetrics
}

func New() *Metrics {
	reg := NewRegistry()

	return &Metrics{
		Registry: reg,
		reg:      reg,

		HTTP:    newHTTPMetrics(reg),
		Auth:    newAuthMetrics(reg),
		Session: newSessionMetrics(reg),
		Course:  newCourseMetrics(reg),
		Admin:   newAdminMetrics(reg),
		DB:      newDBMetrics(reg),
	}
}

func (m *Metrics) Registerer() prometheus.Registerer { return m.reg }
