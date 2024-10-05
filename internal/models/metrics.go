package models

// Metrics represents a metric model request and response
type Metrics struct {
	// ID is the metric identifier or name of the metric
	ID string `json:"id"`
	// MType is the metric type (Gauge or Counter)
	MType string `json:"type"`
	// Delta is the value for Counter metrics (optional)
	Delta *int64 `json:"delta,omitempty"`
	// Value is the value for Gauge metrics (optional)
	Value *float64 `json:"value,omitempty"`
}

func (m *Metrics) ParseMetricModel(model *Metric) {
	m.ID = model.Name
	m.MType = model.MType
	switch model.MType {
	case "counter":
		m.Delta = model.Counter
	case "gauge":
		m.Value = model.Gauge
	}
}
