package jobs

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/jobs/fsgc"
	"github.com/theduckcompany/duckcloud/src/jobs/usercreate"
	"github.com/theduckcompany/duckcloud/src/jobs/userdelete"
	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/files"
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
	files files.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
	inodes inodes.Service,
	tools tools.Tools,
) {
	fsgcJob := fsgc.NewJob(inodes, files, tools)
	fsgcJobRunner := NewJobRunner(fsgcJob, 5*time.Second, tools)
	fsgcJobRunner.FXRegister(lc)

	userCreateJob := usercreate.NewJob(users, inodes, tools)
	userCreateJobRunner := NewJobRunner(userCreateJob, 2*time.Second, tools)
	userCreateJobRunner.FXRegister(lc)

	userDeleteJob := userdelete.NewJob(users, webSessions, davSessions, oauthSessions, oauthConsents, inodes, tools)
	userDeleteJobRunner := NewJobRunner(userDeleteJob, 10*time.Second, tools)
	userDeleteJobRunner.FXRegister(lc)
}
