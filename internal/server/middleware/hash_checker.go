package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/server/logger"
)

func HashChecker(key string) func(http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hashString := r.Header.Get("HashSHA256")
			if hashString == "" {
				nextHandler.ServeHTTP(w, r)
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
				logger.Log.Info("incorrect hash", zap.String("signature", signature))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if err != nil {
				logger.Log.Error("error closing body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			nextHandler.ServeHTTP(w, r)
		})
	}
}
