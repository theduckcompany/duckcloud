package fsgc

import (
	"context"
	"time"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"go.uber.org/fx"
)

func StartJob(lc fx.Lifecycle, inodes inodes.Service, tools tools.Tools) {
	gc := NewJob(inodes, tools)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			//nolint:contextcheck // The context given with "OnStart" will be cancelled once all the methods
			// have been called. We need a context running for all the server uptime.
			gc.Start(5 * time.Second)
			return nil
		},
		OnStop: func(context.Context) error {
			gc.Stop()
			return nil
		},
	})
}
