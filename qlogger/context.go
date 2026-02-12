package qlogger

import (
	"context"
	"net/http"
	"strings"
	"go.opentelemetry.io/otel/trace"
)

// ContextKey is a custom type for context keys to avoid collisions.
type ContextKey string

const (
	// TraceIDKey is the context key for trace ID.
	TraceIDKey ContextKey = "trace_id"

	// AWSTraceIDKey is the context key for AWS X-Ray trace ID.
	AWSTraceIDKey ContextKey = "X-Amzn-Trace-Id"

	// HTTPRequestKey is the context key for storing *http.Request.
	HTTPRequestKey ContextKey = "http-request"
)

// Header names for trace ID extraction
const (
	HeaderXAmznTraceID       = "X-Amzn-Trace-Id"
	HeaderXAmznTraceIDV2     = "X-Amzn-TraceId"
	HeaderXAmzFirehoseTrace  = "X-Amz-Firehose-TraceId"
	HeaderXTraceID           = "X-Trace-ID"
	HeaderXRequestID         = "X-Request-ID"
)

// GetTraceIDFromCtx extracts the trace ID from the context.
// It checks in the following order:
// 1. OpenTelemetry span context
// 2. AWS X-Ray trace ID from context value (typed key)
// 3. AWS X-Ray trace ID from context value (string key for backward compatibility)
// 4. HTTP headers from stored *http.Request (X-Amzn-Trace-Id, X-Amzn-TraceId, X-Amz-Firehose-TraceId)
// 5. Custom trace ID from context value (typed key)
// 6. HTTP headers: X-Trace-ID
// 7. HTTP headers: X-Request-ID
// 8. x-trace-id context value (backward compatibility)
//
// If the trace ID contains "Root=1-x-y" format, it returns "xy" (x+y concatenated).
// Otherwise, it returns the trace ID as-is.
func GetTraceIDFromCtx(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	var traceID string

	

	// Try AWS X-Ray (trace_id) from context (typed key)
	if traceID == "" {
		if val, ok := ctx.Value(TraceIDKey).(string); ok && val != "" {
			traceID = val
		}
	}

	// Try AWS X-Ray trace ID from context ( for backward compatibility)
	if traceID == "" {
		if val, ok := ctx.Value(AWSTraceIDKey).(string); ok && val != "" {
			traceID = val
		}
	}

	// Fallback: try to extract from HTTP request headers stored in context
	if traceID == "" {
		if req, ok := ctx.Value(HTTPRequestKey).(*http.Request); ok && req != nil {
			// Try AWS X-Ray trace ID (X-Amzn-Trace-Id)
			if val := req.Header.Get(HeaderXAmznTraceID); val != "" {
				traceID = val
			}
			// Try AWS X-Ray trace ID (alternative format X-Amzn-TraceId)
			if traceID == "" {
				if val := req.Header.Get(HeaderXAmznTraceIDV2); val != "" {
					traceID = val
				}
			}
			// Try AWS Firehose trace ID (X-Amz-Firehose-TraceId)
			if traceID == "" {
				if val := req.Header.Get(HeaderXAmzFirehoseTrace); val != "" {
					traceID = val
				}
			}
		}
	}

	// Try custom trace ID from context (typed key)
	if traceID == "" {
		if val, ok := ctx.Value(TraceIDKey).(string); ok && val != "" {
			traceID = val
		}
	}

	// Fallback: try to extract X-Trace-ID from HTTP request headers
	if traceID == "" {
		if req, ok := ctx.Value(HTTPRequestKey).(*http.Request); ok && req != nil {
			if val := req.Header.Get(HeaderXTraceID); val != "" {
				traceID = val
			}
		}
	}

	// Fallback: try to extract X-Request-ID from HTTP request headers
	if traceID == "" {
		if req, ok := ctx.Value(HTTPRequestKey).(*http.Request); ok && req != nil {
			if val := req.Header.Get(HeaderXRequestID); val != "" {
				traceID = val
			}
		}
	}

	// Fallback: try string key for backward compatibility
	if traceID == "" {
		if val, ok := ctx.Value("x-trace-id").(string); ok && val != "" {
			traceID = val
		}
	}
	// Try OpenTelemetry span context 
	if traceID ==  "" {
		span := trace.SpanFromContext(ctx)
		if span != nil {
			spanContext := span.SpanContext()
			if spanContext.IsValid() {
				return spanContext.TraceID().String()
			}
		}
	}
	
	

	// Parse and normalize the trace ID
	// If it contains "Root=1-x-y" format, extract and return "xy"
	// Otherwise, return as-is
	return NormalizeTraceID(traceID)
}

// getTraceIDFromHeaders extracts trace ID from HTTP headers.
// It checks headers in the following order:
// 1. X-Amzn-Trace-Id (AWS X-Ray format)
// 2. X-Amzn-TraceId (alternative format)
// 3. X-Amz-Firehose-TraceId (AWS Firehose)
// 4. X-Trace-ID
// 5. X-Request-ID
//
// If the trace ID contains "Root=1-x-y" format, it returns "xy" (x+y concatenated).
// Otherwise, it returns the trace ID as-is.
func getTraceIDFromHeaders(headers http.Header) string {
	var traceID string

	// Try AWS X-Ray trace ID (X-Amzn-Trace-Id)
	if traceID == "" {
		if val := headers.Get(HeaderXAmznTraceID); val != "" {
			traceID = val
		}
	}

	// Try AWS X-Ray trace ID (alternative format X-Amzn-TraceId)
	if traceID == "" {
		if val := headers.Get(HeaderXAmznTraceIDV2); val != "" {
			traceID = val
		}
	}

	// Try AWS Firehose trace ID (X-Amz-Firehose-TraceId)
	if traceID == "" {
		if val := headers.Get(HeaderXAmzFirehoseTrace); val != "" {
			traceID = val
		}
	}

	// Try X-Trace-ID header
	if traceID == "" {
		if val := headers.Get(HeaderXTraceID); val != "" {
			traceID = val
		}
	}

	// Try X-Request-ID header
	if traceID == "" {
		if val := headers.Get(HeaderXRequestID); val != "" {
			traceID = val
		}
	}

	// Parse and normalize the trace ID
	return NormalizeTraceID(traceID)
}



// Example:
//
//	Input:  "Root=1-5759e988-bd862e3fe1be46a994272793"
//	Output: "5759e988bd862e3fe1be46a994272793"
//
//	Input:  "Self=1-698c225b-59083319108f1ccd0dc8599b;Root=1-f448e00d-bac3c863c2aea11848aeda61;Parent=834acb99c8c80028;Sampled=1"
//	Output: "f448e00dbac3c863c2aea11848aeda61"
func parseAWSTraceID(awsTraceID string) string {
	rootIndex := strings.Index(awsTraceID, "Root=")
	if rootIndex == -1 {
		return awsTraceID
	}

	// Extract everything after "Root="
	rootValue := awsTraceID[rootIndex+5:] // Skip "Root=" (5 chars)

	// Find the end of the Root field (marked by semicolon or end of string)
	if idx := strings.Index(rootValue, ";"); idx != -1 {
		rootValue = rootValue[:idx]
	}

	// Expected format: "1-x-y" where 1 is version, x is timestamp, y is unique id
	parts := strings.Split(rootValue, "-")
	if len(parts) != 3 {
		return ""
	}

	// parts[0] = version (e.g., "1")
	// parts[1] = x (timestamp, 8 hex chars)
	// parts[2] = y (unique id, 24 hex chars)
	// Return x + y concatenated
	return parts[1] + parts[2]
}

// NormalizeTraceID normalizes a trace ID to a consistent format.
// If the trace ID contains "Root=1-x-y" format (AWS X-Ray), it extracts and returns "xy".
// Otherwise, it returns the trace ID as-is.
//
// Examples:
//
//	Input:  "Root=1-5759e988-bd862e3fe1be46a994272793"
//	Output: "5759e988bd862e3fe1be46a994272793"
//
//
//	Input:  ""
//	Output: ""
func NormalizeTraceID(traceID string) string {
	if traceID == "" {
		return ""
	}

	// Check if it contains "Root=" prefix (AWS X-Ray format)
	if strings.Contains(traceID, "Root=") {
		if parsed := parseAWSTraceID(traceID); parsed != "" {
			return parsed
		}
	}

	// Return as-is if not AWS format or parsing failed
	return traceID
}

// WithAWSTraceID returns a new context with the AWS X-Ray trace ID.
// The trace ID should be in the format "Root=1-x-y" as received from the header.
func WithAWSTraceID(ctx context.Context, awsTraceID string) context.Context {
	return context.WithValue(ctx, AWSTraceIDKey, awsTraceID)
}

// WithHTTPRequest returns a new context with the HTTP request stored.
// This allows GetTraceIDFromCtx to extract trace IDs from request headers.
//
// Example:
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    ctx := wlogger.WithHTTPRequest(r.Context(), r)
//	    wlogger.Info(ctx, "handling request")
//	}
func WithHTTPRequest(ctx context.Context, req *http.Request) context.Context {
	return context.WithValue(ctx, HTTPRequestKey, req)
}

// ExtractTraceIDFromRequest extracts trace ID directly from an HTTP request's headers.
// This is a convenience function for cases where you have the request but not a context.
func ExtractTraceIDFromRequest(req *http.Request) string {
	if req == nil {
		return ""
	}
	return getTraceIDFromHeaders(req.Header)
}

// WithTraceID returns a new context with the given trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}
