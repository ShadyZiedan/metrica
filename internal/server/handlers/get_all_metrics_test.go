package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shadyziedan/metrica/internal/models"
)

func TestGetAll_Success(t *testing.T) {
	// Arrange
	metrics := []*models.Metric{
		models.NewCounterMetric("metric1", 1),
		models.NewGaugeMetric("metric2", 5.5),
		models.NewCounterMetric("metric3", 10),
	}
	repo := &MockRepository{}
	repo.On("FindAll", mock.Anything).Return(metrics, nil)
	handler := &MetricHandler{repository: repo}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	// Act
	handler.GetAll(rw, req)

	// Assert
	repo.AssertExpectations(t)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "text/html", rw.Header().Get("Content-Type"))
	assert.Contains(t, rw.Body.String(), "metric1")
	assert.Contains(t, rw.Body.String(), "10")
	assert.Contains(t, rw.Body.String(), "5.5")
}

func TestGetAll_Error(t *testing.T) {
	// Arrange
	repo := &MockRepository{}
	repo.On("FindAll", mock.Anything).Return([]*models.Metric{}, errors.New("failed to fetch metrics"))
	handler := &MetricHandler{repository: repo}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	// Act
	handler.GetAll(rw, req)

	// Assert
	repo.AssertExpectations(t)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	assert.Equal(t, "text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
	assert.Equal(t, "failed to fetch metrics\n", rw.Body.String())
}

func TestGetAll_NoMetrics(t *testing.T) {
	// Arrange
	repo := &MockRepository{}
	repo.On("FindAll", mock.Anything).Return([]*models.Metric{}, nil)
	handler := &MetricHandler{repository: repo}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	// Act
	handler.GetAll(rw, req)

	// Assert
	repo.AssertExpectations(t)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "text/html", rw.Header().Get("Content-Type"))
	assert.NotContains(t, rw.Body.String(), "metric1")
	assert.NotContains(t, rw.Body.String(), "10")
	assert.NotContains(t, rw.Body.String(), "5.5")
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Find(ctx context.Context, name string) (*models.Metric, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*models.Metric), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, name string, mType string) error {
	args := m.Called(ctx, name, mType)
	return args.Error(0)
}

func (m *MockRepository) FindOrCreate(ctx context.Context, name string, mType string) (*models.Metric, error) {
	args := m.Called(ctx, name, mType)
	return args.Get(0).(*models.Metric), args.Error(1)
}

func (m *MockRepository) UpdateCounter(ctx context.Context, name string, delta int64) error {
	args := m.Called(ctx, name, delta)
	return args.Error(0)
}

func (m *MockRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *MockRepository) FindAllByName(ctx context.Context, names []string) ([]*models.Metric, error) {
	args := m.Called(ctx, names)
	return args.Get(0).([]*models.Metric), args.Error(1)
}

func (m *MockRepository) FindAll(ctx context.Context) ([]*models.Metric, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.Metric), args.Error(1)
}
