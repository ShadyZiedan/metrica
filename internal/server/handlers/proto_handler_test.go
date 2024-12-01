package handlers

import (
	"context"
	"github.com/shadyziedan/metrica/internal/models"
	"google.golang.org/protobuf/proto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/shadyziedan/metrica/proto"
)

type MockMetricsRepository struct {
	mock.Mock
}

// Create mocks the Create method
func (r *MockMetricsRepository) Create(ctx context.Context, name string, mType string) error {
	args := r.Called(ctx, name, mType)
	return args.Error(0)
}

// FindAll mocks the FindAll method
func (r *MockMetricsRepository) FindAll(ctx context.Context) ([]*models.Metric, error) {
	args := r.Called(ctx)
	return args.Get(0).([]*models.Metric), args.Error(1)
}

// FindAllByName mocks the FindAllByName method
func (r *MockMetricsRepository) FindAllByName(ctx context.Context, names []string) ([]*models.Metric, error) {
	args := r.Called(ctx, names)
	return args.Get(0).([]*models.Metric), args.Error(1)
}

func (r *MockMetricsRepository) FindOrCreate(ctx context.Context, id string, mType string) (*models.Metric, error) {
	args := r.Called(ctx, id, mType)
	return args.Get(0).(*models.Metric), args.Error(1)
}

func (r *MockMetricsRepository) UpdateCounter(ctx context.Context, name string, delta int64) error {
	return r.Called(ctx, name, delta).Error(0)
}

func (r *MockMetricsRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	return r.Called(ctx, name, value).Error(0)
}

func (r *MockMetricsRepository) Find(ctx context.Context, name string) (*models.Metric, error) {
	args := r.Called(ctx, name)
	return args.Get(0).(*models.Metric), args.Error(1)
}

func TestUpdateBatchMetrics_Success(t *testing.T) {
	mockRepo := &MockMetricsRepository{}
	server := NewGRPCMetricServiceServer(mockRepo)

	ctx := context.Background()

	// Input data
	req := &pb.UpdateBatchRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "test_counter",
				MType: pb.Metric_COUNTER,
				Delta: proto.Int64(5),
			},
			{
				Id:    "test_gauge",
				MType: pb.Metric_GAUGE,
				Value: proto.Float64(42.42),
			},
		},
	}

	// Mock repository behavior
	mockRepo.On("FindOrCreate", ctx, "test_counter", "counter").Return(&models.Metric{Name: "test_counter", MType: "counter"}, nil)
	mockRepo.On("UpdateCounter", ctx, "test_counter", int64(5)).Return(nil)
	mockRepo.On("Find", ctx, "test_counter").Return(&models.Metric{Name: "test_counter", MType: "counter", Counter: proto.Int64(10)}, nil)

	mockRepo.On("FindOrCreate", ctx, "test_gauge", "gauge").Return(&models.Metric{Name: "test_gauge", MType: "gauge"}, nil)
	mockRepo.On("UpdateGauge", ctx, "test_gauge", float64(42.42)).Return(nil)
	mockRepo.On("Find", ctx, "test_gauge").Return(&models.Metric{Name: "test_gauge", MType: "gauge", Gauge: proto.Float64(42.42)}, nil)

	// Call method
	resp, err := server.UpdateBatchMetrics(ctx, req)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, resp.Metrics, 2)

	assert.Equal(t, "test_counter", resp.Metrics[0].Id)
	assert.Equal(t, pb.Metric_COUNTER, resp.Metrics[0].MType)
	assert.Equal(t, int64(10), *resp.Metrics[0].Delta)

	assert.Equal(t, "test_gauge", resp.Metrics[1].Id)
	assert.Equal(t, pb.Metric_GAUGE, resp.Metrics[1].MType)
	assert.Equal(t, float64(42.42), *resp.Metrics[1].Value)

	// Verify mock calls
	mockRepo.AssertExpectations(t)
}

func TestUpdateBatchMetrics_InvalidMetricType(t *testing.T) {
	mockRepo := &MockMetricsRepository{}
	server := NewGRPCMetricServiceServer(mockRepo)

	ctx := context.Background()

	req := &pb.UpdateBatchRequest{
		Metrics: []*pb.Metric{
			{Id: "invalid_metric", MType: pb.Metric_MType(-1)},
		},
	}

	resp, err := server.UpdateBatchMetrics(ctx, req)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}
