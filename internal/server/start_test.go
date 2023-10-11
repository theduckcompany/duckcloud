package server

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"go.uber.org/fx"
)

func TestServerStart(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)

	app := start(ctx, db, afero.NewMemMapFs(), "/test-dir", fx.Invoke(func(*router.API) {}))
	assert.NoError(t, app.Err())
}
