package logger

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

func NewRouterLogger(log *slog.Logger) func(next http.Handler) http.Handler {
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

	return &structuredLoggerEntry{logger: handler}
}

type structuredLoggerEntry struct {
	logger slog.Handler
}

func (l *structuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	var level slog.Level

	entries := []slog.Attr{
		slog.Int("resp_status", status),
		slog.Int("resp_byte_length", bytes),
		slog.Float64("resp_elapsed_ms", float64(elapsed.Nanoseconds())/1000000.0),
	}

	switch {
	case status >= 200 && status <= 299:
		level = slog.LevelDebug
	case status >= 300 && status <= 399:
		entries = append(entries, slog.String("location", header.Get("Location")))

		level = slog.LevelInfo
	case status >= 400 && status <= 499:
		level = slog.LevelWarn
	default:
		level = slog.LevelError
	}

	slog.New(l.logger).LogAttrs(context.Background(), level, "request complete", entries...)
}

func (l *structuredLoggerEntry) Panic(v interface{}, stack []byte) {
	slog.New(l.logger).LogAttrs(context.Background(), slog.LevelError, "panic!",
		slog.String("stack", string(stack)),
		slog.String("panic", fmt.Sprintf("%+v", v)),
	)
}

func LogEntrySetAttrs(ctx context.Context, attrs ...slog.Attr) {
	if entry, ok := ctx.Value(middleware.LogEntryCtxKey).(*structuredLoggerEntry); ok {
		entry.logger = entry.logger.WithAttrs(attrs)
	}
}

func LogEntrySetError(ctx context.Context, err error) {
	if entry, ok := ctx.Value(middleware.LogEntryCtxKey).(*structuredLoggerEntry); ok {
		entry.logger = entry.logger.WithAttrs([]slog.Attr{slog.String("error", err.Error())})
	}
}
