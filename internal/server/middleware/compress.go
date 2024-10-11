// Package middleware contains HTTP middleware functions for handling various aspects of HTTP requests and responses.
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressResponseWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressResponseWriter {
	return &compressResponseWriter{
		w,
		gzip.NewWriter(w),
	}
}

func (c *compressResponseWriter) Write(b []byte) (int, error) {
	return c.zw.Write(b)
}

func (c *compressResponseWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gzipReader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r: r, zr: gzipReader}, nil
}

func (cr compressReader) Read(b []byte) (int, error) {
	return cr.zr.Read(b)
}

func (cr *compressReader) Close() error {
	// Close the wrapped gzip.Reader first
	if cr.zr != nil {
		if err := cr.zr.Close(); err != nil {
			return err
		}
	}

	// Close the underlying io.ReadCloser
	if cr.r != nil {
		return cr.r.Close()
	}

	return nil
}

// Compress reads compressed request data and returns compressed response data
func Compress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wo := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		acceptsGzip := strings.Contains(acceptEncoding, "gzip")

		if acceptsGzip {
			cw := newCompressWriter(w)
			defer cw.Close()
			wo = cw
			wo.Header().Set("Content-Encoding", "gzip")
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(wo, err.Error(), http.StatusInternalServerError)
				return
			}
			defer cr.Close()
			r.Body = cr
		}

		next.ServeHTTP(wo, r)
	})
}
