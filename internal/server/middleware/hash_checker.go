package middleware

import (
	"bytes"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/server/logger"
)

type hasher interface {
	Hash(body []byte) (string, error)
}

// HashChecker is a middleware function that checks the HashSHA256 header of incoming HTTP requests.
// If the header doesn't match the SHA256 hash of the request body, it returns a 400 Bad Request status.
// If the hasher is nil, it returns the next handler without any modifications.
//
// The function takes a hasher interface as a parameter, which must implement the Hash method.
// The Hash method takes a byte slice and returns a string representing the SHA256 hash of the input.
//
// The function returns a new http.Handler that wraps the provided nextHandler.
// It reads the request body, calculates the SHA256 hash, compares it with the HashSHA256 header,
// and sets the HashSHA256 header in the response if the hasher is not nil.
func HashChecker(hasher hasher) func(http.Handler) http.Handler {
	if hasher == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responseWriter := newHashResponseWriter(w, hasher)
			hashString := r.Header.Get("HashSHA256")
			if hashString == "" {
				nextHandler.ServeHTTP(responseWriter, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("Error reading body", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))
			signature, err := hasher.Hash(body)
			if err != nil {
				logger.Log.Error("Error hashing body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			if signature != hashString {
				logger.Log.Info("Invalid signature", zap.String("signature", signature), zap.String("received hash", hashString))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			nextHandler.ServeHTTP(responseWriter, r)
		})
	}
}

type hashResponseWriter struct {
	http.ResponseWriter
	hasher hasher
}

func newHashResponseWriter(w http.ResponseWriter, hasher hasher) *hashResponseWriter {
	return &hashResponseWriter{ResponseWriter: w, hasher: hasher}
}
func (w *hashResponseWriter) Write(buf []byte) (int, error) {
	hashString, err := w.hasher.Hash(buf)
	if err != nil {
		return 0, err
	}
	w.ResponseWriter.Header().Add("HashSHA256", hashString)
	return w.ResponseWriter.Write(buf)
}
