package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/shadyziedan/metrica/internal/server/logger"
)

type (
	responseData struct {
		wroteHeader bool
		status      int
		size        int
	}

	loggerResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func newLoggerResponseWriter(w http.ResponseWriter) *loggerResponseWriter {
	return &loggerResponseWriter{ResponseWriter: w, responseData: new(responseData)}
}

func (w *loggerResponseWriter) Write(buf []byte) (int, error) {
	w.maybeWriteHeader()
	w.responseData.size = len(buf)
	return w.ResponseWriter.Write(buf)
}

func (w *loggerResponseWriter) maybeWriteHeader() {
	if !w.responseData.wroteHeader {
		w.responseData.status = 200
	}
}

func (w *loggerResponseWriter) WriteHeader(statusCode int) {
	if !w.responseData.wroteHeader {
		w.responseData.wroteHeader = true
		w.responseData.status = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func RequestLogger(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {

		loggerRw := newLoggerResponseWriter(w)
		timeStart := time.Now()
		defer func() {
			logger.Log.Info(
				"incoming request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", loggerRw.responseData.status),
				zap.String("duration", time.Since(timeStart).String()),
				zap.Int("response size", loggerRw.responseData.size),
			)
		}()

		next.ServeHTTP(loggerRw, r)

	}
	return http.HandlerFunc(logFn)
}
