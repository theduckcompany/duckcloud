package files

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const tableName = "files"

var errNotFound = errors.New("not found")

var allFields = []string{"id", "size", "mimetype", "checksum", "key", "uploaded_at"}

// sqlStorage use to save/retrieve files metadatas
type sqlStorage struct {
	db sqlstorage.Querier
}

// newSqlStorage instantiates a new Storage based on sql.
func newSqlStorage(db sqlstorage.Querier) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, meta *FileMeta) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(meta.id, meta.size, meta.mimetype, meta.checksum, meta.key, ptr.To(sqlstorage.SQLTime(meta.uploadedAt))).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*FileMeta, error) {
	return s.getByKeys(ctx, sq.Eq{"id": id})
}

func (s *sqlStorage) GetByChecksum(ctx context.Context, checksum string) (*FileMeta, error) {
	return s.getByKeys(ctx, sq.Eq{"checksum": checksum})
}

func (s *sqlStorage) Delete(ctx context.Context, fileID uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": string(fileID)}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) getByKeys(ctx context.Context, wheres ...any) (*FileMeta, error) {
	query := sq.
		Select(allFields...).
		From(tableName)

	for _, where := range wheres {
		query = query.Where(where)
	}

	var res FileMeta
	var sqlUploadedAt sqlstorage.SQLTime

	err := query.
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.size, &res.mimetype, &res.checksum, &res.key, &sqlUploadedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.uploadedAt = sqlUploadedAt.Time()

	return &res, nil
}
