package models

type Metric struct {
	Name    string
	MType   string
	Gauge   float64
	Counter int64
}

func (m *Metric) UpdateCounter(num int64) {
	m.Counter += num
}

func (m *Metric) UpdateGauge(num float64) {
	m.Gauge = num
}
