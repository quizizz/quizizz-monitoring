# qlogger

A production-ready, high-performance logging package for Go built on top of [Uber's Zap](https://github.com/uber-go/zap) with configurable log sampling support.

## Features

- **Built on Zap**: Leverages the performance and reliability of Uber's Zap logger
- **Configurable Sampling**: Control log volume with per-level sampling rules
- **Context-Aware**: Automatic trace ID propagation via context
- **OpenTelemetry Integration**: Native support for OpenTelemetry trace context
- **Gin Framework Support**: First-class support for Gin web framework
- **AWS X-Ray Support**: Automatic parsing of AWS X-Ray trace IDs
- **Functional Options**: Flexible configuration using the functional options pattern

## Installation

To install qlogger:

```bash
go get github.com/quizizz/quizizz-monitoring/qlogger
```

## Quick Start

### Development Environment

```go
package main

import (
    "context"
    
    "github.com/quizizz/quizizz-monitoring/qlogger"
)

func main() {
    // Create a development logger (colored, human-readable output)
    logger, err := qlogger.NewDevelopment()
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    ctx := context.Background()
    ctx = qlogger.WithTraceID(ctx, "abc-123-xyz")
    
    // Log with trace ID from context
    logger.InfoWithTraceIdCtx(ctx, "Application started",
        qlogger.String("version", "1.0.0"),
    )
    
    // Log without context
    logger.Info("Simple log message",
        qlogger.String("key", "value"),
    )
}
```

### Production Environment

```go
package main

import (
    "context"
    "time"
    
    "github.com/quizizz/quizizz-monitoring/qlogger"
    "github.com/quizizz/quizizz-monitoring/qlogger/sampler"
)

func main() {
    // Create a production logger with sampling
    logger, err := qlogger.NewProduction(
        sampler.WithInfoSampling(100, 10),  // First 100/sec, then 1 in 10
        sampler.WithWarnSampling(50, 5),    // First 50/sec, then 1 in 5
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    ctx := context.Background()
    ctx = qlogger.WithTraceID(ctx, "request-trace-123")
    
    logger.InfoWithTraceIdCtx(ctx, "Request processed",
        qlogger.String("user_id", "12345"),
        qlogger.Duration("latency", time.Millisecond*150),
    )
}
```

### Using Options Pattern

```go
// Auto-detect environment from ENV variable
logger, err := qlogger.New(
    qlogger.WithEnvironment("prod"),        // or "local", "dev", "development"
    qlogger.WithInfoSampling(100, 10),
    qlogger.WithWarnSampling(50, 5),
)
```

## Logger Constructors

| Constructor | Description |
|-------------|-------------|
| `NewDevelopment()` | Colored, human-readable output. No sampling. |
| `NewProduction(opts...)` | JSON output with configurable sampling. |
| `NewNop()` | No-op logger that discards all logs (for testing). |
| `New(opts...)` | Auto-detect based on environment options. |
| `MustNew(opts...)` | Same as `New()` but panics on error. |

## Configuration Options

| Option | Description |
|--------|-------------|
| `WithEnvironment(env)` | Set environment ("prod", "staging", "local", "dev") |
| `WithDevelopmentMode(bool)` | Enable/disable development mode |
| `WithInfoSampling(initial, thereafter)` | Configure Info level sampling |
| `WithWarnSampling(initial, thereafter)` | Configure Warn level sampling |
| `WithDebugSampling(initial, thereafter)` | Configure Debug level sampling |
| `WithErrorSampling(initial, thereafter)` | Configure Error level sampling (use cautiously) |
| `WithCaller(enabled)` | Enable/disable caller information |
| `WithCallerSkip(skip)` | Set stack frames to skip for caller info |

### Sampling Explained

Sampling controls log volume by limiting how many log entries at each level are recorded:

- **Initial**: The number of logs per second that are always recorded
- **Thereafter**: After `Initial`, only every Nth log is recorded

Example: `WithInfoSampling(100, 10)` means:
- First 100 Info logs per second are always logged
- After 100, only every 10th Info log is recorded

**Note**: Sampling is only applied in production mode.

## Context Support

### Trace ID Propagation

The logger automatically extracts trace IDs from context in the following order:

1. OpenTelemetry span context
2. Custom trace ID (`trace_id` context key)
3. AWS X-Ray trace ID (`X-Amzn-Trace-Id` header)
4. AWS Firehose trace ID (`X-Amz-Firehose-TraceId` header)
5. Custom headers (`X-Trace-ID`, `X-Request-ID`)

```go
// Set trace ID manually
ctx = qlogger.WithTraceID(ctx, "abc-123-xyz")

// From HTTP request headers
ctx = qlogger.WithHTTPRequest(ctx, req)

// From AWS X-Ray
ctx = qlogger.WithAWSTraceID(ctx, "Root=1-5759e988-bd862e3fe1be46a994272793")

logger.InfoWithTraceIdCtx(ctx, "Request processed") // trace_id automatically included
```

### AWS X-Ray Support

AWS X-Ray trace IDs are automatically parsed from the `Root=1-x-y` format:

```go
// Input: "Root=1-5759e988-bd862e3fe1be46a994272793"
// Extracted trace_id: "5759e988bd862e3fe1be46a994272793"
```

## Gin Framework Support

### Gin Logging Methods

```go
import "github.com/quizizz/quizizz-monitoring/qlogger"

func MyHandler(c *gin.Context) {
    // Automatically extracts trace ID from gin.Context headers
    logger.InfoWithTraceId(c, "Processing request",
        qlogger.String("path", c.Request.URL.Path),
    )
    
    logger.ErrorWithTraceId(c, "Something went wrong",
        qlogger.Err(err),
    )
}
```

### Gin Middleware Example

```go
func LoggingMiddleware(logger *qlogger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        logger.InfoWithTraceId(c, "http-response",
            qlogger.String("method", c.Request.Method),
            qlogger.String("path", c.Request.URL.Path),
            qlogger.Int("status", c.Writer.Status()),
            qlogger.Duration("latency", time.Since(start)),
        )
    }
}
```

### Gin Methods Reference

| Method | Description |
|--------|-------------|
| `logger.InfoWithTraceId(c, msg, fields...)` | Info log with gin.Context |
| `logger.ErrorWithTraceId(c, msg, fields...)` | Error log with gin.Context |
| `logger.WarnWithTraceId(c, msg, fields...)` | Warn log with gin.Context |
| `logger.DebugWithTraceId(c, msg, fields...)` | Debug log with gin.Context |
| `logger.FatalWithTraceId(c, msg, fields...)` | Fatal log with gin.Context |
| `logger.PanicWithTraceId(c, msg, fields...)` | Panic log with gin.Context |
| `logger.DPanicWithTraceId(c, msg, fields...)` | DPanic log with gin.Context |
| `qlogger.GetTraceIDFromGinCtx(c)` | Extract trace ID from gin.Context |

## Field Constructors

The package provides convenient field constructors (aliases to zap):

```go
// Common fields
qlogger.String("key", "value")
qlogger.Int("key", 42)
qlogger.Int64("key", int64(42))
qlogger.Float64("key", 3.14)
qlogger.Bool("key", true)
qlogger.Any("key", someStruct)
qlogger.Err(err)                    // Error field (key: "error")

// Additional fields
qlogger.Duration("latency", time.Second)
qlogger.Time("timestamp", time.Now())
qlogger.Stringer("addr", netAddr)
qlogger.Int32("port", 8080)
qlogger.Uint("count", 100)
qlogger.Uint64("bytes", 1024)
qlogger.Binary("data", []byte{...})
qlogger.ByteString("raw", []byte("hello"))
qlogger.Namespace("request")
qlogger.Stack("stacktrace")
qlogger.Object("user", userMarshaler)
qlogger.Array("items", itemsMarshaler)
qlogger.Reflect("config", configStruct)
```

You can also use `zap.Field` directly:

```go
import "go.uber.org/zap"

logger.Info(ctx, "message", zap.String("key", "value"))
```

## Logging Methods

### With Trace ID from Context

```go
logger.InfoWithTraceIdCtx(ctx, msg, fields...)
logger.WarnWithTraceIdCtx(ctx, msg, fields...)
logger.ErrorWithTraceIdCtx(ctx, msg, fields...)
logger.DebugWithTraceIdCtx(ctx, msg, fields...)  // Suppressed in prod
logger.FatalWithTraceIdCtx(ctx, msg, fields...)  // Calls os.Exit(1)
logger.PanicWithTraceIdCtx(ctx, msg, fields...)  // Panics
logger.DPanicWithTraceIdCtx(ctx, msg, fields...) // Panics in dev, logs in prod
```

### Without Context (Standard Logging)

```go
logger.Info(msg, fields...)
logger.Warn(msg, fields...)
logger.Error(msg, fields...)
logger.Debug(msg, fields...)   // Suppressed in prod
logger.Fatal(msg, fields...)   // Calls os.Exit(1)
logger.Panic(msg, fields...)   // Panics
logger.DPanic(msg, fields...)  // Panics in dev, logs in prod
```

## Utility Methods

```go
// Create child logger with fields
childLogger := logger.With(qlogger.String("component", "auth"))

// Create logger with context fields
ctxLogger := logger.WithContext(ctx)

// Access underlying zap.Logger
zapLogger := logger.Zap()

// Get environment
env := logger.Environment()

// Flush buffered logs (call before exit)
logger.Sync()
```

## Context Helper Functions

```go
// Set trace ID in context
ctx = qlogger.WithTraceID(ctx, "trace-123")

// Set AWS X-Ray trace ID
ctx = qlogger.WithAWSTraceID(ctx, "Root=1-xxx-yyy")

// Store HTTP request for header extraction
ctx = qlogger.WithHTTPRequest(ctx, req)

// Extract trace ID from context
traceID := qlogger.GetTraceIDFromCtx(ctx)

// Extract trace ID from HTTP request directly
traceID := qlogger.ExtractTraceIDFromRequest(req)

// Extract trace ID from Gin context
traceID := qlogger.GetTraceIDFromGinCtx(c)
```

## Project Structure

```
quizizz-monitoring/
├── go.mod
├── go.sum
├── README.md
└── qlogger/
    ├── logger.go           # Main logger implementation
    ├── context.go          # Context helpers (trace ID extraction)
    ├── gin_context.go      # Gin framework support
    ├── options.go          # Logger configuration options
    ├── sampler/
    │   ├── rule.go         # Sampling rule definition
    │   ├── config.go       # Sampler configuration
    │   ├── options.go      # Functional options for sampler
    │   └── builder.go      # Builder pattern implementation
    └── examples/
        ├── basic/
        │   └── main.go
        ├── custom_sampling/
        │   └── main.go
        └── http_middleware/
            └── main.go
```

## Best Practices

1. **Choose the Right Constructor**: Use `NewDevelopment()` locally, `NewProduction()` in production
2. **Defer Sync**: Always `defer logger.Sync()` after creation
3. **Dependency Injection**: Pass logger to components that need it
4. **Use Context**: Pass context through your application for trace correlation
5. **Don't Sample Errors**: Avoid sampling error logs unless absolutely necessary
6. **Tune Sampling**: Adjust sampling based on your traffic patterns

## Output Format

### Development Mode

```
2024-01-15T10:30:45.123Z    INFO    myapp/main.go:42    Request processed    {"trace_id": "abc-123", "user_id": "12345"}
```

### Production Mode (JSON)

```json
{"level":"info","ts":1705312245.123,"caller":"myapp/main.go:42","msg":"Request processed","trace_id":"abc-123","user_id":"12345"}
```

## Contributing

Contributions are welcome! Please ensure your code follows Go best practices and includes appropriate tests.

## License

MIT License - see LICENSE file for details.
