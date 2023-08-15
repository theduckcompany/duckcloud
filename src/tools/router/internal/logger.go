package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// StructuredLogger is a simple, but powerful implementation of a custom structured
// logger backed on log/slog. I encourage users to copy it, adapt it and make it their
// own. Also take a look at https://github.com/go-chi/httplog for a dedicated pkg based
// on this work, designed for context-based http routers.

func NewStructuredLogger(log *slog.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&structuredLogger{logger: log.Handler()})
}

type structuredLogger struct {
	logger slog.Handler
}

func (l *structuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	var logFields []slog.Attr
	logFields = append(logFields, slog.String("ts", time.Now().UTC().Format(time.RFC1123)))

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, slog.String("req_id", reqID))
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	handler := l.logger.WithGroup("http").WithAttrs(append(logFields,
		slog.String("scheme", scheme),
		slog.String("proto", r.Proto),
		slog.String("method", r.Method),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("user_agent", r.UserAgent()),
		slog.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI))))

	return &structuredLoggerEntry{logger: slog.New(handler)}
}

type structuredLoggerEntry struct {
	logger *slog.Logger
}

func (l *structuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, "request complete",
		slog.Int("resp_status", status),
		slog.Int("resp_byte_length", bytes),
		slog.Float64("resp_elapsed_ms", float64(elapsed.Nanoseconds())/1000000.0),
	)
}

func (l *structuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.logger.LogAttrs(context.Background(), slog.LevelInfo, "",
		slog.String("stack", string(stack)),
		slog.String("panic", fmt.Sprintf("%+v", v)),
	)
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.
//
// This is a useful pattern to use to set state on the entry as it
// passes through the handler chain, which at any point can be logged
// with a call to .Print(), .Info(), etc.

func GetLogEntry(r *http.Request) *slog.Logger {
	entry := middleware.GetLogEntry(r).(*structuredLoggerEntry)
	return entry.logger
}

func LogEntrySetField(r *http.Request, key string, value interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*structuredLoggerEntry); ok {
		entry.logger = entry.logger.With(key, value)
	}
}

func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
	if entry, ok := r.Context().Value(middleware.LogEntryCtxKey).(*structuredLoggerEntry); ok {
		for k, v := range fields {
			entry.logger = entry.logger.With(k, v)
		}
	}
}
