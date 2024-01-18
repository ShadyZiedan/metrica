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

func newCounterCollection() *CounterCollection {
	return &CounterCollection{collection: make(map[string]int)}
}

func newGaugeCollection() *GaugeCollection {
	return &GaugeCollection{collection: make(map[string]float64)}
}

func newAgentMetrics() *AgentMetrics {
	return &AgentMetrics{Gauge: newGaugeCollection(), Counter: newCounterCollection()}
}

func (gc *GaugeCollection) UpdateMetric(name string, value float64) {
	gc.collection[name] = value
}

func (cc *CounterCollection) UpdateMetric(name string, value int) {
	cc.collection[name]++
}

func (gc *GaugeCollection) GetAll() []GaugeMetric {
	var gaugeMetrics []GaugeMetric
	for name, val := range gc.collection {
		gaugeMetrics = append(gaugeMetrics, GaugeMetric{Name: name, Value: val})
	}
	return gaugeMetrics
}

func (cc *CounterCollection) GetAll() []CounterMetric {
	var counterMetrics []CounterMetric
	for name, val := range cc.collection {
		counterMetrics = append(counterMetrics, CounterMetric{Name: name, Value: val})
	}
	return counterMetrics
}
