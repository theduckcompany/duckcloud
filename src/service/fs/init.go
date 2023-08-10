package fs

import (
	"context"
	"os"
	"time"

	"github.com/Peltoche/neurone/src/service/fs/internal"
	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools"
	"go.uber.org/fx"
)

type Service interface {
	Mkdir(ctx context.Context, name string, perm os.FileMode) error
	OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (File, error)
	RemoveAll(ctx context.Context, name string) error
	Rename(ctx context.Context, oldName, newName string) error
	Stat(ctx context.Context, name string) (os.FileInfo, error)
}

func StartGC(lc fx.Lifecycle, inodes inodes.Service, tools tools.Tools) {
	gc := internal.NewGCService(inodes, tools)

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
