package qlogger

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// GetTraceIDFromGinCtx extracts the trace ID from gin.Context.
// It checks in the following order:
// 1. OpenTelemetry span context
// 2. trace_id from context value
// 3. X-Amzn-Trace-Id header (AWS X-Ray format)
// 4. X-Amzn-TraceId header (alternative format)
// 5. X-Amz-Firehose-TraceId header (AWS Firehose)
// 7. X-Trace-ID header
// 8. X-Request-ID header
// 9. trace_id context value (backward compatibility)
//
// If the trace ID contains "Root=1-x-y" format, it returns "xy" (x+y concatenated).
// Otherwise, it returns the trace ID as-is.
func GetTraceIDFromGinCtx(c *gin.Context) string {
	if c == nil {
		return ""
	}

	ctx := c.Request.Context()
	var traceID string

	// Try OpenTelemetry span context first (already in correct format)
	span := trace.SpanFromContext(ctx)
	if span != nil {
		spanContext := span.SpanContext()
		if spanContext.IsValid() {
			return spanContext.TraceID().String()
		}
	}

	// Try custom trace ID from context (typed key)
	if traceID == "" {
		if val, ok := ctx.Value(TraceIDKey).(string); ok && val != "" {
			traceID = val
		}
	}

	// Try AWS X-Ray trace ID from context (typed key)
	if val, ok := ctx.Value(AWSTraceIDKey).(string); ok && val != "" {
		traceID = val
	}

	// Try AWS X-Ray trace ID from header (X-Amzn-Trace-Id)
	if traceID == "" {
		if val := c.Request.Header.Get(HeaderXAmznTraceID); val != "" {
			traceID = val
		}
	}

	// Try AWS X-Ray trace ID from header (X-Amzn-TraceId - alternative format)
	if traceID == "" {
		if val := c.Request.Header.Get(HeaderXAmznTraceIDV2); val != "" {
			traceID = val
		}
	}

	// Try AWS Firehose trace ID from header (X-Amz-Firehose-TraceId)
	if traceID == "" {
		if val := c.Request.Header.Get(HeaderXAmzFirehoseTrace); val != "" {
			traceID = val
		}
	}


	// Try X-Trace-ID header
	if traceID == "" {
		if val := c.Request.Header.Get(HeaderXTraceID); val != "" {
			traceID = val
		}
	}

	// Try X-Request-ID header
	if traceID == "" {
		if val := c.Request.Header.Get(HeaderXRequestID); val != "" {
			traceID = val
		}
	}

	// Fallback: try string key for backward compatibility
	if traceID == "" {
		if val, ok := ctx.Value("x-trace-id").(string); ok && val != "" {
			traceID = val
		}
	}

	// Parse and normalize the trace ID
	// If it contains "Root=1-x-y" format, extract and return "xy"
	// Otherwise, return as-is
	return normalizeTraceID(traceID)
}

// addGinDefaultFields adds standard fields like traceId to log entries.
func addGinDefaultFields(c *gin.Context, fields ...Field) []Field {
	traceID := GetTraceIDFromGinCtx(c)
		fields = append(fields, String("trace_id", traceID))
	return fields
}

// --- Gin logging methods ---

// InfoGin logs a message at Info level using gin.Context.
func (l *Logger) InfoGin(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Info(msg, fields...)
}

// ErrorGin logs a message at Error level using gin.Context.
func (l *Logger) ErrorGin(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Error(msg, fields...)
}

// WarnGin logs a message at Warn level using gin.Context.
func (l *Logger) WarnGin(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Warn(msg, fields...)
}

// DebugGin logs a message at Debug level using gin.Context.
// In production, debug logs are suppressed.
func (l *Logger) DebugGin(c *gin.Context, msg string, fields ...Field) {
	if l.environment == "prod" || l.environment == "production" {
		return
	}
	fields = addGinDefaultFields(c, fields...)
	l.log.Debug(msg, fields...)
}

// FatalGin logs a message at Fatal level using gin.Context, then exits.
func (l *Logger) FatalGin(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Fatal(msg, fields...)
}

// PanicGin logs a message at Panic level using gin.Context, then panics.
func (l *Logger) PanicGin(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Panic(msg, fields...)
}
