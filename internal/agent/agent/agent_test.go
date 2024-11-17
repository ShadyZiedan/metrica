package agent

import (
	"context"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shadyziedan/metrica/internal/agent/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shadyziedan/metrica/internal/agent/services"
	"github.com/shadyziedan/metrica/internal/models"
	"github.com/shadyziedan/metrica/internal/server/middleware"
)

// MockMetricsCollector is a mock implementation of the metricsCollector interface
type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) Collect() *services.AgentMetrics {
	args := m.Called()
	return args.Get(0).(*services.AgentMetrics)
}

func (m *MockMetricsCollector) IncreasePollCount() {
	m.Called()
}

// TestNewAgent tests the NewAgent function
func TestNewAgent(t *testing.T) {
	mc := new(MockMetricsCollector)
	cnf := config.Config{
		Address:        "http://example.com",
		PollInterval:   config.Duration{Duration: time.Second * 5},
		ReportInterval: config.Duration{Duration: time.Second * 10},
		RateLimit:      2,
	}

	a := NewAgent(cnf, mc, nil)

	assert.NotNil(t, a)
	assert.Equal(t, 5*time.Second, a.PollInterval)
	assert.Equal(t, 10*time.Second, a.ReportInterval)
	assert.Equal(t, 2, a.RateLimit)
}

// TestSendMetricsToServer tests the sendMetricsToServer method
func TestSendMetricsToServer(t *testing.T) {
	// Set up a mock server to respond to the request
	server := httptest.NewServer(middleware.Compress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var receivedMetrics []*models.Metrics
		err := json.NewDecoder(r.Body).Decode(&receivedMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(receivedMetrics))
		w.WriteHeader(http.StatusOK)
	})))
	defer server.Close()

	mc := new(MockMetricsCollector)
	cnf := config.Config{
		Address:        server.URL,
		ReportInterval: config.Duration{Duration: time.Second * 5},
		PollInterval:   config.Duration{Duration: time.Second * 10},
		RateLimit:      2,
	}
	ms := services.NewMetricsSender(resty.New().SetBaseURL(server.URL), nil)
	a := NewAgent(cnf, mc, ms)

	metrics := services.NewAgentMetrics()

	metrics.Gauge.UpdateMetric("test_gauge", 123.45)
	metrics.Counter.UpdateMetric("test_counter", 1)

	err := a.sendMetricsToServer(context.Background(), metrics)
	assert.NoError(t, err)
}

// TestRun tests the Run method of the Agent
func TestRun(t *testing.T) {
	mc := new(MockMetricsCollector)
	cnf := config.Config{
		Address:        "http://example.com",
		ReportInterval: config.Duration{Duration: time.Second * 2},
		PollInterval:   config.Duration{Duration: time.Second * 1},
		RateLimit:      2,
	}

	a := NewAgent(cnf, mc, services.NewMetricsSender(resty.New().SetBaseURL(cnf.Address), nil))

	// Set up a mock for metrics collection
	mc.On("IncreasePollCount").Return()
	mc.On("Collect").Return(services.NewAgentMetrics())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.Run(ctx)

	time.Sleep(3 * time.Second) // Wait for a few ticks

	cancel() // Stop the agent

	// Check if the metrics collector methods were called
	mc.AssertExpectations(t)
}

// TestSendMetricsFailure tests the error handling in sendMetricsToServer
func TestSendMetricsFailure(t *testing.T) {
	// Set up a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Internal Server Error")
	}))
	defer server.Close()

	mc := new(MockMetricsCollector)
	cnf := config.Config{
		Address:        server.URL,
		ReportInterval: config.Duration{Duration: time.Second * 5},
		PollInterval:   config.Duration{Duration: time.Second * 10},
		RateLimit:      2,
	}
	a := NewAgent(cnf, mc, services.NewMetricsSender(resty.New().SetBaseURL(server.URL), nil))

	metrics := services.NewAgentMetrics()
	metrics.Gauge.UpdateMetric("test_gauge", 123.45)
	metrics.Counter.UpdateMetric("test_counter", 1)

	err := a.sendMetricsToServer(context.Background(), metrics)
	assert.Error(t, err)
	assert.Equal(t, "request failed with status 500: Internal Server Error", err.Error())
}

func TestMetricsRealIP(t *testing.T) {
	// Set up a mock server to respond to the request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Header.Get("X-Real-IP"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mc := new(MockMetricsCollector)
	cnf := config.Config{
		Address:        server.URL,
		ReportInterval: config.Duration{Duration: time.Second * 5},
		PollInterval:   config.Duration{Duration: time.Second * 10},
		RateLimit:      2,
	}
	a := NewAgent(cnf, mc, services.NewMetricsSender(resty.New().SetBaseURL(server.URL), nil))

	metrics := services.NewAgentMetrics()
	metrics.Gauge.UpdateMetric("test_gauge", 123.45)
	metrics.Counter.UpdateMetric("test_counter", 1)
	err := a.sendMetricsToServer(context.Background(), metrics)
	assert.NoError(t, err)
}
