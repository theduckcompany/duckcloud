package files

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeFileBuilder struct {
	t    *testing.T
	file *FileMeta
}

func NewFakeFile(t *testing.T) *FakeFileBuilder {
	t.Helper()

	// uuidProvider := uuid.NewProvider()
	uploadedAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())
	masterKey, err := secret.NewKey()
	require.NoError(t, err)
	fileKey, err := secret.NewKey()
	require.NoError(t, err)
	key, err := secret.SealKey(masterKey, fileKey)
	require.NoError(t, err)

	content := []byte(gofakeit.Phrase())

	return &FakeFileBuilder{
		t: t,
		file: &FileMeta{
			uploadedAt: uploadedAt,
			id:         uuid.NewProvider().New(),
			key:        key,
			mimetype:   "text/plain; charset=utf-8",
			checksum:   base64.RawStdEncoding.Strict().EncodeToString(sha256.New().Sum(content)),
			size:       uint64(len(content)),
		},
	}
}

func (f *FakeFileBuilder) WithContent(content []byte) *FakeFileBuilder {
	f.file.checksum = base64.RawStdEncoding.Strict().EncodeToString(sha256.New().Sum(content))
	f.file.size = uint64(len(content))

	return f
}

func (f *FakeFileBuilder) Build() *FileMeta {
	return f.file
}

func (f *FakeFileBuilder) BuildAndStore(ctx context.Context, db sqlstorage.Querier) *FileMeta {
	f.t.Helper()

	storage := newSqlStorage(db)

	err := storage.Save(ctx, f.file)
	require.NoError(f.t, err)

	return f.file
}
