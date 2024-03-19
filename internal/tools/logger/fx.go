package logger

import (
	"context"
	"log/slog"
	"strings"

	"go.uber.org/fx/fxevent"
)

// FxLogger is an Fx event logger that logs events to slog.
type FxLogger struct {
	Logger *slog.Logger
}

func NewFxLogger(logger *slog.Logger) *FxLogger {
	return &FxLogger{
		Logger: logger,
	}
}

func (l *FxLogger) logError(msg string, fields ...slog.Attr) {
	l.Logger.LogAttrs(context.Background(), slog.LevelWarn, msg, fields...)
}

func (l *FxLogger) logEvent(msg string, fields ...slog.Attr) {
	l.Logger.LogAttrs(context.Background(), slog.LevelDebug, msg, fields...)
}

// LogEvent logs the given event to the provided Zap logger.
func (l *FxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logError("OnStart hook failed", slog.String("error", e.Err.Error()))
		}
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logError("OnStop hook failed", slog.String("error", e.Err.Error()))
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.logError("error encountered while applying options", slog.String("error", e.Err.Error()))
		}
	case *fxevent.Provided:
		if e.Err != nil {
			l.logError("error encountered while applying options", slog.String("error", e.Err.Error()))
		}
	case *fxevent.Replaced:
		if e.Err != nil {
			l.logError("error encountered while replacing", slog.String("error", e.Err.Error()))
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			l.logError("error encountered while applying options", slog.String("error", e.Err.Error()))
		}
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logError("invoke failed", slog.String("error", e.Err.Error()))
		}
	case *fxevent.Stopping:
		l.logEvent("received signal", slog.String("signal", strings.ToUpper(e.Signal.String())))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.logError("stop failed", slog.String("error", e.Err.Error()))
		}
	case *fxevent.RollingBack:
		l.logError("start failed, rolling back", slog.String("error", e.StartErr.Error()))
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.logError("rollback failed", slog.String("error", e.Err.Error()))
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.logError("start failed", slog.String("error", e.Err.Error()))
		} else {
			l.logEvent("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.logError("custom logger initialization failed", slog.String("error", e.Err.Error()))
		}
	}
}
