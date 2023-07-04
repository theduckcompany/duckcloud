package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestServerStart(t *testing.T) {
	cfg := NewDefaultConfig()

	app := start(cfg, fx.Invoke(func(*http.Server) {}))
	assert.NoError(t, app.Err())
}
