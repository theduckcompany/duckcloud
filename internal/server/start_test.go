package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
	"go.uber.org/fx"
)

func TestServerStart(t *testing.T) {
	ctx := context.Background()

	serv := startutils.NewServer(t)
	serv.Bootstrap(t)

	app := start(ctx, serv.DB, serv.FS, "/test-dir", fx.Invoke(func(*router.API) {}))
	assert.NoError(t, app.Err())
}
