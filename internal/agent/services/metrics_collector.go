package services

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// MetricsCollector collects metrics from various sources and provides a way to access them.
type MerticsCollector struct {
	pollCount int
}

// NewMetricsCollector creates a new instance of MetricsCollector.
func NewMetricsCollector() *MerticsCollector {
	return &MerticsCollector{}
}

// Collect collects metrics from various sources and returns them.
func (mc *MerticsCollector) Collect() *AgentMetrics {
	metrics := newAgentMetrics()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	metrics.Gauge.UpdateMetric("Alloc", float64(stats.Alloc))
	metrics.Gauge.UpdateMetric("BuckHashSys", float64(stats.BuckHashSys))
	metrics.Gauge.UpdateMetric("Frees", float64(stats.Frees))
	metrics.Gauge.UpdateMetric("GCCPUFraction", float64(stats.GCCPUFraction))
	metrics.Gauge.UpdateMetric("GCSys", float64(stats.GCSys))
	metrics.Gauge.UpdateMetric("HeapAlloc", float64(stats.HeapAlloc))
	metrics.Gauge.UpdateMetric("HeapIdle", float64(stats.HeapIdle))
	metrics.Gauge.UpdateMetric("HeapInuse", float64(stats.HeapInuse))
	metrics.Gauge.UpdateMetric("HeapObjects", float64(stats.HeapObjects))
	metrics.Gauge.UpdateMetric("HeapReleased", float64(stats.HeapReleased))
	metrics.Gauge.UpdateMetric("HeapSys", float64(stats.HeapSys))
	metrics.Gauge.UpdateMetric("LastGC", float64(stats.LastGC))
	metrics.Gauge.UpdateMetric("Lookups", float64(stats.Lookups))
	metrics.Gauge.UpdateMetric("MCacheInuse", float64(stats.MCacheInuse))
	metrics.Gauge.UpdateMetric("MCacheSys", float64(stats.MCacheSys))
	metrics.Gauge.UpdateMetric("Mallocs", float64(stats.Mallocs))
	metrics.Gauge.UpdateMetric("NextGC", float64(stats.NextGC))
	metrics.Gauge.UpdateMetric("NumForcedGC", float64(stats.NumForcedGC))
	metrics.Gauge.UpdateMetric("NumGC", float64(stats.NumGC))
	metrics.Gauge.UpdateMetric("OtherSys", float64(stats.OtherSys))
	metrics.Gauge.UpdateMetric("PauseTotalNs", float64(stats.PauseTotalNs))
	metrics.Gauge.UpdateMetric("StackInuse", float64(stats.StackInuse))
	metrics.Gauge.UpdateMetric("StackSys", float64(stats.StackSys))
	metrics.Gauge.UpdateMetric("Sys", float64(stats.Sys))
	metrics.Gauge.UpdateMetric("TotalAlloc", float64(stats.TotalAlloc))
	metrics.Gauge.UpdateMetric("MSpanInuse", float64(stats.MSpanInuse))
	metrics.Gauge.UpdateMetric("MSpanSys", float64(stats.MSpanSys))

	metrics.Gauge.UpdateMetric("RandomValue", rand.Float64())

	metrics.Counter.UpdateMetric("PollCount", mc.pollCount)

	// collecting virtual memory info
	memory, err := mem.VirtualMemory()
	if err == nil {
		metrics.Gauge.UpdateMetric("TotalMemory", float64(memory.Total))
		metrics.Gauge.UpdateMetric("FreeMemory", float64(memory.Free))
	}

	// collecting cpu utilization info
	cpuUsages, err := cpu.Percent(0, true)
	if err == nil {
		for i, cpuUsage := range cpuUsages {
			metrics.Gauge.UpdateMetric(fmt.Sprintf("CPUutilization%d", i), cpuUsage)
		}
	}

	return metrics
}

// IncreasePollCount increases the poll count in the agent metrics.
func (mc *MerticsCollector) IncreasePollCount() {
	mc.pollCount++
}
