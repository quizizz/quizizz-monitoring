package qlogger

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)


// If the trace ID contains "Root=1-x-y" format, it returns "xy" (x+y concatenated).
// Otherwise, it returns the trace ID as-is.
func GetTraceIDFromGinCtx(c *gin.Context) string {
	if c == nil {
		return ""
	}

	ctx := c.Request.Context()
	var traceID string

	

	// Try custom trace ID from context (typed key)
	if traceID == "" {
		if val, ok := ctx.Value(TraceIDKey).(string); ok && val != "" {
			traceID = val
		}
	}

	// Try AWS X-Ray trace ID from context (typed key)
	if traceID == "" {
		if val, ok := ctx.Value(AWSTraceIDKey).(string); ok && val != "" {
			traceID = val
		}
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

	// Try OpenTelemetry span context first (already in correct format)
	if traceID == "" {
		span := trace.SpanFromContext(ctx)
			if span != nil {
				spanContext := span.SpanContext()
				if spanContext.IsValid() {
					return spanContext.TraceID().String()
				}
			}
    }

	return NormalizeTraceID(traceID)
}

// addGinDefaultFields adds standard fields like traceId to log entries.
func addGinDefaultFields(c *gin.Context, fields ...Field) []Field {
	traceID := GetTraceIDFromGinCtx(c)
	if traceID != "" {
		fields = append(fields, String("trace_id", traceID))
	}
	return fields
}

// --- Gin logging methods ---

// InfoWithTraceId logs a message at Info level using gin.Context.
func (l *Logger) InfoWithTraceId(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Info(msg, fields...)
}

// ErrorWithTraceId logs a message at Error level using gin.Context.
func (l *Logger) ErrorWithTraceId(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Error(msg, fields...)
}

// WarnWithTraceId logs a message at Warn level using gin.Context.
func (l *Logger) WarnWithTraceId(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Warn(msg, fields...)
}

// DebugWithTraceId logs a message at Debug level using gin.Context.
// In production, debug logs are suppressed.
func (l *Logger) DebugWithTraceId(c *gin.Context, msg string, fields ...Field) {
	if l.environment == "prod" || l.environment == "production" {
		return
	}
	fields = addGinDefaultFields(c, fields...)
	l.log.Debug(msg, fields...)
}

// FatalWithTraceId logs a message at Fatal level using gin.Context, then exits.
func (l *Logger) FatalWithTraceId(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Fatal(msg, fields...)
}

// PanicWithTraceId logs a message at Panic level using gin.Context, then panics.
func (l *Logger) PanicWithTraceId(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.Panic(msg, fields...)
}

// DPanicWithTraceId logs a message at DPanic level using gin.Context, then panics.
func (l *Logger) DPanicWithTraceId(c *gin.Context, msg string, fields ...Field) {
	fields = addGinDefaultFields(c, fields...)
	l.log.DPanic(msg, fields...)
}