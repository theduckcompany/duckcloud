package jobs

import (
	"github.com/theduckcompany/duckcloud/src/jobs/fsgc"
	"github.com/theduckcompany/duckcloud/src/jobs/userdelete"
	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/src/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"go.uber.org/fx"
)

func StartJobs(
	lc fx.Lifecycle,
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	inodes inodes.Service,
	tools tools.Tools,
) {
	fsgc.StartJob(lc, inodes, tools)
	userdelete.StartJob(lc, users, webSessions, davSessions, oauthSessions, oauthConsents, inodes, tools)
}
