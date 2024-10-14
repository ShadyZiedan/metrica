// Package services contains the business logic for the agent metrics.
package services

// AgentMetrics represents the metrics collected by the agent.
type AgentMetrics struct {
	Gauge   *GaugeCollection
	Counter *CounterCollection
}

// GaugeMetric represents a gauge metric.
type GaugeMetric struct {
	Name  string
	Value float64
}

// CounterMetric represents a counter metric.
type CounterMetric struct {
	Name  string
	Value int
}

// GaugeCollection represents a collection of gauge metrics.
type GaugeCollection struct {
	collection map[string]float64
}

// CounterCollection represents a collection of counter metrics.
type CounterCollection struct {
	collection map[string]int
}

func NewAgentMetrics() *AgentMetrics {
	return &AgentMetrics{
		Gauge:   &GaugeCollection{collection: make(map[string]float64)},
		Counter: &CounterCollection{collection: make(map[string]int)},
	}
}

// UpdateMetric updates a gauge metric in the collection.
func (gc *GaugeCollection) UpdateMetric(name string, value float64) {
	gc.collection[name] = value
}

// UpdateMetric updates a counter metric in the collection.
func (cc *CounterCollection) UpdateMetric(name string, value int) {
	cc.collection[name] = value
}

// GetAll returns all gauge metrics in the collection.
func (gc *GaugeCollection) GetAll() []GaugeMetric {
	gaugeMetrics := make([]GaugeMetric, 0, len(gc.collection))
	for name, val := range gc.collection {
		gaugeMetrics = append(gaugeMetrics, GaugeMetric{Name: name, Value: val})
	}
	return gaugeMetrics
}

// GetAll returns all counter metrics in the collection.
func (cc *CounterCollection) GetAll() []CounterMetric {
	counterMetrics := make([]CounterMetric, 0, len(cc.collection))
	for name, val := range cc.collection {
		counterMetrics = append(counterMetrics, CounterMetric{Name: name, Value: val})
	}
	return counterMetrics
}
