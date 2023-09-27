package server

import (
	"testing"

	"github.com/neilotoole/slogt"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"go.uber.org/fx"
)

func TestServerStart(t *testing.T) {
	fs := afero.NewMemMapFs()
	cfg := NewDefaultConfig()
	log := slogt.New(t)

	db, err := storage.Init(fs, &cfg.Storage, log)
	require.NoError(t, err)

	app := start(cfg, db, afero.NewMemMapFs(), fx.Invoke(func(*router.API) {}))
	assert.NoError(t, app.Err())
}
