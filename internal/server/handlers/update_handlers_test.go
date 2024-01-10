package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shadyziedan/metrica/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers(t *testing.T) {
	memStorage := storage.NewMemStorage()

	tests := []struct {
		title      string
		request    string
		method     string
		metricName string
		want       struct {
			statusCode int
			counter    int64
			gauge      float64
			err        bool
		}
	}{
		{
			title:      "Not allowed method",
			request:    "/update/gauge/Alloc/10.0",
			method:     "GET",
			metricName: "Alloc",
			want: struct {
				statusCode int
				counter    int64
				gauge      float64
				err        bool
			}{
				statusCode: http.StatusMethodNotAllowed,
				err:        true,
			},
		},
		{
			title:      "Adding new gauge metric",
			request:    "/update/gauge/Alloc/10.0",
			method:     "POST",
			metricName: "Alloc",
			want: struct {
				statusCode int
				counter    int64
				gauge      float64
				err        bool
			}{
				statusCode: http.StatusOK,
				counter:    0,
				gauge:      10.0,
			},
		},
		{
			title:      "Adding new counter metric",
			request:    "/update/counter/PollCount/20",
			method:     "POST",
			metricName: "PollCount",
			want: struct {
				statusCode int
				counter    int64
				gauge      float64
				err        bool
			}{
				statusCode: http.StatusOK,
				counter:    20,
				gauge:      0,
			},
		},
		{
			title:      "Adding unknown metric type",
			request:    "/update/unknown/SYS/20",
			method:     "POST",
			metricName: "SYS",
			want: struct {
				statusCode int
				counter    int64
				gauge      float64
				err        bool
			}{
				statusCode: http.StatusInternalServerError,
				err:        true,
			},
		},
		{
			title:      "updating counter metric",
			request:    "/update/counter/PollCount/20",
			method:     "POST",
			metricName: "PollCount",
			want: struct {
				statusCode int
				counter    int64
				gauge      float64
				err        bool
			}{
				statusCode: http.StatusOK,
				counter:    40,
				gauge:      0,
			},
		},
		{
			title:      "updating gauge metric",
			request:    "/update/gauge/Alloc/30",
			method:     "POST",
			metricName: "Alloc",
			want: struct {
				statusCode int
				counter    int64
				gauge      float64
				err        bool
			}{
				statusCode: http.StatusOK,
				counter:    0,
				gauge:      30.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			h := UpdateMetricHandler(memStorage)
			req := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			defer w.Result().Body.Close()
			assert.Equal(t, tt.want.statusCode, w.Result().StatusCode)
			if !tt.want.err {
				metric, err := memStorage.Find(tt.metricName)
				require.NoError(t, err)
				assert.Equal(t, tt.want.counter, metric.GetCounter())
				assert.Equal(t, tt.want.gauge, metric.GetGauge())
			}
		})
	}
}
