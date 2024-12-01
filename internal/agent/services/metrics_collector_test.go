package services

import (
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stretchr/testify/assert"
	"testing"
)

// MockMemoryProvider is a mock implementation of the MemoryProvider interface.
type MockMemoryProvider struct{}

func (m *MockMemoryProvider) VirtualMemory() (*mem.VirtualMemoryStat, error) {
	return &mem.VirtualMemoryStat{
		Total: 8589934592,
		Free:  4294967296,
	}, nil
}

// MockCPUProvider is a mock implementation of the CPUProvider interface.
type MockCPUProvider struct{}

func (m *MockCPUProvider) Percent(interval uint64, percpu bool) ([]float64, error) {
	return []float64{30.5, 40.3, 20.2}, nil
}

// Test NewMetricsCollector creates a new MetricsCollector instance
func TestNewMetricsCollector(t *testing.T) {
	memProvider := &MockMemoryProvider{}
	cpuProvider := &MockCPUProvider{}
	metricsCollector := NewMetricsCollector(memProvider, cpuProvider)

	assert.NotNil(t, metricsCollector)
	assert.Equal(t, 0, metricsCollector.pollCount) // Ensure initial poll count is 0
}

// Test Collect collects memory, CPU, and other metrics correctly
func TestCollect(t *testing.T) {
	memProvider := &MockMemoryProvider{}
	cpuProvider := &MockCPUProvider{}
	metricsCollector := NewMetricsCollector(memProvider, cpuProvider)

	metrics := metricsCollector.Collect()

	// Check if the metrics were collected correctly
	assert.NotNil(t, metrics)
	assert.Len(t, metrics.Gauge.GetAll(), 33)  // There should be 33 gauge metrics updated
	assert.Len(t, metrics.Counter.GetAll(), 1) // There should be 1 counter metric updated

	// Verify specific metrics
	gaugeMetrics := metrics.Gauge.GetAll()
	var totalMemoryFound bool
	var freeMemoryFound bool
	var cpuUtilizationFound bool
	for _, metric := range gaugeMetrics {
		if metric.Name == "TotalMemory" {
			totalMemoryFound = true
			assert.Equal(t, float64(8589934592), metric.Value)
		}
		if metric.Name == "FreeMemory" {
			freeMemoryFound = true
			assert.Equal(t, float64(4294967296), metric.Value)
		}
		if metric.Name == "CPUutilization0" {
			cpuUtilizationFound = true
			assert.Equal(t, 30.5, metric.Value)
		}

	}
	assert.True(t, totalMemoryFound, "TotalMemory found")
	assert.True(t, freeMemoryFound, "FreeMemory found")
	assert.True(t, cpuUtilizationFound, "CPUutilization0 found")

}
