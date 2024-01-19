package services

type AgentMetrics struct {
	Gauge   *GaugeCollection
	Counter *CounterCollection
}

type GaugeMetric struct {
	Name  string
	Value float64
}

type CounterMetric struct {
	Name  string
	Value int
}

type GaugeCollection struct {
	collection map[string]float64
}

type CounterCollection struct {
	collection map[string]int
}

func newAgentMetrics() *AgentMetrics {
	return &AgentMetrics{
		Gauge:   &GaugeCollection{collection: make(map[string]float64)},
		Counter: &CounterCollection{collection: make(map[string]int)},
	}
}

func (gc *GaugeCollection) UpdateMetric(name string, value float64) {
	gc.collection[name] = value
}

func (cc *CounterCollection) UpdateMetric(name string, value int) {
	cc.collection[name] = value
}

func (gc *GaugeCollection) GetAll() []GaugeMetric {
	gaugeMetrics := make([]GaugeMetric, 0, len(gc.collection))
	for name, val := range gc.collection {
		gaugeMetrics = append(gaugeMetrics, GaugeMetric{Name: name, Value: val})
	}
	return gaugeMetrics
}

func (cc *CounterCollection) GetAll() []CounterMetric {
	counterMetrics := make([]CounterMetric, 0, len(cc.collection))
	for name, val := range cc.collection {
		counterMetrics = append(counterMetrics, CounterMetric{Name: name, Value: val})
	}
	return counterMetrics
}
