package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// RequestLogger provides structured logging for incoming HTTP requests.
// It captures status code, duration, bytes written, and selected request metadata.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			var panicValue any

			defer func() {
				if logger != nil {
					duration := time.Since(start)
					level := levelFromStatus(wrapped.status)
					id := chimw.GetReqID(r.Context())
					attrs := []slog.Attr{
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.Int("status", wrapped.status),
						slog.Float64("duration_ms", float64(duration)/float64(time.Millisecond)),
						slog.Int("bytes", wrapped.bytes),
						slog.String("remote_addr", r.RemoteAddr),
						slog.String("user_agent", r.UserAgent()),
					}
					if id != "" {
						attrs = append(attrs, slog.String("request_id", id))
					}
					if panicValue != nil {
						level = slog.LevelError
						attrs = append(attrs, slog.Any("panic", panicValue))
					}
					logger.LogAttrs(r.Context(), level, "http request", attrs...)
				}
				if panicValue != nil {
					panic(panicValue)
				}
			}()

			func() {
				defer func() {
					if rec := recover(); rec != nil {
						if !wrapped.wroteHeader {
							wrapped.status = http.StatusInternalServerError
						}
						panicValue = rec
					}
				}()
				next.ServeHTTP(wrapped, r)
			}()
		})
	}
}

func levelFromStatus(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

type responseWriter struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.status = http.StatusOK
		w.wroteHeader = true
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}
