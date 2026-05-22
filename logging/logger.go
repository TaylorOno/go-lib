package logging

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"
	"testing"
)

var (
	lvl          slog.Level
	enableJSON   bool
	enableSource bool
	handler      slog.Handler
)

func init() {
	flag.TextVar(&lvl, "log-level", slog.LevelInfo, "log level: [debug info warn error]")
	flag.BoolVar(&enableJSON, "log-json", false, "enable structured logging")
	flag.BoolVar(&enableSource, "log-source", false, "enable logging of source file and line")
}

func Level() slog.Level {
	return lvl
}

// InitLogger initializes the base logger configured via program flags.
func InitLogger(_ context.Context) {
	if testing.Testing() {
		handler = slog.Default().Handler()
		return
	}

	flag.Parse()
	opts := &slog.HandlerOptions{Level: lvl, AddSource: enableSource}
	writer := &WriterReporter{os.Stdout}
	if !enableJSON {
		handler = &BaseHandler{slog.NewTextHandler(writer, opts)}
	} else {
		handler = &BaseHandler{slog.NewJSONHandler(writer, opts)}
	}

	slog.SetDefault(slog.New(handler))
	return
}

// WithEnabledFunction allows you to set a function for dynamically determining log levels.
// A typical use case is to use config.GetLogLevel, which can update the log level at runtime.
func WithEnabledFunction(lvlFunc func(key string) slog.Level) func(context.Context) {
	return func(ctx context.Context) {
		logLevelFunc = lvlFunc
	}
}

func GetHandler() slog.Handler {
	if handler == nil {
		InitLogger(context.Background())
	}

	return handler
}

type WriterReporter struct {
	io.Writer
}

func (w *WriterReporter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	go bytesLoggedFunc(n)
	return n, err
}
