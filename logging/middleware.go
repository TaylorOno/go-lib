package logging

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
)

// HttpLoggingMiddleware creates a middleware that logs the full request and response
func HttpLoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		recorder := &responseRecorder{ResponseWriter: w, body: &bytes.Buffer{}}

		next(recorder, r)

		logResponse(recorder)
	}
}

func logRequest(r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		slog.Error("failed to dump request", slog.String("error", err.Error()))
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		slog.Debug("HTTP Request", slog.String("dump", string(requestDump)))
		return
	}

	// We need to parse the dump to separate headers and body
	parts := strings.SplitN(string(requestDump), "\r\n\r\n", 2)
	if len(parts) != 2 {
		slog.Debug("HTTP Request", slog.String("dump", string(requestDump)))
		return
	}

	if indented, ok := prettyJSON([]byte(parts[1])); ok {
		slog.Debug("HTTP Request", slog.String("headers", parts[0]), slog.String("body", indented))
		return
	}

	slog.Debug("HTTP Request", slog.String("dump", string(requestDump)))
}

func logResponse(recorder *responseRecorder) {
	contentType := recorder.Header().Get("Content-Type")
	bodyStr := recorder.body.String()

	if !strings.HasPrefix(contentType, "application/json") {
		slog.Debug("HTTP Response", slog.Int("status", recorder.statusCode), slog.String("body", bodyStr))
		return
	}

	if indented, ok := prettyJSON(recorder.body.Bytes()); ok {
		slog.Debug("HTTP Response", slog.Int("status", recorder.statusCode), slog.String("body", indented))
		return
	}

	slog.Debug("HTTP Response", slog.Int("status", recorder.statusCode), slog.String("body", bodyStr))
}

func prettyJSON(b []byte) (string, bool) {
	var indented bytes.Buffer
	if err := json.Indent(&indented, b, "", "  "); err != nil {
		return "", false
	}
	return indented.String(), true
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
