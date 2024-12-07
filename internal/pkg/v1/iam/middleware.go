package iam

import (
	"github.com/zhughes3/go-accelerate/internal/pkg/v1/user"
	"github.com/zhughes3/go-accelerate/pkg/slog"
	"net/http"
)

func NewAuthMiddleware(logger slog.Logger) func(next http.Handler) http.Handler {
	return authMiddleware{logger: logger}.handler
}

type authMiddleware struct {
	logger slog.Logger

	// need a user service to fetch user ID from token potentially
}

func (a authMiddleware) handler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var ctx = r.Context()
		// TODO get user id from token
		ctx = user.WithUserID(ctx, "fake-user-id")
		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(fn)
}
