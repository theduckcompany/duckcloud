package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/assets"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"go.uber.org/fx"
)

var testConfig = Config{
	FS:       afero.NewMemMapFs(),
	Listener: router.Config{},
	Assets:   assets.Config{},
	Storage:  sqlstorage.Config{Path: ":memory:"},
	Tools:    tools.Config{Log: logger.Config{Output: io.Discard}},
	HTML:     html.Config{},
	Folder:   "/foo",
}

func TestServerStart(t *testing.T) {
	ctx := context.Background()

	app := start(ctx, testConfig, fx.Invoke(func(*router.API) {}))
	require.NoError(t, app.Err())
}

func TestServerRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := startutils.GetFreePort(t)

	testConfig.Listener.Addr = fmt.Sprintf("localhost:%d", port)

	wg := sync.WaitGroup{}
	wg.Add(1)
	var runErr error
	go func() {
		defer wg.Done()
		_, runErr = Run(ctx, testConfig)
	}()

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/login", port), nil)
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

	require.NoError(t, runErr)
}
