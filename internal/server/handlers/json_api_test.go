package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shadyziedan/metrica/internal/server/storage"
)

func TestJSONApi(t *testing.T) {
	tests := []struct {
		title       string
		request     string
		requestBody string
		method      string
		metricName  string
		want        struct {
			statusCode   int
			responseBody string
			err          bool
		}
	}{
		{
			title:      "Not allowed method",
			request:    "/update/",
			method:     "GET",
			metricName: "Alloc",
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode: http.StatusMethodNotAllowed,
				err:        true,
			},
		},
		{
			title:       "Adding new gauge metric",
			request:     "/update/",
			method:      "POST",
			requestBody: `{ "id": "Alloc", "type": "gauge", "value": 55.05 }`,
			metricName:  "Alloc",
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode:   http.StatusOK,
				responseBody: `{ "id": "Alloc", "type": "gauge", "value": 55.05 }`,
			},
		},
		{
			title:       "Adding new counter metric",
			request:     "/update/",
			method:      "POST",
			metricName:  "PollCount",
			requestBody: `{ "id": "PollCount", "type": "counter", "delta": 100 }`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode:   http.StatusOK,
				responseBody: `{ "id": "PollCount", "type": "counter", "delta": 100 }`,
			},
		},
		{
			title:       "Adding unknown metric type",
			request:     "/update/",
			method:      "POST",
			metricName:  "SYS",
			requestBody: `{ "id": "PollCount", "type": "unknown", "delta": 100 }`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode: http.StatusBadRequest,
				err:        true,
			},
		},
		{
			title:       "updating counter metric",
			request:     "/update/",
			method:      "POST",
			metricName:  "PollCount",
			requestBody: `{ "id": "PollCount", "type": "counter", "delta": 50 }`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode:   http.StatusOK,
				responseBody: `{ "id": "PollCount", "type": "counter", "delta": 150 }`,
			},
		},
		{
			title:       "updating gauge metric",
			request:     "/update/",
			method:      "POST",
			metricName:  "Alloc",
			requestBody: `{ "id": "Alloc", "type": "gauge", "value": 55.065 }`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode:   http.StatusOK,
				responseBody: `{ "id": "Alloc", "type": "gauge", "value": 55.065 }`,
			},
		},
		{
			title:       "getting gauge metric value",
			request:     "/value/",
			method:      "POST",
			metricName:  "Alloc",
			requestBody: `{ "id": "Alloc", "type": "gauge"}`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode:   http.StatusOK,
				responseBody: `{ "id": "Alloc", "type": "gauge", "value": 55.065 }`,
			},
		},
		{
			title:       "getting counter metric value",
			request:     "/value/",
			method:      "POST",
			metricName:  "Alloc",
			requestBody: `{ "id": "PollCount", "type": "counter"}`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode:   http.StatusOK,
				responseBody: `{ "id": "PollCount", "type": "counter", "delta": 150 }`,
			},
		},
		{
			title:       "getting unknown metric value",
			request:     "/value/",
			method:      "POST",
			metricName:  "Alloc",
			requestBody: `{ "id": "SYS", "type": "counter"}`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode: http.StatusNotFound,
				err:        true,
			},
		},
		{
			title:       "update batch metric value",
			request:     "/updates/",
			method:      "POST",
			requestBody: `[{ "id": "Alloc123", "type": "gauge", "value": 55.05 }, { "id": "PollCount123", "type": "counter", "delta": 100 } ]`,
			want: struct {
				statusCode   int
				responseBody string
				err          bool
			}{
				statusCode:   http.StatusOK,
				responseBody: `[{ "id": "Alloc123", "type": "gauge", "value": 55.05 }, { "id": "PollCount123", "type": "counter", "delta": 100 } ]`,
			},
		},
	}

	memStorage := storage.NewMemStorage()
	router := NewRouter(nil, memStorage)
	srv := httptest.NewServer(router)
	defer srv.Close()
	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			url := srv.URL + tt.request
			req, err := http.NewRequest(tt.method, url, strings.NewReader(tt.requestBody))
			require.NoError(t, err)
			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)

			if !tt.want.err {
				require.NoError(t, err)
				responseBody, err := io.ReadAll(res.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, tt.want.responseBody, string(responseBody))
			}
		})
	}
}
