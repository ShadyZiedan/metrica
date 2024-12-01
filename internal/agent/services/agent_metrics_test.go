package services

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test NewAgentMetrics creates a new AgentMetrics instance with empty collections
func TestNewAgentMetrics(t *testing.T) {
	agentMetrics := NewAgentMetrics()

	assert.NotNil(t, agentMetrics)
	assert.NotNil(t, agentMetrics.Gauge)
	assert.NotNil(t, agentMetrics.Counter)
	assert.Empty(t, agentMetrics.Gauge.collection)
	assert.Empty(t, agentMetrics.Counter.collection)
}

// Test UpdateMetric for GaugeCollection updates the gauge correctly
func TestUpdateGaugeMetric(t *testing.T) {
	agentMetrics := NewAgentMetrics()

	// Update a new gauge metric
	agentMetrics.Gauge.UpdateMetric("temperature", 23.5)

	// Check that the metric was correctly updated
	gaugeMetrics := agentMetrics.Gauge.GetAll()
	assert.Len(t, gaugeMetrics, 1)
	assert.Equal(t, "temperature", gaugeMetrics[0].Name)
	assert.Equal(t, 23.5, gaugeMetrics[0].Value)
}

// Test UpdateMetric for CounterCollection updates the counter correctly
func TestUpdateCounterMetric(t *testing.T) {
	agentMetrics := NewAgentMetrics()

	// Update a new counter metric
	agentMetrics.Counter.UpdateMetric("requests", 100)

	// Check that the metric was correctly updated
	counterMetrics := agentMetrics.Counter.GetAll()
	assert.Len(t, counterMetrics, 1)
	assert.Equal(t, "requests", counterMetrics[0].Name)
	assert.Equal(t, 100, counterMetrics[0].Value)
}

// Test GetAll for GaugeCollection returns all gauge metrics
func TestGetAllGaugeMetrics(t *testing.T) {
	agentMetrics := NewAgentMetrics()

	// Add some gauge metrics
	agentMetrics.Gauge.UpdateMetric("temperature", 23.5)
	agentMetrics.Gauge.UpdateMetric("humidity", 50.0)

	// Retrieve all gauge metrics and verify
	gaugeMetrics := agentMetrics.Gauge.GetAll()
	assert.Len(t, gaugeMetrics, 2)
	var tempFound bool
	var humidityFound bool
	for _, gaugeMetric := range gaugeMetrics {
		if gaugeMetric.Name == "temperature" {
			tempFound = true
			assert.Equal(t, float64(23.5), gaugeMetric.Value)
		}
		if gaugeMetric.Name == "humidity" {
			humidityFound = true
			assert.Equal(t, float64(50.0), gaugeMetric.Value)
		}
	}
	assert.True(t, tempFound)
	assert.True(t, humidityFound)
}

// Test GetAll for CounterCollection returns all counter metrics
func TestGetAllCounterMetrics(t *testing.T) {
	agentMetrics := NewAgentMetrics()

	// Add some counter metrics
	agentMetrics.Counter.UpdateMetric("requests", 100)
	agentMetrics.Counter.UpdateMetric("errors", 5)

	// Retrieve all counter metrics and verify
	counterMetrics := agentMetrics.Counter.GetAll()
	assert.Len(t, counterMetrics, 2)
	var requestFound bool
	var errorFound bool
	for _, counterMetric := range counterMetrics {
		if counterMetric.Name == "requests" {
			requestFound = true
			assert.Equal(t, 100, counterMetric.Value)
		}
		if counterMetric.Name == "errors" {
			errorFound = true
			assert.Equal(t, 5, counterMetric.Value)
		}
	}
	assert.True(t, requestFound)
	assert.True(t, errorFound)
}
