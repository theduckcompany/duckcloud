package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx/fxevent"
)

func TestFxLogger(t *testing.T) {
	t.Parallel()

	someError := errors.New("some error")

	tests := []struct {
		give        fxevent.Event
		wantFields  map[string]interface{}
		name        string
		wantMessage string
	}{
		{
			name: "OnStopExecutedError",
			give: &fxevent.OnStopExecuted{
				FunctionName: "hook.onStart1",
				CallerName:   "bytes.NewBuffer",
				Err:          fmt.Errorf("some error"),
			},
			wantMessage: "OnStop hook failed",
			wantFields: map[string]interface{}{
				"error": "some error",
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
				"error": "some error",
			},
		},
		{
			name:        "SuppliedError",
			give:        &fxevent.Supplied{TypeName: "*bytes.Buffer", Err: someError},
			wantMessage: "error encountered while applying options",
			wantFields: map[string]interface{}{
				"error": "some error",
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
			name: "Replace/Error",
			give: &fxevent.Replaced{Err: someError},

			wantMessage: "error encountered while replacing",
			wantFields: map[string]interface{}{
				"error": "some error",
			},
		},
		{
			name:        "Invoked/Error",
			give:        &fxevent.Invoked{Err: someError},
			wantMessage: "invoke failed",
			wantFields: map[string]interface{}{
				"error": "some error",
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
