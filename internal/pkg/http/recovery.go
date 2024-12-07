package http

import (
	"errors"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"net/http"
)

// NewRecoveryMiddleware returns [http.Handler] middleware that converts a panic
// into an error that is logged and handled by the given [EncodeErrorFunc].
func NewRecoveryMiddleware(logger slog.Logger, enc EncodeErrorFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			defer func() {
				p := recover()
				if p != nil {
					switch t := p.(type) {
					case string:
						err = errors.New(t)
					case error:
						err = t
					default:
						err = errors.New("unknown error")
					}

					logger.WithError(err).ErrorContextf(r.Context(), "Problem handling request: panic: %T{%+v}", p, p)
					enc(r.Context(), err, w)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
