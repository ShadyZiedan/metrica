package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/storage"
)

type mockMetricsRepository struct {
	metrics map[string]*models.Metric
}

func (m *mockMetricsRepository) FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error) {
	if metric, ok := m.metrics[name]; ok {
		return metric, nil
	}
	newMetric := &models.Metric{Name: name, MType: mType}
	m.metrics[name] = newMetric
	return newMetric, nil
}

func (m *mockMetricsRepository) FindAll(ctx context.Context) ([]*models.Metric, error) {
	var result []*models.Metric
	for _, metric := range m.metrics {
		result = append(result, metric)
	}
	return result, nil
}

func (m *mockMetricsRepository) Attach(observer storage.MetricsObserver) {}
func (m *mockMetricsRepository) Detach(observer storage.MetricsObserver) {}

func TestFileStorageService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_file_storage_service")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	metricsRepository := &mockMetricsRepository{metrics: make(map[string]*models.Metric)}
	fileStorageService := NewFileStorageService(metricsRepository, FileStorageServiceConfig{
		FileStoragePath: tempDir,
		StoreInterval:   0, //Sync mode
		Restore:         false,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go fileStorageService.Run(ctx)

	// Test saving and restoring metrics
	gauge := 10.5
	metric := &models.Metric{Name: "test_metric", MType: "gauge", Gauge: &gauge}
	metricsRepository.metrics[metric.Name] = metric

	// Wait for the metric to be saved to the file storage system
	time.Sleep(100 * time.Millisecond)

	fileStorageService.StopObserving()

	// Create a new file storage service to restore the metrics
	newFileStorageService := NewFileStorageService(metricsRepository, FileStorageServiceConfig{
		FileStoragePath: tempDir,
		StoreInterval:   0, //Sync mode
		Restore:         true,
	})

	go newFileStorageService.Run(ctx)

	// Wait for the metrics to be restored from the file storage system
	time.Sleep(100 * time.Millisecond)

	newMetric, ok := metricsRepository.metrics["test_metric"]
	if !ok {
		t.Error("Failed to restore metric from file storage system")
	}

	if newMetric.Name != metric.Name || newMetric.MType != metric.MType || newMetric.Gauge != metric.Gauge {
		t.Errorf("Restored metric does not match the original metric: %+v != %+v", newMetric, metric)
	}
}
