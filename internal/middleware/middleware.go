package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/metrics"
	"github.com/varun-2122/flashcart/internal/response"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// Middleware type alias for standard HTTP middleware functions.
type Middleware func(http.Handler) http.Handler

// Chain applies a slice of middleware functions in order.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// RequestID middleware assigns a unique request ID to context and response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			reqID = hex.EncodeToString(b)
		}

		ctx := context.WithValue(r.Context(), logger.RequestIDKey, reqID)
		w.Header().Set("X-Request-ID", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriterInterceptor captures HTTP status code for logging.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
	bytesWritten int
}

func (rw *responseWriterInterceptor) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterInterceptor) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

// Logger middleware logs incoming requests and outgoing status/duration in JSON.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start)

		logger.Info(r.Context(), "http request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapper.statusCode,
			"bytes", wrapper.bytesWritten,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

// Recovery middleware recovers from panics and logs stack traces cleanly.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				logger.Error(r.Context(), "panic recovered in HTTP handler",
					"error", fmt.Sprintf("%v", err),
					"stack", stack,
					"method", r.Method,
					"path", r.URL.Path,
				)

				response.InternalServerError(w, "An unexpected server error occurred")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Timeout middleware attaches context timeout to limit long-running HTTP operations.
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			panicChan := make(chan any, 1)

			go func() {
				defer func() {
					if p := recover(); p != nil {
						panicChan <- p
					}
				}()
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case p := <-panicChan:
				panic(p)
			case <-done:
				return
			case <-ctx.Done():
				if ctx.Err() == context.DeadlineExceeded {
					logger.Warn(r.Context(), "http request timed out",
						"method", r.Method,
						"path", r.URL.Path,
						"timeout", timeout.String(),
					)
					response.ErrorJSON(w, http.StatusGatewayTimeout, "REQUEST_TIMEOUT", "Request execution timed out", nil)
				}
			}
		})
	}
}

// CORS middleware handles cross-origin requests.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Metrics middleware instruments every HTTP request with Prometheus counters and
// latency histograms, labelled by HTTP method, path pattern, and status code.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapper := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapper, r)

		duration := time.Since(start).Seconds()
		path := r.URL.Path
		method := r.Method
		status := strconv.Itoa(wrapper.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}

// Tracing middleware intercepts HTTP requests, extracts any parent traces from headers,
// and starts a new OpenTelemetry span for the HTTP handler.
func Tracing(next http.Handler) http.Handler {
	tracer := otel.Tracer("flashcart-http")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract incoming trace context from HTTP headers
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// Add basic HTTP attributes
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.route", r.URL.Path),
		)

		// Create a response writer interceptor to capture the status code
		wrapper := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

		// Inject trace ID into request logger context if logger supports it
		if spanContext := span.SpanContext(); spanContext.IsValid() {
			traceID := spanContext.TraceID().String()
			w.Header().Set("X-Trace-ID", traceID)
			ctx = context.WithValue(ctx, logger.RequestIDKey, traceID)
		}

		next.ServeHTTP(wrapper, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.status_code", wrapper.statusCode))
		if wrapper.statusCode >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	})
}
