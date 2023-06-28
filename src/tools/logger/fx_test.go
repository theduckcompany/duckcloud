package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"golang.org/x/exp/slog"
)

func TestFxLogger(t *testing.T) {
	t.Parallel()

	someError := errors.New("some error")

	tests := []struct {
		name        string
		give        fxevent.Event
		wantMessage string
		wantFields  map[string]interface{}
	}{
		{
			name: "OnStartExecuting",
			give: &fxevent.OnStartExecuting{
				FunctionName: "hook.onStart",
				CallerName:   "bytes.NewBuffer",
			},
			wantMessage: "OnStart hook executing",
			wantFields: map[string]interface{}{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStart",
			},
		},
		{
			name: "OnStopExecuting",
			give: &fxevent.OnStopExecuting{
				FunctionName: "hook.onStop1",
				CallerName:   "bytes.NewBuffer",
			},
			wantMessage: "OnStop hook executing",
			wantFields: map[string]interface{}{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStop1",
			},
		},
		{
			name: "OnStopExecutedError",
			give: &fxevent.OnStopExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Err:          fmt.Errorf("some error"),
			},
			wantMessage: "OnStop hook failed",
			wantFields: map[string]interface{}{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStart1",
				"error":  "some error",
			},
		},
		{
			name: "OnStopExecuted",
			give: &fxevent.OnStopExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Runtime:      time.Millisecond * 3,
			},
			wantMessage: "OnStop hook executed",
			wantFields: map[string]interface{}{
				"caller":  "bytes.NewBuffer",
				"callee":  "hook.onStart1",
				"runtime": "3ms",
			},
		},
		{
			name: "OnStartExecutedError",
			give: &fxevent.OnStartExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Err:          fmt.Errorf("some error"),
			},
			wantMessage: "OnStart hook failed",
			wantFields: map[string]interface{}{
				"caller": "bytes.NewBuffer",
				"callee": "hook.onStart1",
				"error":  "some error",
			},
		},
		{
			name: "OnStartExecuted",
			give: &fxevent.OnStartExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Runtime:      time.Millisecond * 3,
			},
			wantMessage: "OnStart hook executed",
			wantFields: map[string]interface{}{
				"caller":  "bytes.NewBuffer",
				"callee":  "hook.onStart1",
				"runtime": "3ms",
			},
		},
		{
			name:        "Supplied",
			give:        &fxevent.Supplied{TypeName: "*bytes.Buffer"},
			wantMessage: "supplied",
			wantFields: map[string]interface{}{
				"type": "*bytes.Buffer",
			},
		},
		{
			name:        "SuppliedError",
			give:        &fxevent.Supplied{TypeName: "*bytes.Buffer", Err: someError},
			wantMessage: "error encountered while applying options",
			wantFields: map[string]interface{}{
				"type":  "*bytes.Buffer",
				"error": "some error",
			},
		},
		{
			name: "Provide",
			give: &fxevent.Provided{
				ConstructorName: "bytes.NewBuffer()",
				ModuleName:      "myModule",
				OutputTypeNames: []string{"*bytes.Buffer"},
			},
			wantMessage: "provided",
			wantFields: map[string]interface{}{
				"constructor": "bytes.NewBuffer()",
				"type":        "*bytes.Buffer",
				"module":      "myModule",
			},
		},
		{
			name:        "Provide with Error",
			give:        &fxevent.Provided{Err: someError},
			wantMessage: "error encountered while applying options",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name: "Replace",
			give: &fxevent.Replaced{
				ModuleName:      "myModule",
				OutputTypeNames: []string{"*bytes.Buffer"},
			},
			wantMessage: "replaced",
			wantFields: map[string]interface{}{
				"type":   "*bytes.Buffer",
				"module": "myModule",
			},
		},
		{
			name: "Replace/Error",
			give: &fxevent.Replaced{Err: someError},

			wantMessage: "error encountered while replacing",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name: "Decorate",
			give: &fxevent.Decorated{
				DecoratorName:   "bytes.NewBuffer()",
				ModuleName:      "myModule",
				OutputTypeNames: []string{"*bytes.Buffer"},
			},
			wantMessage: "decorated",
			wantFields: map[string]interface{}{
				"decorator": "bytes.NewBuffer()",
				"type":      "*bytes.Buffer",
				"module":    "myModule",
			},
		},
		{
			name:        "Decorate with Error",
			give:        &fxevent.Decorated{Err: someError},
			wantMessage: "error encountered while applying options",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name:        "Invoking/Success",
			give:        &fxevent.Invoking{ModuleName: "myModule", FunctionName: "bytes.NewBuffer()"},
			wantMessage: "invoking",
			wantFields: map[string]interface{}{
				"function": "bytes.NewBuffer()",
				"module":   "myModule",
			},
		},
		{
			name:        "Invoked/Error",
			give:        &fxevent.Invoked{FunctionName: "bytes.NewBuffer()", Err: someError},
			wantMessage: "invoke failed",
			wantFields: map[string]interface{}{
				"error":    "some error",
				"stack":    "",
				"function": "bytes.NewBuffer()",
			},
		},
		{
			name:        "StartError",
			give:        &fxevent.Started{Err: someError},
			wantMessage: "start failed",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name:        "Stopping",
			give:        &fxevent.Stopping{Signal: os.Interrupt},
			wantMessage: "received signal",
			wantFields: map[string]interface{}{
				"signal": "INTERRUPT",
			},
		},
		{
			name:        "Stopped",
			give:        &fxevent.Stopped{Err: someError},
			wantMessage: "stop failed",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name:        "RollingBack",
			give:        &fxevent.RollingBack{StartErr: someError},
			wantMessage: "start failed, rolling back",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name:        "RolledBackError",
			give:        &fxevent.RolledBack{Err: someError},
			wantMessage: "rollback failed",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name:        "Started",
			give:        &fxevent.Started{},
			wantMessage: "started",
			wantFields:  map[string]interface{}{},
		},
		{
			name:        "LoggerInitialized Error",
			give:        &fxevent.LoggerInitialized{Err: someError},
			wantMessage: "custom logger initialization failed",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name:        "LoggerInitialized",
			give:        &fxevent.LoggerInitialized{ConstructorName: "bytes.NewBuffer()"},
			wantMessage: "initialized custom fxevent.Logger",
			wantFields: map[string]interface{}{
				"function": "bytes.NewBuffer()",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: slog.LevelDebug,
				ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
					if attr.Key == slog.TimeKey || attr.Key == slog.LevelKey {
						return slog.Attr{}
					}
					return attr
				},
			})
			l := NewFxLogger(slog.New(handler))

			l.LogEvent(tt.give)

			var got map[string]interface{}
			require.NoError(t, json.Unmarshal(buf.Bytes(), &got))

			assert.Equal(t, tt.wantMessage, got["msg"])
			delete(got, "msg")
			assert.Equal(t, tt.wantFields, got)
		})
	}
}

func ExampleNew() {
	app := fx.New(
		fx.Provide(func() *slog.Logger {
			return slog.New(slog.NewJSONHandler(os.Stdout, nil))
		}),
		fx.WithLogger(func(logger *slog.Logger) fxevent.Logger {
			return NewFxLogger(logger)
		}),
	)
	defer app.Stop(context.TODO())
	app.Start(context.TODO())
}
