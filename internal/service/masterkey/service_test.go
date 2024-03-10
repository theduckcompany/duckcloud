package masterkey

import (
	"context"
	"path"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"golang.org/x/crypto/argon2"
)

func TestMasterKeyService(t *testing.T) {
	ctx := context.Background()

	password := secret.NewText("super secret")

	rawMasterKey, err := secret.NewKey()
	require.NoError(t, err)

	passKey, err := secret.KeyFromRaw(argon2.Key([]byte(password.Raw()), []byte(password.Raw()), 3, 32*1024, 4, 32))
	require.NoError(t, err)

	sealedKey, err := secret.SealKey(passKey, rawMasterKey)
	require.NoError(t, err)

	t.Run("The master key is not loaded by default", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		assert.False(t, svc.IsMasterKeyLoaded())
	})

	t.Run("IsMasterKeyRegistered with a key registered success", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(sealedKey, nil).Once()

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.NoError(t, err)
		assert.True(t, res)
	})

	t.Run("IsMasterKeyRegistered with no key registered success", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrNotFound).Once()

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.NoError(t, err)
		assert.False(t, res)
	})

	t.Run("IsMasterKeyRegistered with a GetMasterKeyError", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrInternal).Once()

		res, err := svc.IsMasterKeyRegistered(ctx)
		require.ErrorIs(t, err, errs.ErrInternal)
		assert.False(t, res)
	})

	t.Run("LoadMasterKeyFromPassword success", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(sealedKey, nil).Once()

		err := svc.LoadMasterKeyFromPassword(ctx, &password)
		require.NoError(t, err)

		assert.True(t, svc.IsMasterKeyLoaded())
	})

	t.Run("LoadMasterKeyFromPassword with a master key already loaded", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		svc.enclave = memguard.NewEnclaveRandom(32)

		err := svc.LoadMasterKeyFromPassword(ctx, &password)
		require.ErrorIs(t, err, ErrKeyAlreadyDeciphered)

		assert.True(t, svc.IsMasterKeyLoaded())
	})

	t.Run("LoadMasterKeyFromPassword with no master key found", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrNotFound).Once()

		err := svc.LoadMasterKeyFromPassword(ctx, &password)
		require.ErrorIs(t, err, errs.ErrBadRequest)
		require.ErrorIs(t, err, ErrMasterKeyNotFound)
		assert.False(t, svc.IsMasterKeyLoaded())
	})

	t.Run("LoadMasterKeyFromPassword with a GetMasterKey error", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrInternal).Once()

		err := svc.LoadMasterKeyFromPassword(ctx, &password)
		require.ErrorIs(t, err, errs.ErrInternal)
		assert.False(t, svc.IsMasterKeyLoaded())
	})

	t.Run("GenerateMasterKey success", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrNotFound).Once()
		configSvcMock.On("SetMasterKey", mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.GenerateMasterKey(ctx, &password)
		require.NoError(t, err)
		assert.True(t, svc.IsMasterKeyLoaded())
	})

	t.Run("GenerateMasterKey with a master key already set", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(sealedKey, nil).Once()

		err := svc.GenerateMasterKey(ctx, &password)
		require.ErrorIs(t, err, ErrAlreadyExists)
	})

	t.Run("GenerateMasterKey with a GetMasterKey error", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrInternal).Once()

		err := svc.GenerateMasterKey(ctx, &password)
		require.ErrorIs(t, err, errs.ErrInternal)
		assert.False(t, svc.IsMasterKeyLoaded())
	})

	t.Run("GenerateMasterKey with a SetMasterKey error", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, fs)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrNotFound).Once()
		configSvcMock.On("SetMasterKey", mock.Anything, mock.Anything).Return(errs.ErrBadRequest).Once()

		err := svc.GenerateMasterKey(ctx, &password)
		require.ErrorIs(t, err, errs.ErrBadRequest)
		assert.False(t, svc.IsMasterKeyLoaded())
	})

	t.Run("loadPasswordFromSystemdCreds success", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		credsDir := "/tmp/test/creds"

		err := afs.MkdirAll(credsDir, 0o755)
		require.NoError(t, err)

		err = afero.WriteFile(afs, path.Join(credsDir, "password"), []byte(password.Raw()), 0o644)
		require.NoError(t, err)

		t.Setenv("CREDENTIALS_DIRECTORY", credsDir)

		res, err := svc.loadPasswordFromSystemdCreds()
		require.NoError(t, err)
		assert.Equal(t, &password, res)
	})

	t.Run("loadPasswordFromSystemdCreds with an env variable not set", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		// Missing: t.Setenv("CREDENTIALS_DIRECTORY", credsDir)

		res, err := svc.loadPasswordFromSystemdCreds()
		require.ErrorIs(t, err, ErrCredsDirNotSet)
		assert.Nil(t, res)
	})

	t.Run("loadPasswordFromSystemdCreds with an non existing directory", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		t.Setenv("CREDENTIALS_DIRECTORY", "/some/unexisting/dir")

		res, err := svc.loadPasswordFromSystemdCreds()
		require.ErrorIs(t, err, errs.ErrInternal)
		require.ErrorContains(t, err, "failed to open")
		assert.Nil(t, res)
	})

	t.Run("loadOrRegisterMasterKeyFromSystemdCreds success with a master key already registered", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		credsDir := "/tmp/test/creds"

		err := afs.MkdirAll(credsDir, 0o755)
		require.NoError(t, err)

		err = afero.WriteFile(afs, path.Join(credsDir, "password"), []byte(password.Raw()), 0o644)
		require.NoError(t, err)

		t.Setenv("CREDENTIALS_DIRECTORY", credsDir)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(sealedKey, nil).Twice()

		err = svc.loadOrRegisterMasterKeyFromSystemdCreds(ctx)
		require.NoError(t, err)
		assert.True(t, svc.IsMasterKeyLoaded())
	})

	t.Run("loadOrRegisterMasterKeyFromSystemdCreds success with no master key registered", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		credsDir := "/tmp/test/creds"

		err := afs.MkdirAll(credsDir, 0o755)
		require.NoError(t, err)

		err = afero.WriteFile(afs, path.Join(credsDir, "password"), []byte(password.Raw()), 0o644)
		require.NoError(t, err)

		t.Setenv("CREDENTIALS_DIRECTORY", credsDir)

		configSvcMock.On("GetMasterKey", mock.Anything).Return(nil, errs.ErrNotFound).Twice()
		configSvcMock.On("SetMasterKey", mock.Anything, mock.Anything).Return(nil).Once()

		err = svc.loadOrRegisterMasterKeyFromSystemdCreds(ctx)
		require.NoError(t, err)
		assert.True(t, svc.IsMasterKeyLoaded())
	})

	t.Run("LoadMasterKeyFromPassword with a systemd-cred related error", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		// systemd-cred related files not set

		err = svc.loadOrRegisterMasterKeyFromSystemdCreds(ctx)
		require.ErrorIs(t, err, ErrCredsDirNotSet)
	})

	t.Run("SealKey / Open  success", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		svc.enclave = memguard.NewEnclave(passKey.Raw())

		someKey, err := secret.NewKey()
		require.NoError(t, err)

		sealedKey, err := svc.SealKey(someKey)
		require.NoError(t, err)

		res, err := svc.Open(sealedKey)
		require.NoError(t, err)

		assert.Equal(t, someKey.Base64(), res.Base64())
	})

	t.Run("SealKey with not master key available", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		// svc.enclave not set
		require.False(t, svc.IsMasterKeyLoaded())

		someKey, err := secret.NewKey()
		require.NoError(t, err)

		sealedKey, err := svc.SealKey(someKey)
		require.ErrorIs(t, err, ErrMasterKeyNotFound)
		assert.Nil(t, sealedKey)
	})

	t.Run("Open with not master key available", func(t *testing.T) {
		afs := afero.NewMemMapFs()
		configSvcMock := config.NewMockService(t)
		svc := newService(configSvcMock, afs)

		// svc.enclave not set
		require.False(t, svc.IsMasterKeyLoaded())

		res, err := svc.Open(nil)
		require.ErrorIs(t, err, ErrMasterKeyNotFound)
		assert.Nil(t, res)
	})
}
