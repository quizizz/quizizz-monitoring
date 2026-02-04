# qlogger

A production-ready, high-performance logging package for Go built on top of [Uber's Zap](https://github.com/uber-go/zap) with configurable log sampling support.

## Features

- **Built on Zap**: Leverages the performance and reliability of Uber's Zap logger
- **Configurable Sampling**: Control log volume with per-level sampling rules
- **Context-Aware**: Automatic trace ID propagation via context
- **OpenTelemetry Integration**: Native support for OpenTelemetry trace context
- **Gin Framework Support**: First-class support for Gin web framework
- **SOLID Principles**: Clean, extensible architecture following SOLID design principles
- **Functional Options**: Flexible configuration using the functional options pattern

## Installation

```bash
go get github.com/wayground/wlogger
```

## Quick Start

```go
package main

import (
    "context"
    
    "github.com/wayground/wlogger"
    "go.uber.org/zap"
)

func main() {
    // Create logger instance
    logger, err := wlogger.New(
        wlogger.WithEnvironment("prod"),
        wlogger.WithInfoSampling(100, 10),  // First 100/sec, then 1 in 10
        wlogger.WithWarnSampling(50, 5),    // First 50/sec, then 1 in 5
    )
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    ctx := context.Background()
    ctx = wlogger.WithTraceID(ctx, "abc-123-xyz")
    
    // Log with context - automatically includes traceId
    logger.Info(ctx, "Application started",
        zap.String("version", "1.0.0"),
    )
    
    // Log without context
    logger.InfoWithoutCtx("Configuration loaded",
        zap.Int("items", 42),
    )
}
```

### HTTP Response Logging

```go
logger.Info(ctx, "http",
    zap.String("type", "http-response"),
    zap.Any("data", HTTPResponseData{
        Req: parseRequest(c),
        Res: parseResponse(c),
    }),
)
```

## Configuration

### Initialization Options

| Option | Description |
|--------|-------------|
| `WithEnvironment(env)` | Set environment ("prod", "staging", "local", "dev") |
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

**Note**: If no sampling options are configured, no sampling is performed (all logs pass through).

### Custom Sampling with Sampler Package

For fine-grained control, use the sampler package directly:

```go
import (
    "github.com/wayground/wlogger"
    "github.com/wayground/wlogger/sampler"
    "go.uber.org/zap/zapcore"
)

// Define custom sampling options
samplingOpts := []sampler.Option{
    sampler.WithInfoSampling(50, 5),
    sampler.WithWarnSampling(20, 2),
    sampler.WithSampling(zapcore.DebugLevel, 10, 10),
    // sampler.WithErrorSampling(10, 2), // Not recommended
}

logger, err := wlogger.New(
    wlogger.WithEnvironment("prod"),
    wlogger.WithSamplerOptions(samplingOpts...),
)
```

## Context Support

### Trace ID Propagation

The logger automatically extracts trace IDs from context in the following order:
1. OpenTelemetry span context
2. AWS X-Ray trace ID (`X-Amzn-Trace-Id` header)
3. AWS Firehose trace ID (`X-Amz-Firehose-TraceId` header)
4. Custom trace ID (`X-Trace-ID` or `X-Request-ID` header)

```go
// From OpenTelemetry span context (automatic)
span := trace.SpanFromContext(ctx)

// Or set manually
ctx = wlogger.WithTraceID(ctx, "abc-123-xyz")

logger.Info(ctx, "Request processed") // traceId automatically included
```

### AWS X-Ray Support

AWS X-Ray trace IDs are automatically parsed from the `Root=1-x-y` format:

```go
// Input: "Root=1-5759e988-bd862e3fe1be46a994272793"
// Extracted traceId: "5759e988bd862e3fe1be46a994272793"
```

### HTTP Request Context

Store HTTP request in context for automatic header-based trace ID extraction:

```go
ctx = wlogger.WithHTTPRequest(ctx, req)
logger.Info(ctx, "Processing request") // Extracts trace ID from headers
```

## Gin Framework Support

The package provides first-class support for the Gin web framework.

### Gin Logging Methods

```go
import "github.com/wayground/wlogger"

func MyHandler(c *gin.Context) {
    // Automatically extracts trace ID from headers:
    // - X-Amzn-Trace-Id (AWS X-Ray)
    // - X-Amzn-TraceId (alternative format)
    // - X-Amz-Firehose-TraceId (AWS Firehose)
    // - X-Trace-ID
    // - X-Request-ID
    
    logger.InfoGin(c, "Processing request",
        wlogger.String("path", c.Request.URL.Path),
    )
    
    logger.ErrorGin(c, "Something went wrong",
        wlogger.Err(err),
    )
}
```

### Gin Middleware Example

```go
func LoggingMiddleware(logger *wlogger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        logger.InfoGin(c, "http-response",
            wlogger.String("method", c.Request.Method),
            wlogger.String("path", c.Request.URL.Path),
            wlogger.Int("status", c.Writer.Status()),
            wlogger.Int64("latency_ms", time.Since(start).Milliseconds()),
        )
    }
}
```

### Gin Methods Reference

| Method | Description |
|--------|-------------|
| `logger.InfoGin(c, msg, fields...)` | Info log with gin.Context |
| `logger.ErrorGin(c, msg, fields...)` | Error log with gin.Context |
| `logger.WarnGin(c, msg, fields...)` | Warn log with gin.Context |
| `logger.DebugGin(c, msg, fields...)` | Debug log with gin.Context |
| `logger.FatalGin(c, msg, fields...)` | Fatal log with gin.Context |
| `logger.PanicGin(c, msg, fields...)` | Panic log with gin.Context |
| `wlogger.GetTraceIDFromGinCtx(c)` | Extract trace ID from gin.Context |

## API Reference

### Field Constructors

The package provides convenient field constructors (aliases to zap):

```go
wlogger.String("key", "value")     // String field
wlogger.Int("key", 42)             // Int field
wlogger.Int64("key", int64(42))    // Int64 field
wlogger.Float64("key", 3.14)       // Float64 field
wlogger.Bool("key", true)          // Bool field
wlogger.Any("key", someStruct)     // Any type field
wlogger.Err(err)                   // Error field (key: "error")
```

You can also use `zap.Field` directly:

```go
import "go.uber.org/zap"

logger.Info(ctx, "message", zap.String("key", "value"))
```

### Logging Methods

#### Context-Aware

```go
logger.Info(ctx, msg, fields...)
logger.Warn(ctx, msg, fields...)
logger.Error(ctx, msg, fields...)
logger.Debug(ctx, msg, fields...)  // Suppressed in prod
logger.Fatal(ctx, msg, fields...)  // Calls os.Exit(1)
logger.Panic(ctx, msg, fields...)  // Panics
```

#### Without Context

```go
logger.InfoWithoutCtx(msg, fields...)
logger.WarnWithoutCtx(msg, fields...)
logger.ErrorWithoutCtx(msg, fields...)
logger.DebugWithoutCtx(msg, fields...)
logger.FatalWithoutCtx(msg, fields...)
logger.PanicWithoutCtx(msg, fields...)
```

### Utility Methods

```go
// Create child logger with fields
childLogger := logger.With(zap.String("component", "auth"))

// Create logger with context fields
ctxLogger := logger.WithContext(ctx)

// Access underlying zap.Logger
zapLogger := logger.Zap()

// Get environment
env := logger.Environment()

// Flush buffered logs (call before exit)
logger.Sync()
```

### Creating Logger

```go
// With error handling
logger, err := wlogger.New(
    wlogger.WithEnvironment("prod"),
    wlogger.WithInfoSampling(100, 10),
)
if err != nil {
    panic(err)
}

// Or panic on error (use when configuration is known to be valid)
logger := wlogger.MustNew(
    wlogger.WithEnvironment("prod"),
)
```

## Architecture

The package follows SOLID principles:

### Single Responsibility
- `sampler/rule.go`: Sampling rule definition
- `sampler/config.go`: Sampler configuration
- `sampler/options.go`: Functional options for sampler
- `sampler/builder.go`: Builder pattern for creating sampled loggers
- `context.go`: Context helper functions
- `gin_context.go`: Gin framework context helpers
- `options.go`: Logger configuration options
- `logger.go`: Main logger implementation

### Open/Closed Principle
- Functional options pattern allows extension without modification
- Custom sampling rules can be added for any `zapcore.Level`

### Dependency Inversion
- Core logger depends on abstractions (zapcore interfaces)
- Easy to swap underlying components

## Project Structure

```
wlogger/
├── go.mod
├── go.sum
├── README.md
├── logger.go           # Main logger interface and methods
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

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ENV` | Environment name (prod, staging, local, dev) | `prod` |

## Best Practices

1. **Create Once**: Create logger instance at application startup
2. **Defer Sync**: Always `defer logger.Sync()` after creation
3. **Dependency Injection**: Pass logger to components that need it
4. **Use Context**: Pass context through your application for trace correlation
5. **Don't Sample Errors**: Avoid sampling error logs unless absolutely necessary
6. **Tune Sampling**: Adjust sampling based on your traffic patterns

## Comparison with Raw Zap

| Feature | Raw Zap | wlogger |
|---------|---------|---------|
| Sampling | Manual setup | Built-in, configurable |
| Context support | Manual | Automatic trace ID propagation |
| OpenTelemetry | Manual | Native integration |
| Gin support | Manual | Built-in methods |
| Environment handling | Manual | Automatic dev/prod modes |

## Contributing

Contributions are welcome! Please ensure your code follows Go best practices and includes appropriate tests.

## License

MIT License - see LICENSE file for details.
