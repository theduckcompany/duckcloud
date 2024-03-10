package masterkey

import (
	"context"
	"path"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func Test_Integration_masterKey_with_manual_install(t *testing.T) {
	tools := tools.NewToolboxForTest(t)
	ctx := context.Background()
	afs := afero.NewMemMapFs()
	db := sqlstorage.NewTestStorage(t)
	configSvc := config.Init(db)

	userSecret := secret.NewText("super secret")

	var svc Service
	var err error

	t.Run("init the service", func(t *testing.T) {
		svc, err = Init(ctx, configSvc, afs, tools)
		require.NoError(t, err)
	})

	t.Run("at first boot no key is loader or registered", func(t *testing.T) {
		require.False(t, svc.IsMasterKeyLoaded())

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.NoError(t, err)
		require.False(t, res)
	})

	t.Run("register a master key", func(t *testing.T) {
		require.False(t, svc.IsMasterKeyLoaded())

		err := svc.GenerateMasterKey(ctx, &userSecret)
		require.NoError(t, err)
	})

	t.Run("after registration the master key is registered and available", func(t *testing.T) {
		require.True(t, svc.IsMasterKeyLoaded())

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.NoError(t, err)
		require.True(t, res)
	})

	t.Run("you can use SealKey / Open", func(t *testing.T) {
		someKey, err := secret.NewKey()
		require.NoError(t, err)

		// Seal
		sealedKey, err := svc.SealKey(someKey)
		require.NoError(t, err)
		require.NotNil(t, sealedKey)

		// Open
		res, err := svc.Open(sealedKey)
		require.NoError(t, err)
		require.Equal(t, someKey.Base64(), res.Base64())
	})

	t.Run("restart the service", func(t *testing.T) {
		svc, err = Init(ctx, configSvc, afs, tools)
		require.NoError(t, err)
	})

	t.Run("after restart the master key is registered but not loaded", func(t *testing.T) {
		require.False(t, svc.IsMasterKeyLoaded())

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.NoError(t, err)
		require.True(t, res)
	})
}

func Test_Integration_masterKey_with_systemd_creds(t *testing.T) {
	tools := tools.NewToolboxForTest(t)
	ctx := context.Background()
	afs := afero.NewMemMapFs()
	db := sqlstorage.NewTestStorage(t)
	configSvc := config.Init(db)

	userSecret := secret.NewText("super secret")

	// Setup the file systeme to emulate the files and env variables setup by
	// systemd at startup.
	credsDir := "/tmp/test/creds"
	err := afs.MkdirAll(credsDir, 0o755)
	require.NoError(t, err)
	err = afero.WriteFile(afs, path.Join(credsDir, "password"), []byte(userSecret.Raw()), 0o644)
	require.NoError(t, err)

	var svc Service

	t.Run("init the service", func(t *testing.T) {
		t.Setenv("CREDENTIALS_DIRECTORY", credsDir)

		svc, err = Init(ctx, configSvc, afs, tools)
		require.NoError(t, err)
	})

	t.Run("at first boot the key is automatically registered and loaded", func(t *testing.T) {
		require.True(t, svc.IsMasterKeyLoaded())

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.NoError(t, err)
		require.True(t, res)
	})

	t.Run("register is not possible as the key is already exists", func(t *testing.T) {
		require.True(t, svc.IsMasterKeyLoaded())

		err := svc.GenerateMasterKey(ctx, &userSecret)
		require.ErrorIs(t, err, ErrAlreadyExists)
	})

	t.Run("you can use SealKey / Open", func(t *testing.T) {
		someKey, err := secret.NewKey()
		require.NoError(t, err)

		// Seal
		sealedKey, err := svc.SealKey(someKey)
		require.NoError(t, err)
		require.NotNil(t, sealedKey)

		// Open
		res, err := svc.Open(sealedKey)
		require.NoError(t, err)
		require.Equal(t, someKey.Base64(), res.Base64())
	})

	t.Run("restart the service", func(t *testing.T) {
		// Systemd will recreate the file and setup the env variables.
		err = afero.WriteFile(afs, path.Join(credsDir, "password"), []byte(userSecret.Raw()), 0o644)
		require.NoError(t, err)
		t.Setenv("CREDENTIALS_DIRECTORY", credsDir)

		svc, err = Init(ctx, configSvc, afs, tools)
		require.NoError(t, err)
	})

	t.Run("the key is still registered and available", func(t *testing.T) {
		require.True(t, svc.IsMasterKeyLoaded())

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.NoError(t, err)
		require.True(t, res)
	})
}
