package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/shadyziedan/metrica/internal/models"
)

func TestGetMetric_Counter(t *testing.T) {
	repo := &MockRepository{}
	metricHandler := &MetricHandler{repository: repo}

	repo.On("Find", mock.Anything, "test").Return(models.NewCounterMetric("test", 10), nil)

	req := httptest.NewRequest(http.MethodGet, "/value/counter/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", "counter")
	rctx.URLParams.Add("metricName", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx)) // Attach route context

	rr := httptest.NewRecorder()
	metricHandler.GetMetric(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "10", rr.Body.String())
}

func TestGetMetric_Gauge(t *testing.T) {
	repo := &MockRepository{}
	metricHandler := &MetricHandler{
		repository: repo,
	}

	repo.On("Find", mock.Anything, "test").Return(models.NewGaugeMetric("test", 10.5), nil)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", "gauge")
	rctx.URLParams.Add("metricName", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx)) // Attach route context

	rr := httptest.NewRecorder()
	metricHandler.GetMetric(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "10.5", rr.Body.String())
}

func TestGetMetric_UnknownMetricType(t *testing.T) {
	repo := &MockRepository{}
	metricHandler := &MetricHandler{
		repository: repo,
	}

	repo.On("Find", mock.Anything, "test").Return(models.NewCounterMetric("test", 10), nil)

	req := httptest.NewRequest(http.MethodGet, "/value/unknown/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("metricType", "unknown")
	rctx.URLParams.Add("metricName", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx)) // Attach route context

	rr := httptest.NewRecorder()
	metricHandler.GetMetric(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "unknown metric type: unknown")
}
