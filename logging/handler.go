package logging

import (
	"context"
	"log/slog"
	"strings"
)

var logLevelFunc func(key string) slog.Level

type ComponentHandler struct {
	name string
	slog.Handler
}

func ComponentLoggerFor(name string) slog.Handler {
	return &ComponentHandler{name: strings.Join([]string{name, "logging"}, "."), Handler: handler}
}

func (h *ComponentHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if logLevelFunc != nil {
		return logLevelFunc(h.name) <= level
	}

	return slog.Default().Enabled(ctx, level)
}

type BaseHandler struct {
	slog.Handler
}

func (h *BaseHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Attrs(emitMetric(r.Level))
	r.AddAttrs(slogCtx(ctx)...)
	return h.Handler.Handle(ctx, r)
}
