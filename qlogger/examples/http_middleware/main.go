package main

import (
	"net/http"
	"time"

	"go.quizizz.org/monitoring/qlogger"
	"go.uber.org/zap"
)

// HTTPRequestData represents logged request data
type HTTPRequestData struct {
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     string            `json:"query,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	UserAgent string            `json:"user_agent,omitempty"`
}

// HTTPResponseData represents logged response data
type HTTPResponseData struct {
	Status    int           `json:"status"`
	Latency   time.Duration `json:"latency"`
	LatencyMs int64         `json:"latency_ms"`
}

// logger is the application logger instance
var logger *qlogger.Logger

// LoggingMiddleware wraps an HTTP handler with logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Extract or generate trace ID
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}

		// Add trace ID to context
		ctx := qlogger.WithTraceID(r.Context(), traceID)
		r = r.WithContext(ctx)

		// Create response wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log request
		logger.Info(ctx, "http-request",
			zap.Any("request", HTTPRequestData{
				Method:    r.Method,
				Path:      r.URL.Path,
				Query:     r.URL.RawQuery,
				UserAgent: r.UserAgent(),
			}),
		)

		// Call next handler
		next.ServeHTTP(rw, r)

		// Log response
		latency := time.Since(start)
		logger.Info(ctx, "http-response",
			zap.String("type", "http-response"),
			zap.Any("response", HTTPResponseData{
				Status:    rw.statusCode,
				Latency:   latency,
				LatencyMs: latency.Milliseconds(),
			}),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func generateTraceID() string {
	// In production, use a proper UUID generator
	return "generated-trace-id"
}

func main() {
	// Create logger instance with HTTP-appropriate sampling
	var err error
	logger, err = qlogger.New(
		qlogger.WithEnvironment("prod"),
		qlogger.WithInfoSampling(200, 10), // Higher initial for HTTP traffic
		qlogger.WithWarnSampling(100, 5),
	)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Use logger with context in your handlers
		logger.Info(ctx, "Processing request",
			zap.String("handler", "main"),
		)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Wrap with logging middleware
	http.Handle("/", LoggingMiddleware(handler))

	logger.InfoWithoutCtx("Server starting", zap.String("addr", ":8080"))
	http.ListenAndServe(":8080", nil)
}
