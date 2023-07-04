package logger

import (
	"context"
	"strings"

	"go.uber.org/fx/fxevent"
	"golang.org/x/exp/slog"
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
	case *fxevent.OnStartExecuting:
		l.logEvent("OnStart hook executing",
			slog.String("callee", e.FunctionName),
			slog.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.logError("OnStart hook failed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("error", e.Err.Error()),
			)
		} else {
			l.logEvent("OnStart hook executed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.OnStopExecuting:
		l.logEvent("OnStop hook executing",
			slog.String("callee", e.FunctionName),
			slog.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.logError("OnStop hook failed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("error", e.Err.Error()),
			)
		} else {
			l.logEvent("OnStop hook executed",
				slog.String("callee", e.FunctionName),
				slog.String("caller", e.CallerName),
				slog.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.logError("error encountered while applying options",
				slog.String("type", e.TypeName),
				moduleName(e.ModuleName),
				slog.String("error", e.Err.Error()))
		} else {
			l.logEvent("supplied",
				slog.String("type", e.TypeName),
				moduleName(e.ModuleName),
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("provided",
				slog.String("constructor", e.ConstructorName),
				moduleName(e.ModuleName),
				slog.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while applying options",
				moduleName(e.ModuleName),
				slog.String("error", e.Err.Error()),
			)
		}
	case *fxevent.Replaced:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("replaced",
				moduleName(e.ModuleName),
				slog.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while replacing",
				moduleName(e.ModuleName),
				slog.String("error", e.Err.Error()),
			)
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.logEvent("decorated",
				slog.String("decorator", e.DecoratorName),
				moduleName(e.ModuleName),
				slog.String("type", rtype),
			)
		}
		if e.Err != nil {
			l.logError("error encountered while applying options",
				moduleName(e.ModuleName),
				slog.String("error", e.Err.Error()),
			)
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		l.logEvent("invoking",
			slog.String("function", e.FunctionName),
			moduleName(e.ModuleName),
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.logError("invoke failed",
				slog.String("error", e.Err.Error()),
				slog.String("stack", e.Trace),
				slog.String("function", e.FunctionName),
				moduleName(e.ModuleName),
			)
		}
	case *fxevent.Stopping:
		l.logEvent("received signal",
			slog.String("signal", strings.ToUpper(e.Signal.String())))
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
		} else {
			l.logEvent("initialized custom fxevent.Logger", slog.String("function", e.ConstructorName))
		}
	}
}

func moduleName(name string) slog.Attr {
	if name == "" {
		return slog.Attr{}
	}
	return slog.String("module", name)
}
