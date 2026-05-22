package profiler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
)

func ListenAndServe(ctx context.Context, port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))

	profileServer := &http.Server{
		Addr:    net.JoinHostPort("", port),
		Handler: mux,
	}

	go func() {
		slog.Info(fmt.Sprintf("debug on port %s", profileServer.Addr))
		if err := profileServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error serving debug", slog.String("", err.Error()))
			ctx.Done()
		}
	}()
}
