package http

import (
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"net/http"
	"time"
)

func NewLoggingMiddleware(logger slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func(begin time.Time) {
				fields := newHTTPLogFields(ww, r)
				logger.WithFields(fields).WithDur(time.Since(begin)).InfoContext(r.Context(), "Handler complete")
			}(time.Now())

			next.ServeHTTP(w, r)

		}
		return http.HandlerFunc(fn)
	}
}

func newHTTPLogFields(ww middleware.WrapResponseWriter, r *http.Request) map[string]any {
	m := map[string]any{
		"status":      ww.Status(),
		"bytes":       ww.BytesWritten(),
		"method":      r.Method,
		"path":        r.URL.Path,
		"userAgent":   r.UserAgent(),
		"httpVersion": r.Proto,
		"ip":          ClientIP(r),
	}

	if len(r.URL.RawQuery) > 0 {
		m["query"] = r.URL.RawQuery
	}

	return m
}
