package metrics

type BusinessMetrics struct {
	client MetricsClient
}

func NewBusinessMetrics(mc MetricsClient) *BusinessMetrics {
	return &BusinessMetrics{
		client: mc,
	}
}

func (bm *BusinessMetrics) TaskSubmitted(course, assignment string) {
	bm.client.Inc("tasks_submitted_total", map[string]string{"course": course, "assignment": assignment})
}

func (bm *BusinessMetrics) TaskCompleted(course, assignment string) {
	bm.client.Inc("tasks_completed_total", map[string]string{
		"course":     course,
		"assignment": assignment,
		"status":     "success"})
}

func (bm *BusinessMetrics) TaskFailed(course, assignment string) {
	bm.client.Inc("tasks_failed_total", map[string]string{
		"course":     course,
		"assignment": assignment,
		"status":     "failure",
		"reason":     "compilation_error",
	})
}

func (bm *BusinessMetrics) PipelineDuration(course string, seconds float64) {
	bm.client.Histogram("pipeline_duration_seconds", seconds,
		map[string]string{"course": course})
}
