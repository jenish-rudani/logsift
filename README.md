# logsift

A Go logging library that wraps [logrus](https://github.com/sirupsen/logrus) with **filter-based conditional logging**, automatic source tracking, structured fields, and runtime configuration via HTTP.

## Features

- **Filter-based logging** — selectively enable/disable log output by topic
- **Automatic source tracking** — every log line includes file and line number
- **Structured fields** — attach key-value context with `With` / `WithFields`
- **Multiple output formats** — JSON, text, colored, and no-color
- **HTTP config handler** — change log level, format, and filters at runtime without restarting
- **Prometheus metrics** — built-in error counter for observability
- **Thread-safe filters** — concurrent-safe filter implementation by default
- **Drop-in logger interface** — use the package-level API or inject `Logger` instances

## Installation

```sh
go get github.com/jenish-rudani/logsift
```

## Quick Start

```go
package main

import "github.com/jenish-rudani/logsift"

func main() {
    logsift.SetLevel("debug")
    logsift.SetFormat("text")

    logsift.Info("server starting")
    logsift.Debugf("listening on port %d", 8080)
    logsift.Error("something went wrong")
}
```

## Configuration

### Log Level

```go
logsift.SetLevel("debug") // trace, debug, info, warn, error, fatal, panic
level := logsift.GetLevel()
```

Invalid levels default to `info`.

### Output Format

```go
logsift.SetFormat("json") // json, text, nocolor, forceColor
format := logsift.GetFormat()
```

| Format       | Description                          |
|--------------|--------------------------------------|
| `json`       | JSON structured output               |
| `text`       | Text with auto-detected colors (default) |
| `nocolor`    | Text without colors                  |
| `forceColor` | Text with forced color output        |

### Source Format

```go
logsift.SetSourceFormat("short") // short (default) or long
```

- `short` — filename and line: `main.go:42`
- `long` — full path: `/home/user/app/main.go:42`

### Output Writer

```go
logsift.SetOutput(os.Stderr)
```

## Filtered Logging

Filters let you selectively enable log output for specific topics or modules without changing log levels.

```go
// Enable filters
logsift.AddFilter("auth")
logsift.AddFilter("db")

// These only log if their filter is active
logsift.DebugFilter("auth", "token refreshed")
logsift.InfoFilter("db", "query executed")
logsift.DebugFilter("cache", "this won't print — filter not active")

// Multiple filters — logs if any one matches
logsift.DebugFilters([]string{"auth", "cache"}, "will print because auth is active")

// Formatted variants
logsift.DebugFilterf("auth", "user %s logged in", "alice")
logsift.InfoFilterf("db", "query took %dms", 42)

// Manage filters
logsift.RemoveFilter("db")
logsift.UpdateFilter(map[string]bool{"auth": true, "api": true}) // replace all

// Control behavior when no filters are set
logsift.SetAllowEmptyFilter(true) // if true, filtered logs pass when filter map is empty
```

## Structured Fields

```go
// Single field
logger := logsift.With("request_id", "abc-123")
logger.Info("handling request")

// Multiple fields
logger = logsift.WithFields(map[string]interface{}{
    "user":   "alice",
    "action": "login",
})
logger.Info("user authenticated")
```

## HTTP Runtime Configuration

Expose an endpoint to change logging configuration at runtime:

```go
http.Handle("/log", logsift.Handler())
http.ListenAndServe(":8080", nil)
```

Then adjust via query parameters:

```
GET /log?level=debug&format=json
GET /log?filter=auth,db&allowEmptyFilter=false
GET /log?resetFilter=true
```

| Parameter          | Type   | Description                            |
|--------------------|--------|----------------------------------------|
| `level`            | string | Set log level                          |
| `format`           | string | Set output format                      |
| `sourceFormat`     | string | Set source format (`short` / `long`)   |
| `filter`           | string | Comma-separated filters to enable      |
| `allowEmptyFilter` | bool   | Allow logging when no filters are set  |
| `resetFilter`      | bool   | Clear all active filters               |

## Prometheus Metrics

logsift exposes a Prometheus counter for tracking logged errors:

```go
// Metric: service_error_counter
// Labels: line
logsift.ErrorCounter.WithLabelValues("main.go:55").Inc()
```

Register with your Prometheus setup as needed — the counter is created via `promauto` and auto-registers with the default registry.

## Logger Interface

Use the `Logger` interface for dependency injection:

```go
func NewService(log logsift.Logger) *Service {
    log.Info("service created")
    return &Service{log: log}
}

// Use the default logger
svc := NewService(logsift.Default())

// Or a logger with pre-attached fields
svc := NewService(logsift.With("component", "service"))
```

## Filter Implementations

Two `Filter` implementations are available:

| Implementation           | Thread-Safe | Use Case                                    |
|--------------------------|-------------|---------------------------------------------|
| `NewConcurrentMapFilter` | Yes         | Default — safe for concurrent goroutines    |
| `NewUnsafeMapFilter`     | No          | Single-threaded or externally synchronized  |

The default logger uses `ConcurrentMapFilter`.

## Logrus Interop

```go
// Access the underlying logrus entry
entry := logsift.Entry()

// Add logrus hooks
logsift.AddHook(myHook)
```

## Examples

Runnable examples are in the [examples/](examples/) directory:

| Example | Description | Run |
|---------|-------------|-----|
| [basic](examples/basic/) | Log levels, output formats, source tracking | `go run ./examples/basic` |
| [filters](examples/filters/) | Filter-based conditional logging | `go run ./examples/filters` |
| [fields](examples/fields/) | Structured fields and Logger interface DI | `go run ./examples/fields` |
| [httpconfig](examples/httpconfig/) | Runtime config via HTTP endpoint | `go run ./examples/httpconfig` |

## License

MIT
