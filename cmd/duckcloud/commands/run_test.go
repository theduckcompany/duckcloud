package commands

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
)

func Test_NewRunCmd(t *testing.T) {
	t.Setenv("DUCKCLOUD_DEV", "true")

	t.Run("success with default args", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		t.Setenv("DUCKCLOUD_DEBUG", "true")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		port := startutils.GetFreePort(t)

		// --memory-fs is used to leave no trace to the host
		cmd.SetArgs([]string{"--dev", "--memory-fs", "--folder=/duckcloud-test", fmt.Sprintf("--http-port=%d", port)})
		var cmdErr error
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmdErr = cmd.ExecuteContext(ctx)
		}()

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/", port), nil)
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
		require.NoError(t, cmdErr)
	})

	t.Run("with some env variable setup", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		port := startutils.GetFreePort(t)

		t.Setenv("DUCKCLOUD_HTTP-PORT", strconv.Itoa(port))
		t.Setenv("DUCKCLOUD_LOG-LEVEL", "info")
		t.Setenv("DUCKCLOUD_FOLDER", "duckloud-test")

		cmd.SetArgs([]string{"--memory-fs", "--dev"})
		var cmdErr error
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmdErr = cmd.ExecuteContext(ctx)
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

		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		cancel()
		wg.Wait()
		require.NoError(t, cmdErr)
	})

	t.Run("with a self-signed-certificate", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		port := startutils.GetFreePort(t)

		cmd.SetArgs([]string{"--self-signed-cert", "--memory-fs", "--dev", "--folder=/duckcloud-test", "--log-level=info", fmt.Sprintf("--http-port=%d", port)})
		var cmdErr error
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmdErr = cmd.ExecuteContext(ctx)
		}()

		// As we use a self-signed certificate we need to use a client with some verifications
		// removed.
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client := &http.Client{Transport: tr}

		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://localhost:%d/", port), nil)
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
		require.NoError(t, cmdErr)
	})

	t.Run("with --self-signed-args and --tls-key should failed", func(t *testing.T) {
		cmd := NewRunCmd("duckcloud-test")

		cmd.SetErr(io.Discard)
		cmd.SetOut(io.Discard)

		cmd.SetArgs([]string{"--self-signed-cert", "--tls-key=/foo/bar", "--memory-fs", "--dev", "--folder=/foobar"})
		err := cmd.Execute()

		require.EqualError(t, err, ErrConflictTLSConfig.Error())
	})
}
