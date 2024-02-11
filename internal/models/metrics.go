package models

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение Gauge или Counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи Counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи Gauge
}

func (m *Metrics) ParseMetricModel(model Metric) {
	m.ID = model.Name
	m.MType = model.MType
	switch model.MType {
	case "counter":
		m.Delta = &model.Counter
	case "gauge":
		m.Value = &model.Gauge
	}
}
