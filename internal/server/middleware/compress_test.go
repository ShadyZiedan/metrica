package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to gzip data
func gzipData(t *testing.T, data string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(data)); err != nil {
		t.Fatal(err.Error())
	}
	gz.Close()
	return buf.Bytes()
}

func TestCompress_ResponseWithGzip(t *testing.T) {
	// Create a simple handler that writes "hello, world" as the response
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, world"))
	})

	// Wrap it with the Compress middleware
	handler := Compress(next)

	// Create a request with Accept-Encoding: gzip
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Record the response
	rec := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rec, req)

	// Assert that Content-Encoding is gzip
	assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

	// Check if the response is gzip-compressed
	respBody := rec.Body.Bytes()
	reader, err := gzip.NewReader(bytes.NewReader(respBody))
	assert.NoError(t, err)
	defer reader.Close()

	// Read the decompressed response
	body, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.Equal(t, "hello, world", string(body))
}

func TestCompress_RequestWithGzip(t *testing.T) {
	// Gzip the request body
	gzippedBody := gzipData(t, "hello, server")

	// Create a simple handler that reads the request body
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		// Assert that the request body is correctly decompressed
		assert.Equal(t, "hello, server", string(body))
		w.Write([]byte("ok"))
	})

	// Wrap it with the Compress middleware
	handler := Compress(next)

	// Create a request with Content-Encoding: gzip
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(gzippedBody))
	req.Header.Set("Content-Encoding", "gzip")

	// Record the response
	rec := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rec, req)

	// Assert that the response status is 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCompress_NoGzip(t *testing.T) {
	// Create a simple handler that writes a plain response
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("no gzip here"))
	})

	// Wrap it with the Compress middleware
	handler := Compress(next)

	// Create a request without gzip encoding
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Record the response
	rec := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(rec, req)

	// Assert that Content-Encoding is not set
	assert.Empty(t, rec.Header().Get("Content-Encoding"))

	// Assert that the response body is not compressed
	assert.Equal(t, "no gzip here", rec.Body.String())
}
