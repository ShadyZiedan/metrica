package models

type Metric struct {
	Name    string
	gauge   float64
	counter int64
}

func (m *Metric) UpdateCounter(num int64) {
	m.counter += num
}

func (m *Metric) UpdateGauge(num float64) {
	m.gauge = num
}

func (m Metric) GetGauge() float64 {
	return m.gauge
}

func (m Metric) GetCounter() int64 {
	return m.counter
}
