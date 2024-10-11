package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shadyziedan/metrica/internal/models"
)

func TestFileStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "file_storage_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileName := filepath.Join(tempDir, "metrics.json")
	fs := NewFileStorage(fileName, Sync)

	counterValue := int64(10)
	gaugeValue := 20.5
	metrics := []*models.Metrics{
		{ID: "metric1", MType: "counter", Delta: &counterValue},
		{ID: "metric2", MType: "gauge", Value: &gaugeValue},
	}

	for _, metric := range metrics {
		err = fs.SaveMetric(metric)
		if err != nil {
			t.Errorf("failed to save metric: %v", err)
		}
	}

	// wait for all to ba saved
	time.Sleep(100 * time.Millisecond)

	savedMetrics, err := fs.ReadMetrics()
	if err != nil {
		t.Errorf("failed to read metrics: %v", err)
	}

	if len(savedMetrics) != len(metrics) {
		t.Errorf("expected %d metrics, got %d", len(metrics), len(savedMetrics))
	}

	for i, savedMetric := range savedMetrics {
		expectedMetric := metrics[i]
		if savedMetric.ID != expectedMetric.ID ||
			savedMetric.MType != expectedMetric.MType {
			t.Errorf("expected %+v, got %+v", expectedMetric, savedMetric)
		}
		if savedMetric.MType == "counter" && *savedMetric.Delta != *expectedMetric.Delta {
			t.Errorf("expected %d, got %d", *expectedMetric.Delta, *savedMetric.Delta)
		}
		if savedMetric.MType == "gauge" && *savedMetric.Value != *expectedMetric.Value {
			t.Errorf("expected %f, got %f", *expectedMetric.Value, *savedMetric.Value)
		}
	}
}

func BenchmarkFileStorageSave(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "file_storage_benchmark")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileName := filepath.Join(tempDir, "metrics.json")
	fs := NewFileStorage(fileName, Normal)

	delta := int64(10)
	metric := &models.Metrics{ID: "benchmark", MType: "counter", Delta: &delta}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		err := fs.SaveMetric(metric)
		if err != nil {
			b.Fatalf("failed to save metric: %v", err)
		}
	}
}

func BenchmarkFileStorageRead(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "file_storage_benchmark")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileName := filepath.Join(tempDir, "metrics.json")
	fs := NewFileStorage(fileName, Normal)

	delta := int64(10)
	value := 20.5
	metrics := []*models.Metrics{
		{ID: "benchmark1", MType: "counter", Delta: &delta},
		{ID: "benchmark2", MType: "gauge", Value: &value},
	}

	for _, metric := range metrics {
		err := fs.SaveMetric(metric)
		if err != nil {
			b.Fatalf("failed to save metric: %v", err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := fs.ReadMetrics()
		if err != nil {
			b.Fatalf("failed to read metrics: %v", err)
		}
	}
}
