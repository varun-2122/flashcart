package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/varun-2122/flashcart/internal/logger"
	"github.com/varun-2122/flashcart/internal/response"
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
