package taskstorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
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
	GetLastRegisteredTask(ctx context.Context, name string) (*model.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	Delete(ctx context.Context, taskID uuid.UUID) error
}

type sqlStorage struct {
	db sqlstorage.Querier
}

func NewSqlStorage(db sqlstorage.Querier) *sqlStorage {
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
			ptr.To(sqlstorage.SQLTime(task.RegisteredAt)),
			rawArgs,
		).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetLastRegisteredTask(ctx context.Context, name string) (*model.Task, error) {
	var res model.Task
	var rawArgs json.RawMessage
	var sqlRegisteredAt sqlstorage.SQLTime

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"name": name}).
		OrderBy("registered_at DESC").
		Limit(1).
		RunWith(s.db).
		ScanContext(ctx,
			&res.ID,
			&res.Priority,
			&res.Name,
			&res.Status,
			&res.Retries,
			&sqlRegisteredAt,
			&rawArgs,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.RegisteredAt = sqlRegisteredAt.Time()
	err = json.Unmarshal(rawArgs, &res.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the arg: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	row := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"id": id}).
		OrderBy("priority", "registered_at").
		RunWith(s.db).
		QueryRowContext(ctx)

	return s.scanRow(row)
}

func (s *sqlStorage) GetNext(ctx context.Context) (*model.Task, error) {
	row := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"status": model.Queuing}).
		OrderBy("priority", "registered_at").
		RunWith(s.db).
		QueryRowContext(ctx)

	return s.scanRow(row)
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

func (s *sqlStorage) Delete(ctx context.Context, taskID uuid.UUID) error {
	_, err := sq.Delete(tableName).
		Where(sq.Eq{"id": taskID}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) scanRow(row sq.RowScanner) (*model.Task, error) {
	var res model.Task
	var rawArgs json.RawMessage
	var sqlRegisteredAt sqlstorage.SQLTime

	err := row.Scan(
		&res.ID,
		&res.Priority,
		&res.Name,
		&res.Status,
		&res.Retries,
		&sqlRegisteredAt,
		&rawArgs,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to scan the sql result: %w", err)
	}

	res.RegisteredAt = sqlRegisteredAt.Time()
	err = json.Unmarshal(rawArgs, &res.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the args: %w", err)
	}

	return &res, nil
}
