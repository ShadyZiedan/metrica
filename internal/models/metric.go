// Package models contains the data structures and methods for working with metric data.
package models

// Metric represents a metric record in the database
type Metric struct {
	// Name is the name of the metric
	Name string
	// MType is the type of the metric (Gauge or Counter)
	MType string
	// Gauge is the current value of the Gauge metric, if applicable
	Gauge *float64
	// Counter is the current value of the Counter metric, if applicable
	Counter *int64
}

// UpdateCounter increments the Counter value by the given value
func (m *Metric) UpdateCounter(num int64) {
	if m.Counter == nil {
		m.Counter = new(int64)
	}
	*m.Counter += num
}

// UpdateGauge replaces the current value of the Gauge metric
func (m *Metric) UpdateGauge(num float64) {
	if m.Gauge == nil {
		m.Gauge = new(float64)
	}
	*m.Gauge = num
}

func NewCounterMetric(name string, value int64) *Metric {
	return &Metric{Name: name, MType: "counter", Counter: &value}
}

func NewGaugeMetric(name string, value float64) *Metric {
	return &Metric{Name: name, MType: "gauge", Gauge: &value}
}
