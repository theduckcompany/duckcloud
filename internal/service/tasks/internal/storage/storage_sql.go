package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const tableName = "tasks"

var ErrNotFound = errors.New("not found")

var allFields = []string{"id", "priority", "name", "status", "retries", "registered_at", "args"}

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, task *model.Task) error
	GetNext(ctx context.Context) (*model.Task, error)
	Patch(ctx context.Context, taskID uuid.UUID, fields map[string]any) error
}

type sqlStorage struct {
	db *sql.DB
}

func NewSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, task *model.Task) error {
	rawArgs, err := json.Marshal(task.Args)
	if err != nil {
		return fmt.Errorf("failed to marshal the args: %w", err)
	}

	_, err = sq.
		Insert(tableName).
		Columns(allFields...).
		Values(
			task.ID,
			task.Priority,
			task.Name,
			task.Status,
			task.Retries,
			task.RegisteredAt,
			rawArgs,
		).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetNext(ctx context.Context) (*model.Task, error) {
	res := model.Task{}

	var rawArgs json.RawMessage

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"status": model.Queuing}).
		OrderBy("priority", "registered_at").
		RunWith(s.db).
		ScanContext(ctx,
			&res.ID,
			&res.Priority,
			&res.Name,
			&res.Status,
			&res.Retries,
			&res.RegisteredAt,
			&rawArgs,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	json.Unmarshal(rawArgs, &res.Args)

	return &res, nil
}

func (s *sqlStorage) Patch(ctx context.Context, taskID uuid.UUID, fields map[string]any) error {
	_, err := sq.Update(tableName).
		SetMap(fields).
		Where(sq.Eq{"id": taskID}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}