package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"go.uber.org/fx"
)

func TestServerStart(t *testing.T) {
	cfg := NewDefaultConfig()

	app := start(cfg, fx.Invoke(func(*router.API) {}))
	assert.NoError(t, app.Err())
}
