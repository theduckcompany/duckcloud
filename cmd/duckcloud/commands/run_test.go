package commands

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewRunCmd(t *testing.T) {
	t.Run("success with default args", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		t.Setenv("COUCHDB_DEBUG", "true")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// --memory-fs is used to leave no trace to the host
		cmd.SetArgs([]string{"--memory-fs", "--folder=/duckcloud-test"})
		var cmdErr error
		var wg sync.WaitGroup
		go func() {
			wg.Add(1)
			defer wg.Done()
			cmdErr = cmd.ExecuteContext(ctx)
		}()

		req, err := http.NewRequest(http.MethodGet, "http://localhost:5764/", nil)
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

		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		cancel()
		wg.Wait()
		assert.NoError(t, cmdErr)
	})

	t.Run("with some env variable setup", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		t.Setenv("DUCKCLOUD_HTTP-PORT", "8797")
		t.Setenv("DUCKCLOUD_LOG-LEVEL", "info")
		t.Setenv("DUCKCLOUD_FOLDER", "duckloud-test")

		cmd.SetArgs([]string{"--memory-fs"})
		var cmdErr error
		var wg sync.WaitGroup
		go func() {
			wg.Add(1)
			defer wg.Done()
			cmdErr = cmd.ExecuteContext(ctx)
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

		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		cancel()
		wg.Wait()
		assert.NoError(t, cmdErr)
	})

	t.Run("with a self-signed-certificate", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cmd.SetArgs([]string{"--self-signed-cert", "--memory-fs", "--folder=/duckcloud-test", "--log-level=info"})
		var cmdErr error
		var wg sync.WaitGroup
		go func() {
			wg.Add(1)
			defer wg.Done()
			cmdErr = cmd.ExecuteContext(ctx)
		}()

		// As we use a self-signed certificate we need to use a client with some verifications
		// removed.
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: tr}

		req, err := http.NewRequest(http.MethodGet, "https://localhost:5764/", nil)
		require.NoError(t, err)

		var res *http.Response
		for i := 0; i < 50; i++ {
			res, err = client.Do(req)
			if err == nil || !strings.Contains(err.Error(), "connection refused") {
				break
			}

			if res != nil {
				res.Body.Close()
			}
			time.Sleep(20 * time.Millisecond)
		}

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		cancel()
		wg.Wait()
		assert.NoError(t, cmdErr)
	})

	t.Run("with --self-signed-args and --tls-key should failed", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		cmd.SetErr(io.Discard)
		cmd.SetOut(io.Discard)

		cmd.SetArgs([]string{"--self-signed-cert", "--tls-key=/foo/bar", "--memory-fs", "--folder=/foobar"})
		err := cmd.Execute()

		assert.EqualError(t, err, ErrConflictTLSConfig.Error())
	})
}
