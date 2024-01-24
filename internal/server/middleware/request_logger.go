package middleware

import (
	"github.com/shadyziedan/metrica/internal/server/logger"
	"go.uber.org/zap"
	"net/http"
	"time"
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
	w.responseData.wroteHeader = true
	w.responseData.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogger(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {

		loggerRw := newLoggerResponseWriter(w)
		duration := time.Duration(0)
		defer func() {
			logger.Log.Info(
				"incoming request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", loggerRw.responseData.status),
				zap.String("duration", duration.String()),
				zap.Int("response size", loggerRw.responseData.size),
			)
		}()
		timeStart := time.Now()
		next.ServeHTTP(loggerRw, r)
		duration = time.Since(timeStart)

	}
	return http.HandlerFunc(logFn)
}
