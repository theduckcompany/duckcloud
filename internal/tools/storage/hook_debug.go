package storage

import (
	"context"
	"log/slog"
	"time"
)

type storageKey string

const beginKey storageKey = "begin"

type debugHooks struct {
	log *slog.Logger
}

func newDebugHook(log *slog.Logger) *debugHooks {
	return &debugHooks{log}
}

// Before hook will print the query with it's args and return the context with the timestamp
func (h *debugHooks) Before(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	return context.WithValue(ctx, beginKey, time.Now()), nil
}

// After hook will get the timestamp registered on the Before hook and print the elapsed time
func (h *debugHooks) After(ctx context.Context, query string, args ...interface{}) (context.Context, error) {
	begin := ctx.Value(beginKey).(time.Time)

	h.log.Debug(query, slog.Duration("duration", time.Since(begin)), slog.Any("args", args))

	return ctx, nil
}
