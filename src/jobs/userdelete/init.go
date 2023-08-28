package userdelete

import (
	"context"
	"time"

	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"go.uber.org/fx"
)

func StartJob(
	lc fx.Lifecycle,
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	inodes inodes.Service,
	tools tools.Tools,
) {
	gc := NewJob(users, webSessions, davSessions, oauthSessions, oauthConsents, inodes, tools)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			//nolint:contextcheck // The context given with "OnStart" will be cancelled once all the methods
			// have been called. We need a context running for all the server uptime.
			gc.Start(10 * time.Second)
			return nil
		},
		OnStop: func(context.Context) error {
			gc.Stop()
			return nil
		},
	})
}
