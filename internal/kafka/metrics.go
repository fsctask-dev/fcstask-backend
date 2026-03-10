package kafka

import "time"

// Metric is a minimal JSON-friendly payload for downstream consumers.
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags,omitempty"`
}

func NewMetric(name string, value float64, tags map[string]string, timestamp time.Time) Metric {
	if tags == nil {
		tags = make(map[string]string)
	}
	
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}
	
	return Metric{
		Name:      name,
		Value:     value,
		Timestamp: timestamp,
		Tags:      tags,
	}
}
