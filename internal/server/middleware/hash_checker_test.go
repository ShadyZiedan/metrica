package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockHasher is a mock implementation of the hasher interface
type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) Hash(body []byte) (string, error) {
	args := m.Called(body)
	return args.String(0), args.Error(1)
}

func TestHashChecker(t *testing.T) {
	// Setup
	mockHasher := new(MockHasher)
	responseBody := []byte("OK")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(responseBody) // This is the response body that will be hashed
	})

	t.Run("valid hash", func(t *testing.T) {
		// Arrange
		requestBody := []byte("test-body")
		expectedHash := "validhash"

		// Mock the hasher to return the expected hash for both request and response bodies
		mockHasher.On("Hash", requestBody).Return(expectedHash, nil).Once()
		mockHasher.On("Hash", responseBody).Return(expectedHash, nil).Once()

		middleware := HashChecker(mockHasher)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
		req.Header.Set("HashSHA256", expectedHash)
		rec := httptest.NewRecorder()

		// Act
		middleware(handler).ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expectedHash, rec.Header().Get("HashSHA256")) // Ensure response header has the hash
		mockHasher.AssertExpectations(t)
	})

	t.Run("invalid hash", func(t *testing.T) {
		// Arrange
		requestBody := []byte("test-body")
		expectedHash := "validhash"
		invalidHash := "invalidhash"

		// Mock the hasher to return the expected hash
		mockHasher.On("Hash", requestBody).Return(expectedHash, nil).Once()

		middleware := HashChecker(mockHasher)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
		req.Header.Set("HashSHA256", invalidHash) // Incorrect hash
		rec := httptest.NewRecorder()

		// Act
		middleware(handler).ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		mockHasher.AssertExpectations(t)
	})

	t.Run("missing hash header", func(t *testing.T) {
		// Arrange
		requestBody := []byte("test-body")
		middleware := HashChecker(mockHasher)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(requestBody))
		rec := httptest.NewRecorder()

		mockHasher.On("Hash", responseBody).Return("validhash", nil)

		// Act
		middleware(handler).ServeHTTP(rec, req)

		// Assert
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
