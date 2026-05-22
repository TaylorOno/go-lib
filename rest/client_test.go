package rest

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/taylorono/go-lib/metrics"
)

func TestClientFunc_Do(t *testing.T) {
	level := slog.LevelDebug

	header := http.Header{}
	header.Add("x-test", "test")

	opts := &slog.HandlerOptions{Level: level}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))
	client := NewClientBuilder("test").
		WithMetricRegistry(metrics.NewPrometheusReporter()).
		WithMiddleware(
			Verbose().RequestLogger(),
		).
		Build()

	req, err := http.NewRequest("GET", "https://official-joke-api.appspot.com/random_joke", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	do, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, do.StatusCode)
}
