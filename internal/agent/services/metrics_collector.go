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

type AgentMetrics struct {
	Gauge   map[string]float64
	Counter map[string]int
}

func newAgentMetrics() *AgentMetrics {
	return &AgentMetrics{Gauge: make(map[string]float64), Counter: make(map[string]int)}
}

func (mc *MerticsCollector) Collect() *AgentMetrics {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	metrics := newAgentMetrics()

	metrics.Gauge["Alloc"] = float64(stats.Alloc)
	metrics.Gauge["BuckHashSys"] = float64(stats.BuckHashSys)
	metrics.Gauge["Frees"] = float64(stats.Frees)
	metrics.Gauge["GCCPUFraction"] = float64(stats.GCCPUFraction)
	metrics.Gauge["GCSys"] = float64(stats.GCSys)
	metrics.Gauge["HeapAlloc"] = float64(stats.HeapAlloc)
	metrics.Gauge["HeapIdle"] = float64(stats.HeapIdle)
	metrics.Gauge["HeapInuse"] = float64(stats.HeapInuse)
	metrics.Gauge["HeapObjects"] = float64(stats.HeapObjects)
	metrics.Gauge["HeapReleased"] = float64(stats.HeapReleased)
	metrics.Gauge["HeapSys"] = float64(stats.HeapSys)
	metrics.Gauge["LastGC"] = float64(stats.LastGC)
	metrics.Gauge["Lookups"] = float64(stats.Lookups)
	metrics.Gauge["MCacheInuse"] = float64(stats.MCacheInuse)
	metrics.Gauge["MCacheSys"] = float64(stats.MCacheSys)
	metrics.Gauge["Mallocs"] = float64(stats.Mallocs)
	metrics.Gauge["NextGC"] = float64(stats.NextGC)
	metrics.Gauge["NumForcedGC"] = float64(stats.NumForcedGC)
	metrics.Gauge["NumGC"] = float64(stats.NumGC)
	metrics.Gauge["OtherSys"] = float64(stats.OtherSys)
	metrics.Gauge["PauseTotalNs"] = float64(stats.PauseTotalNs)
	metrics.Gauge["StackInuse"] = float64(stats.StackInuse)
	metrics.Gauge["StackSys"] = float64(stats.StackSys)
	metrics.Gauge["Sys"] = float64(stats.Sys)
	metrics.Gauge["TotalAlloc"] = float64(stats.TotalAlloc)

	metrics.Gauge["RandomValue"] = rand.Float64()

	mc.increasePollCount()
	metrics.Counter["PollCount"] = mc.pollCount

	return metrics
}

func (mc *MerticsCollector) increasePollCount() {
	mc.pollCount++
}
