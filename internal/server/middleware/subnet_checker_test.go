package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubnetCheckerMiddleware(t *testing.T) {
	trustedSubnet := "192.168.0.0/24"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Expected to be forbidden")
	})
	middleware, err := NewTrustedSubNetMiddleware(trustedSubnet)
	if err != nil {
		t.Fatalf("Failed to create middleware: %v", err)
	}
	handler := middleware(nextHandler)

	request := httptest.NewRequest("GET", "http://example.com", nil)
	request.Header.Set("X-Real-IP", "192.168.1.1")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	assert.Equal(t, http.StatusForbidden, recorder.Code)
}
