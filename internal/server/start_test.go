package server

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/web"
	"go.uber.org/fx"
)

var testConfig = Config{
	FS:        afero.NewMemMapFs(),
	Listener:  router.Config{},
	Assets:    assets.Config{},
	Storage:   storage.Config{Path: ":memory:"},
	Tools:     tools.Config{Log: logger.Config{Output: io.Discard}},
	Web:       web.Config{},
	MasterKey: masterkey.Config{DevMode: true},
	Folder:    "/foo",
}

func TestServerStart(t *testing.T) {
	ctx := context.Background()

	app := start(ctx, testConfig, fx.Invoke(func(*router.API) {}))
	require.NoError(t, app.Err())
}

func TestServerRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		Run(ctx, testConfig)
	}()

	req, err := http.NewRequest(http.MethodGet, "http://localhost:8797/login", nil)
	require.NoError(t, err)

	var res *http.Response
	for i := 0; i < 50; i++ {
		res, err = http.DefaultClient.Do(req)
		if err == nil || !strings.Contains(err.Error(), "connection refused") {
			break
		}

		if res != nil {
			res.Body.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	wg.Wait()
}
