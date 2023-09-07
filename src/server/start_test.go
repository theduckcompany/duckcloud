package server

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"go.uber.org/fx"
)

func TestServerStart(t *testing.T) {
	cfg := NewDefaultConfig()
	db := storage.NewTestStorage(t)

	app := start(cfg, db, afero.NewMemMapFs(), fx.Invoke(func(*router.API) {}))
	assert.NoError(t, app.Err())
}
