package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RouterLogger(t *testing.T) {
	t.Run("Check the basic informations", func(t *testing.T) {
		buf := new(bytes.Buffer)

		logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		handler := NewRouterLogger(logger)

		srv := chi.NewMux()
		srv.Handle("/*", handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("Hello, World!"))
			w.WriteHeader(http.StatusOK)
		})))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/browsers", nil)
		srv.ServeHTTP(w, r)

		res := map[string]any{}
		err := json.Unmarshal(buf.Bytes(), &res)
		require.NoError(t, err)

		t.Logf("Log line: %+v", buf.String())

		assert.Equal(t, "request complete", res["msg"])
		assert.Equal(t, "DEBUG", res["level"])
		logTime, err := time.Parse(time.RFC3339, res["time"].(string))
		require.NoError(t, err)
		assert.WithinDuration(t, time.Now(), logTime, 10*time.Millisecond)

		httpRes := res["http"].(map[string]any)
		assert.Equal(t, "http", httpRes["scheme"])
		assert.Equal(t, "HTTP/1.1", httpRes["proto"])
		assert.Equal(t, "GET", httpRes["method"])
		assert.Equal(t, "192.0.2.1:1234", httpRes["remote_addr"])
		assert.Equal(t, "", httpRes["user_agent"])
		assert.Equal(t, "http://example.com/settings/browsers", httpRes["uri"])
		assert.EqualValues(t, 13, httpRes["resp_byte_length"])
		assert.EqualValues(t, http.StatusOK, httpRes["resp_status"])
		assert.Less(t, httpRes["resp_elapsed_ms"], float64(0.05))
	})

	t.Run("Check the location field for redirects", func(t *testing.T) {
		buf := new(bytes.Buffer)

		logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		handler := NewRouterLogger(logger)

		srv := chi.NewMux()
		srv.Handle("/*", handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/foobar", http.StatusFound)
		})))

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/settings/browsers", nil)
		srv.ServeHTTP(w, r)

		res := map[string]any{}
		err := json.Unmarshal(buf.Bytes(), &res)
		require.NoError(t, err)

		t.Logf("Log line: %+v", buf.String())

		assert.Equal(t, "request complete", res["msg"])
		assert.Equal(t, "INFO", res["level"])

		httpRes := res["http"].(map[string]any)
		assert.Equal(t, "/foobar", httpRes["location"])
	})
}
