package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/shadyziedan/metrica/internal/server/logger"
)

// Initialize logger for testing
func setupLogger() (*zap.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{}), // Use console encoder for better readability
		zapcore.AddSync(&buf),                              // Write logs to the buffer
		zapcore.DebugLevel,                                 // Log level
	))
	return l, &buf
}

type loggerOutput struct {
	Method       string `json:"method"`
	URI          string `json:"uri"`
	Status       int    `json:"status"`
	Duration     string `json:"duration"`
	ResponseSize int    `json:"response size"`
}

func TestRequestLogger(t *testing.T) {
	// Setup logger and buffer
	l, buf := setupLogger()
	defer l.Sync() // Flush any buffered log entries
	logger.Log = l

	// Define a simple handler to test the middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK")) // Response body
	})

	// Wrap the handler with the RequestLogger middleware
	middleware := RequestLogger(handler)

	// Create a test HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(rec, req)

	// Verify the response status code and body
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	// Check the logged output
	var logOutput loggerOutput
	assert.Nil(t, json.NewDecoder(buf).Decode(&logOutput), "logger output couldn't be parsed")

	assert.Equal(t, "GET", logOutput.Method)
	assert.Equal(t, "/", logOutput.URI)
	assert.Equal(t, 200, logOutput.Status)
	assert.NotEmpty(t, logOutput.Duration)
	assert.Equal(t, 2, logOutput.ResponseSize)
}
