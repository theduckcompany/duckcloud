package config

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestConfig(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)
	svc := NewService(store)

	t.Run("EnableTLS", func(t *testing.T) {
		require.NoError(t, svc.EnableTLS(ctx))

		res, err := svc.IsTLSEnabled(ctx)
		assert.NoError(t, err)
		assert.True(t, res)
	})

	t.Run("DisableTLS", func(t *testing.T) {
		require.NoError(t, svc.DisableTLS(ctx))

		res, err := svc.IsTLSEnabled(ctx)
		assert.NoError(t, err)
		assert.False(t, res)
	})

	t.Run("EnableDevMode", func(t *testing.T) {
		// Off by default
		res, err := svc.IsDevModeEnabled(ctx)
		assert.False(t, res)
		assert.ErrorIs(t, err, ErrNotInitialized)

		// Enable it
		require.NoError(t, svc.EnableDevMode(ctx))

		res, err = svc.IsDevModeEnabled(ctx)
		assert.NoError(t, err)
		assert.True(t, res)
	})

	t.Run("SetSSLPaths with a disabled tls", func(t *testing.T) {
		require.NoError(t, svc.DisableTLS(ctx))

		err := svc.SetSSLPaths(ctx, "foo", "bar")
		assert.ErrorIs(t, err, ErrSSLMustBeEnabled)
	})

	t.Run("SetSSLPaths with some invalid file", func(t *testing.T) {
		require.NoError(t, svc.EnableTLS(ctx))

		dir := t.TempDir()
		filePath := path.Join(dir, "invalid.pem")
		err := os.WriteFile(filePath, []byte("Hello, World!"), 0o644)
		require.NoError(t, err)

		err = svc.SetSSLPaths(ctx, filePath, filePath)
		assert.ErrorIs(t, err, ErrInvalidPEMFormat)
	})

	t.Run("SetSSLPaths success", func(t *testing.T) {
		require.NoError(t, svc.EnableTLS(ctx))

		current, err := os.Getwd()
		require.NoError(t, err)
		dir := path.Join(current, "/testdata")

		err = svc.SetSSLPaths(ctx, path.Join(dir, "cert.pem"), path.Join(dir, "key.pem"))
		assert.NoError(t, err)
	})

	t.Run("GetSSLPaths success", func(t *testing.T) {
		certif, key, err := svc.GetSSLPaths(ctx)
		assert.NoError(t, err)

		current, err := os.Getwd()
		require.NoError(t, err)
		assert.Equal(t, path.Join(current, "/testdata/cert.pem"), certif)
		assert.Equal(t, path.Join(current, "/testdata/key.pem"), key)
	})

	t.Run("SetTrustedHosts success", func(t *testing.T) {
		err := svc.SetTrustedHosts(ctx, []string{"localhost", "foo.bar.baz", "127.0.0.1"})
		assert.NoError(t, err)
	})

	t.Run("GetTrustedHosts", func(t *testing.T) {
		res, err := svc.GetTrustedHosts(ctx)
		assert.NoError(t, err)
		assert.Equal(t, []string{"localhost", "foo.bar.baz", "127.0.0.1"}, res)
	})

	t.Run("SetTrustedHosts with an invalid ip", func(t *testing.T) {
		// Invalid port number
		err := svc.SetTrustedHosts(ctx, []string{"127.0.0.1:1"})
		assert.EqualError(t, err, `invalid host "127.0.0.1:1": must be a valid IP address or DNS name`)
	})

	t.Run("SetTrustedHosts with an invalid host", func(t *testing.T) {
		// Invalid port number
		err := svc.SetTrustedHosts(ctx, []string{"https://foo.bar"})
		assert.EqualError(t, err, `invalid host "https://foo.bar": must be a valid IP address or DNS name`)
	})

	t.Run("SetHostName success", func(t *testing.T) {
		err := svc.SetHostName(ctx, "foo.bar.baz:8080")
		assert.NoError(t, err)
	})

	t.Run("GetHostName success", func(t *testing.T) {
		res, err := svc.GetHostName(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "foo.bar.baz:8080", res)
	})

	t.Run("SetHostName with an invalid ip", func(t *testing.T) {
		err := svc.SetHostName(ctx, "127.0.0.1:1:1")
		assert.EqualError(t, err, `invalid hostname: must be a valid IP address or DNS name`)
	})

	t.Run("SetAddrs success", func(t *testing.T) {
		err := svc.SetAddrs(ctx, []string{"127.0.0.1", "::1"}, 8080)
		require.NoError(t, err)
	})

	t.Run("GetAddrs success", func(t *testing.T) {
		res, err := svc.GetAddrs(ctx)
		require.NoError(t, err)
		assert.Equal(t, []string{"127.0.0.1:8080", "[::1]:8080"}, res)
	})

	t.Run("SetAddrs with an invalid port", func(t *testing.T) {
		err := svc.SetAddrs(ctx, []string{"::1"}, -1)
		assert.ErrorIs(t, err, ErrInvalidPort)
	})

	t.Run("SetAddrs with an invalid host", func(t *testing.T) {
		err := svc.SetAddrs(ctx, []string{":::"}, 8080)
		assert.EqualError(t, err, "invalid host \":::\": must be a valid IP address or DNS name")
	})
}
