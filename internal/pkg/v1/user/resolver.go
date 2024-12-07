package user

import (
	"context"
	acstrings "github.com/zhughes3/go-accelerate/pkg/strings"
)

type userIDKey struct{}

// An IDResolver returns the user identifier from the given context.
type IDResolver func(context.Context) string

func MustResolveID(ctx context.Context) string {
	userID, ok := ResolveID(ctx)
	if !ok {
		panic("no userID on context")
	}

	return userID
}

func ResolveID(ctx context.Context) (string, bool) {
	if ctx.Value(userIDKey{}) == nil {
		return "", false
	}

	userID := ctx.Value(userIDKey{}).(string)

	return userID, acstrings.IsNotBlank(userID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}
