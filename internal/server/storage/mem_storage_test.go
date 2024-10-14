package storage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shadyziedan/metrica/internal/models"
)

type mockObserver struct {
	notifyCalled bool
	metric       *models.Metric
	err          error
}

func (m *mockObserver) Notify(metric *models.Metric) error {
	m.notifyCalled = true
	m.metric = metric
	return m.err
}

func TestMemStorage_Create(t *testing.T) {
	storage := NewMemStorage()

	// Test creating a new metric
	err := storage.Create(context.Background(), "metric1", "gauge")
	require.NoError(t, err)

	// Test creating a duplicate metric
	err = storage.Create(context.Background(), "metric1", "gauge")
	assert.EqualError(t, err, "metric has been already created")
}

func TestMemStorage_Find(t *testing.T) {
	storage := NewMemStorage()
	storage.Create(context.Background(), "metric1", "gauge")

	metric, err := storage.Find(context.Background(), "metric1")
	require.NoError(t, err)
	assert.Equal(t, "metric1", metric.Name)

	// Test finding a non-existent metric
	_, err = storage.Find(context.Background(), "metric2")
	assert.EqualError(t, err, "metric not found")
}

func TestMemStorage_FindAll(t *testing.T) {
	storage := NewMemStorage()
	storage.Create(context.Background(), "metric1", "gauge")
	storage.Create(context.Background(), "metric2", "counter")

	metrics, err := storage.FindAll(context.Background())
	require.NoError(t, err)
	assert.Len(t, metrics, 2)
}

func TestMemStorage_FindOrCreate(t *testing.T) {
	storage := NewMemStorage()
	metric, err := storage.FindOrCreate(context.Background(), "metric1", "gauge")
	require.NoError(t, err)
	assert.Equal(t, "metric1", metric.Name)

	// Test finding an existing metric
	metric, err = storage.FindOrCreate(context.Background(), "metric1", "gauge")
	require.NoError(t, err)
	assert.Equal(t, "metric1", metric.Name)
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()
	storage.Create(context.Background(), "metric1", "gauge")

	// Update counter
	err := storage.UpdateCounter(context.Background(), "metric1", 5)
	require.NoError(t, err)

	metric, err := storage.Find(context.Background(), "metric1")
	require.NoError(t, err)
	assert.Equal(t, "counter", metric.MType)

	// Test updating a non-existent metric
	err = storage.UpdateCounter(context.Background(), "metric2", 5)
	assert.EqualError(t, err, "metric not found")
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()
	storage.Create(context.Background(), "metric1", "gauge")

	// Update gauge
	err := storage.UpdateGauge(context.Background(), "metric1", 10.5)
	require.NoError(t, err)

	metric, err := storage.Find(context.Background(), "metric1")
	require.NoError(t, err)
	assert.Equal(t, "gauge", metric.MType)

	// Test updating a non-existent metric
	err = storage.UpdateGauge(context.Background(), "metric2", 10.5)
	assert.EqualError(t, err, "metric not found")
}

func TestMemStorage_FindAllByName(t *testing.T) {
	storage := NewMemStorage()
	err := storage.Create(context.Background(), "metric123", "gauge")
	require.NoError(t, err)
	err = storage.Create(context.Background(), "metric234", "counter")
	require.NoError(t, err)

	// Test finding metrics by name
	metrics, err := storage.FindAllByName(context.Background(), []string{"metric123", "metric234"})
	require.NoError(t, err)
	assert.Len(t, metrics, 2)

	// Test finding a metric that does not exist
	metrics, err = storage.FindAllByName(context.Background(), []string{"metric3"})
	require.NoError(t, err)
	assert.Len(t, metrics, 0)
}

func TestMemStorage_AttachDetach(t *testing.T) {
	storage := NewMemStorage()
	observer := &mockObserver{}

	// Attach observer
	storage.Attach(observer)
	assert.Len(t, storage.metricsObservers, 1)

	// Detach observer
	storage.Detach(observer)
	assert.Len(t, storage.metricsObservers, 0)
}

func TestMemStorage_Notify(t *testing.T) {
	storage := NewMemStorage()
	observer := &mockObserver{}
	storage.Attach(observer)

	// Create a metric
	storage.Create(context.Background(), "metric1", "gauge")

	// Update the metric to trigger notification
	err := storage.UpdateGauge(context.Background(), "metric1", 10.5)
	require.NoError(t, err)

	assert.True(t, observer.notifyCalled)
	assert.Equal(t, "metric1", observer.metric.Name)
	assert.Equal(t, "gauge", observer.metric.MType)
}
