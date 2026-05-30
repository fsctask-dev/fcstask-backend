package service

import "fcstask-backend/internal/metrics"

func adminOutcome(err error) metrics.AdminOutcome {
	if err == nil {
		return metrics.AdminOutcomeSuccess
	}
	return metrics.AdminOutcomeError
}
