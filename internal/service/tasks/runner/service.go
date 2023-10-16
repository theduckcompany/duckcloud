package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/model"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/storage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
)

const (
	defaultTimeduration = 500 * time.Millisecond
	defaultRetryDelay   = 5 * time.Second
	defaultMaxRetries   = 5
)

type TasksRunner struct {
	storage       storage.Storage
	pauseDuration time.Duration
	runners       map[string]TaskRunner
	cancel        context.CancelFunc
	quit          chan struct{}
	log           *slog.Logger
	clock         clock.Clock
}

func NewService(tools tools.Tools, storage storage.Storage, runners []TaskRunner) *TasksRunner {
	runnerMap := make(map[string]TaskRunner, len(runners))

	for _, runner := range runners {
		runnerMap[runner.Name()] = runner
	}

	return &TasksRunner{
		storage:       storage,
		pauseDuration: defaultTimeduration,
		runners:       runnerMap,
		cancel:        nil,
		quit:          make(chan struct{}),
		log:           tools.Logger(),
		clock:         tools.Clock(),
	}
}

func (t *TasksRunner) Start() {
	ticker := time.NewTicker(t.pauseDuration)
	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	go func() {
		for {
			select {
			case <-ticker.C:
				err := t.RunJob(ctx)
				if err != nil {
					t.log.Error("fs gc error", slog.String("error", err.Error()))
				}
			case <-t.quit:
				ticker.Stop()
				cancel()
			}
		}
	}()
}

func (t *TasksRunner) Stop() {
	close(t.quit)

	if t.cancel != nil {
		t.cancel()
	}
}

func (t *TasksRunner) RunJob(ctx context.Context) error {
	for {
		task, err := t.storage.GetNext(ctx)
		if errors.Is(err, storage.ErrNotFound) {
			// All the tasks have been processed
			return nil
		}

		if err != nil {
			return fmt.Errorf("failed to GetNext task: %w", err)
		}

		log := t.log.With("task", task)

		runner, ok := t.runners[task.Name]
		if !ok {
			log.Error(fmt.Sprintf("unhandled task name: %s", task.Name))

			err := t.storage.Patch(ctx, task.ID, map[string]any{"status": model.Failed})
			if err != nil {
				return fmt.Errorf("failed to Patch task: %w", err)
			}

			continue
		}

		var patchErr error
		err = runner.Run(ctx, task.Args)
		switch {
		case err == nil:
			log.DebugContext(ctx, "task succeed")

			patchErr = t.storage.Patch(ctx, task.ID, map[string]any{"status": model.Succeeded})

		case task.Retries < defaultMaxRetries:
			task.Retries++
			log.With(slog.String("error", err.Error())).
				ErrorContext(ctx, fmt.Sprintf("task failed (#%d), retry later", task.Retries))

			patchErr = t.storage.Patch(ctx, task.ID, map[string]any{
				"retries":       task.Retries,
				"registered_at": time.Now().Add(defaultRetryDelay),
			})

		default:
			task.Status = model.Failed
			log.With(slog.String("error", err.Error())).
				ErrorContext(ctx, "task failed, too many retries")

			patchErr = t.storage.Patch(ctx, task.ID, map[string]any{"status": model.Failed})
		}

		if patchErr != nil {
			return fmt.Errorf("failed to Patch the task status: %w", err)
		}
	}
}
