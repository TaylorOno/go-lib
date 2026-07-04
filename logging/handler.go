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

// Enabled determines if a log entry with the given level should be logged based on the component's log level settings.
// When a logLevelFunc has been configured via WithEnabledFunction, ComponentHandler log levels can be individually controlled with the key `{{name}}.logging`.
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
