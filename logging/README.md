# Logging

The `logging` package provides a structured logging wrapper around Go's standard `log/slog` library. It includes features for metrics reporting, context-aware logging, component-specific handlers, and HTTP middleware.

## Features

- **Structured Logging**: Supports both text and JSON formats.
- **Dynamic Configuration**: Configure log levels, format, and source information via command-line flags.
- **Metrics Integration**: Automatically tracks log statistics (counts by level) and the total number of bytes logged.
- **Context-Aware**: Allows attaching attributes to a `context.Context` that are automatically included in log entries.
- **Component Logging**: Specific handlers for different components with independent log level control.
- **HTTP Middleware**: Middleware for logging full HTTP requests and responses, including pretty-printed JSON bodies.

## Usage

### Initialization

Initialize the global logger at the start of your application. This parses command-line flags and sets the default `slog` logger.

```go
import "github.com/taylorono/go-lib/logging"

func main() {
    logging.InitLogger(context.Background())
    slog.Info("Application started", "version", "1.0.0")
}
```

**Text Output (Default):**
```text
time=2026-07-11T10:22:30.052-06:00 level=INFO msg="Application started" version=1.0.0
```

**JSON Output (`-log-json`):**
```json
{"time":"2026-07-11T10:22:35.5809063-06:00","level":"INFO","msg":"Application started","version":"1.0.0"}
```

#### Command-line Flags

- `-log-level`: Set the minimum log level (`debug`, `info`, `warn`, `error`). Default is `info`.
- `-log-json`: Enable JSON structured logging. Default is `false` (text).
- `-log-source`: Include source file and line number in log entries. Default is `false`.

### Component-Specific Logging

Use `ComponentLoggerFor` to create a logger for a specific part of your application.

```go
logger := logging.ComponentLoggerFor("my-component")
logger.Info("Message from component")
```

**Output:**
```text
time=2026-07-11T10:22:30.053-06:00 level=INFO msg="Message from component"
```

You can dynamically control component log levels by providing an enabled function.  When using the config library, this will automatically configure the log level for each component based on the configuration `{{MY_NAME}}_LOGGING`.
```go
logging.WithEnabledFunction(config.GetLogLevel)
```

### Context-Based Attributes

Attach attributes to a context that will be automatically included in any logs that use that context.

```go
ctx := logging.WithLogContext(context.Background(), slog.String("request_id", "12345"))
slog.InfoContext(ctx, "Processed request")
```

**Output:**
```text
time=2026-07-11T10:22:30.053-06:00 level=INFO msg="Processed request" request_id=12345
```

### Metrics

To enable metrics reporting, provide a `MetricsReporter` implementation

```go
prometheusReporter := metrics.NewPrometheusReporter()
logging.WithMetricReporter(prometheusReporter)
```

You can also include custom metrics in log entries:

```go
slog.InfoContext(ctx, "Operation completed", logging.Metric("operation_status", "success"))
```

This will expose the bytes logged metrics as well as a metric event for any log event that contains a logging.Metric Attribute
```text
# HELP logged_bytes_total amount logged in bytes
# TYPE logged_bytes_total counter
logged_bytes_total 7976
# HELP logger log statistics
# TYPE logger counter
logger{loglevel="INFO",message="success",name="operation_status"} 3
```

### HTTP Middleware

Use `HttpLoggingMiddleware` to log details about incoming HTTP requests and outgoing responses.

```go
http.HandleFunc("/api/data", logging.HttpLoggingMiddleware(myHandler))
```

It automatically detects `application/json` content types and pretty-prints the bodies in the logs at `DEBUG` level.

**Sample HTTP Logs:**
```text
time=2026-07-11T10:22:30.053-06:00 level=DEBUG msg="HTTP Request" headers="POST /api/test HTTP/1.1\r\nHost: example.com\r\nContent-Type: application/json" body="{\n  \"key\": \"value\"\n}"
time=2026-07-11T10:22:30.053-06:00 level=DEBUG msg="HTTP Response" status=200 body="{\n  \"status\": \"ok\"\n}"
```
