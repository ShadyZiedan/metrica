package services

import (
	"bytes"
	"context"
	"google.golang.org/grpc"
	"io"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shadyziedan/metrica/internal/models"
	pb "github.com/shadyziedan/metrica/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) Hash(data []byte) (string, error) {
	args := m.Called(data)
	return args.String(0), args.Error(1)
}

type MockEncryptor struct {
	mock.Mock
}

func (m *MockEncryptor) Encrypt(data []byte) ([]byte, error) {
	args := m.Called(data)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryptor) GetEncryptedKey() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type MockGRPCClient struct {
	mock.Mock
	pb.MetricsServiceClient
}

func (m *MockGRPCClient) UpdateBatchMetrics(
	ctx context.Context,
	req *pb.UpdateBatchRequest,
	opts ...grpc.CallOption,
) (*pb.UpdateBatchResponse, error) {
	args := m.Called(ctx, req, opts)
	return args.Get(0).(*pb.UpdateBatchResponse), args.Error(1)
}

func TestMetricsSenderInitialization(t *testing.T) {
	mockHasher := &MockHasher{}
	mockEncryptor := &MockEncryptor{}
	httpClient := resty.New()
	grpcClient := &MockGRPCClient{}

	sender := NewMetricsSender(httpClient, grpcClient, WithHasher(mockHasher), WithEncryptor(mockEncryptor))
	assert.NotNil(t, sender)
	assert.Equal(t, httpClient, sender.httpClient)
	assert.Equal(t, grpcClient, sender.grpcClient)
	assert.Equal(t, mockHasher, sender.hasher)
	assert.Equal(t, mockEncryptor, sender.encryptor)
}

func TestSendGRPC(t *testing.T) {
	mockGRPCClient := &MockGRPCClient{}
	metrics := []*models.Metrics{
		{ID: "test1", MType: "counter", Delta: func(v int64) *int64 { return &v }(10)},
	}

	mockGRPCClient.On(
		"UpdateBatchMetrics",
		mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil // Validate context
		}),
		mock.MatchedBy(func(req *pb.UpdateBatchRequest) bool {
			return req != nil && len(req.Metrics) > 0 // Validate request
		}),
		mock.Anything, // Skip validation for grpc.CallOption in this case
	).Return(&pb.UpdateBatchResponse{}, nil)

	sender := NewMetricsSender(nil, mockGRPCClient)
	err := sender.Send(context.Background(), metrics)

	assert.NoError(t, err)
	mockGRPCClient.AssertCalled(
		t,
		"UpdateBatchMetrics",
		mock.Anything,                            // Context
		mock.Anything,                            // Request
		mock.AnythingOfType("[]grpc.CallOption"), // Variadic grpc.CallOption as a slice
	)
}

func TestSendJSON(t *testing.T) {
	mockHasher := &MockHasher{}
	mockHasher.On("Hash", mock.Anything).Return("mockHash", nil)

	mockEncryptor := &MockEncryptor{}
	mockEncryptor.On("GetEncryptedKey").Return("mockKey", nil)
	mockEncryptor.On("Encrypt", mock.Anything).Return([]byte("encryptedData"), nil)

	httpClient := resty.New()
	httpClient.SetTransport(&mockTransport{
		status: 200,
		body:   "Success",
	})

	sender := NewMetricsSender(httpClient, nil, WithHasher(mockHasher), WithEncryptor(mockEncryptor))

	metrics := []*models.Metrics{
		{ID: "test1", MType: "gauge", Value: func(v float64) *float64 { return &v }(42.42)},
	}

	err := sender.Send(context.Background(), metrics)

	assert.NoError(t, err)
	mockHasher.AssertCalled(t, "Hash", mock.Anything)
	mockEncryptor.AssertCalled(t, "GetEncryptedKey")
	mockEncryptor.AssertCalled(t, "Encrypt", mock.Anything)
}

// Mock Transport for HTTP requests
type mockTransport struct {
	status int
	body   string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(bytes.NewBufferString(m.body)),
	}
	return resp, nil
}
