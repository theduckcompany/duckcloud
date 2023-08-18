package server

import (
	"testing"

	"github.com/myminicloud/myminicloud/src/tools/router"
	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"
)

func TestServerStart(t *testing.T) {
	cfg := NewDefaultConfig()

	app := start(cfg, fx.Invoke(func(*router.API) {}))
	assert.NoError(t, app.Err())
}
