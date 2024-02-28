package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/security"
	"github.com/shadyziedan/metrica/internal/server/logger"
)

func HashChecker(key string) func(http.Handler) http.Handler {
	if key == "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responseWriter := newHashResponseWriter(w, key)
			hashString := r.Header.Get("HashSHA256")
			if hashString == "" {
				nextHandler.ServeHTTP(responseWriter, r)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Error("error reading body", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			copiedBody := bytes.NewBuffer(body)
			r.Body = io.NopCloser(copiedBody)
			defer r.Body.Close()

			h := hmac.New(sha256.New, []byte(key))
			_, err = h.Write(body)
			if err != nil {
				logger.Log.Error("Error hashing body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			signature := hex.EncodeToString(h.Sum(nil))
			if signature != hashString {
				logger.Log.Info("Invalid signature", zap.String("signature", signature))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if err != nil {
				logger.Log.Error("error closing body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			nextHandler.ServeHTTP(responseWriter, r)
		})
	}
}

type hashResponseWriter struct {
	http.ResponseWriter
	secretKey string
}

func newHashResponseWriter(w http.ResponseWriter, secretKey string) *hashResponseWriter {
	return &hashResponseWriter{ResponseWriter: w, secretKey: secretKey}
}
func (w *hashResponseWriter) Write(buf []byte) (int, error) {
	hashString, err := security.Hash(buf, w.secretKey)
	if err != nil {
		return 0, err
	}
	w.ResponseWriter.Header().Add("HashSHA256", hashString)
	return w.ResponseWriter.Write(buf)
}
