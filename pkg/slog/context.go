package slog

import "context"

type contextExtractor func(context.Context) map[string]any
