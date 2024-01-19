package services

import (
	"math/rand"
	"runtime"
)

type MerticsCollector struct {
	pollCount int
}

func NewMetricsCollector() *MerticsCollector {
	return &MerticsCollector{}
}

func (mc *MerticsCollector) Collect() *AgentMetrics {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	metrics := newAgentMetrics()

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

	metrics.Gauge.UpdateMetric("RandomValue", rand.Float64())

	metrics.Counter.UpdateMetric("PollCount", mc.pollCount)

	return metrics
}

func (mc *MerticsCollector) IncreasePollCount() {
	mc.pollCount++
}
