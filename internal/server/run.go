package server

import (
	"context"
	"fmt"
	"os"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"go.uber.org/fx"
)

func Run(ctx context.Context, cfg Config) (os.Signal, error) {
	// Start server with the HTTP server.
	app := start(ctx, cfg, fx.Invoke(func(*router.API, runner.Service) {}))

	if err := app.Err(); err != nil {
		return nil, err
	}

	startCtx, startCancel := context.WithTimeout(ctx, app.StartTimeout())
	defer startCancel()

	err := app.Start(startCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to start: %w", err)
	}

	var signal os.Signal
	select {
	case shutdown := <-app.Wait():
		signal = shutdown.Signal

	case <-ctx.Done():
		signal = os.Interrupt
	}

	stopCtx, stopCancel := context.WithTimeout(context.WithoutCancel(ctx), app.StopTimeout())
	defer stopCancel()

	err = app.Stop(stopCtx)
	if err != nil {
		return signal, fmt.Errorf("failed to stop: %w", err)
	}

	return signal, nil
}
