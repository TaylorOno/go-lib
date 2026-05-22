package logging

import (
	"context"
	"log/slog"
)

type slogCtxKey struct{}

func slogCtx(ctx context.Context) []slog.Attr {
	if attr, ok := ctx.Value(slogCtxKey{}).([]slog.Attr); ok {
		return attr
	}

	return []slog.Attr{}
}

// WithLogContext adds attributes to the log context that will be included if using framework logger.
func WithLogContext(ctx context.Context, a ...slog.Attr) context.Context {
	attrs := slogCtx(ctx)
	attrs = append(attrs, a...)
	return context.WithValue(ctx, slogCtxKey{}, a)
}
