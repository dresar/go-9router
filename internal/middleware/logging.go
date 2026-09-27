package middleware

import (
	"net/http"
	"time"

	"github.com/dresar/go-9router/internal/logging"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)
		elapsed := time.Since(start)
		logging.Info("HTTP", r.Method+" "+r.URL.Path,
			"status", rw.code,
			"ms", elapsed.Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	code    int
	written bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.code = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}
