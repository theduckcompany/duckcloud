package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"go.uber.org/fx"
)

type Job interface {
	Run(ctx context.Context) error
}

type JobRunner struct {
	job           Job
	log           *slog.Logger
	pauseDuration time.Duration
	cancel        context.CancelFunc
	quit          chan struct{}
}

func NewJobRunner(job Job, pause time.Duration, tools tools.Tools) *JobRunner {
	return &JobRunner{job, tools.Logger(), pause, nil, make(chan struct{})}
}

func (j *JobRunner) FXRegister(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			//nolint:contextcheck // The context given with "OnStart" will be cancelled once all the methods
			// have been called. We need a context running for all the server uptime.
			j.Start()
			return nil
		},
		OnStop: func(context.Context) error {
			j.Stop()
			return nil
		},
	})
}

func (j *JobRunner) Start() {
	ticker := time.NewTicker(j.pauseDuration)
	ctx, cancel := context.WithCancel(context.Background())
	j.cancel = cancel

	go func() {
		for {
			select {
			case <-ticker.C:
				err := j.job.Run(ctx)
				if err != nil {
					j.log.Error("fs gc error", slog.String("error", err.Error()))
				}
			case <-j.quit:
				ticker.Stop()
				cancel()
			}
		}
	}()
}

func (j *JobRunner) Stop() {
	close(j.quit)

	if j.cancel != nil {
		j.cancel()
	}
}
