package model

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type Status string

const (
	Queuing Status = "queuing"
	Failed  Status = "failed"
)

type Task struct {
	ID           uuid.UUID
	Priority     int
	Name         string
	Status       Status
	Retries      int
	RegisteredAt time.Time
	Args         json.RawMessage
}

func (t *Task) LogValue() slog.Value {
	if t == nil {
		return slog.AnyValue(nil)
	}

	return slog.GroupValue(
		slog.String("id", string(t.ID)),
		slog.String("name", t.Name),
		slog.Int("priority", t.Priority),
		slog.String("status", string(t.Status)),
		slog.Int("retries", t.Retries),
		slog.Time("registered_at", t.RegisteredAt),
		slog.String("arguments", string(t.Args)),
	)
}
