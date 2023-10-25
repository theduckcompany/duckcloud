package cron

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"go.uber.org/fx"
)

//go:generate mockery --name CronRunner
type CronRunner interface {
	Run(ctx context.Context) error
}

type Cron struct {
	pauseDuration time.Duration
	quit          chan struct{}
	cancel        context.CancelFunc
	log           *slog.Logger
	job           CronRunner
	lock          *sync.Mutex
}

func New(name string, pauseDuration time.Duration, tools tools.Tools, job CronRunner) *Cron {
	log := tools.Logger().With(slog.String("cron", name))

	return &Cron{
		pauseDuration: pauseDuration,
		quit:          make(chan struct{}),
		cancel:        nil,
		log:           log,
		job:           job,
		lock:          new(sync.Mutex),
	}
}

func (s *Cron) RunLoop() {
	ticker := time.NewTicker(s.pauseDuration)

	s.lock.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.lock.Unlock()

	for {
		select {
		case <-ticker.C:
			err := s.job.Run(ctx)
			if err != nil {
				s.log.Error("fs gc error", slog.String("error", err.Error()))
			}
		case <-s.quit:
			ticker.Stop()
		}
	}
}

func (s *Cron) Stop() {
	close(s.quit)

	// The lock is required to avoid a datarace on `s.cancel`
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *Cron) FXRegister(lc fx.Lifecycle) {
	if lc != nil {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				//nolint:contextcheck // there is no context
				go s.RunLoop()
				return nil
			},
			OnStop: func(context.Context) error {
				s.Stop()
				return nil
			},
		})
	}
}
