package uploads

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const tableName = "fs_uploads"

var allFiels = []string{"id", "folder_id", "file_id", "directory", "file_name", "uploaded_at"}

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, upload *Upload) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFiels...).
		Values(upload.id, upload.folderID, upload.fileID, upload.dir, upload.fileName, upload.uploadedAt).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]Upload, error) {
	rows, err := storage.PaginateSelection(sq.
		Select(allFiels...).
		From(tableName), cmd).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) Delete(ctx context.Context, uploadID uuid.UUID) error {
	_, err := sq.Delete(tableName).
		Where(sq.Eq{"id": uploadID}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]Upload, error) {
	uploads := []Upload{}

	for rows.Next() {
		var res Upload

		err := rows.Scan(&res.id, &res.folderID, &res.fileID, &res.dir, &res.fileName, &res.uploadedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		uploads = append(uploads, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return uploads, nil
}
