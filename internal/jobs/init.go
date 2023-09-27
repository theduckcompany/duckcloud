package jobs

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/jobs/fsgc"
	"github.com/theduckcompany/duckcloud/internal/jobs/usercreate"
	"github.com/theduckcompany/duckcloud/internal/jobs/userdelete"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
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
	folders folders.Service,
	fs fs.Service,
	inodes inodes.Service,
	tools tools.Tools,
) {
	fsgcJob := fsgc.NewJob(inodes, files, folders, tools)
	fsgcJobRunner := NewJobRunner(fsgcJob, 5*time.Second, tools)
	fsgcJobRunner.FXRegister(lc)

	userCreateJob := usercreate.NewJob(users, folders, tools)
	userCreateJobRunner := NewJobRunner(userCreateJob, 2*time.Second, tools)
	userCreateJobRunner.FXRegister(lc)

	userDeleteJob := userdelete.NewJob(users, webSessions, davSessions, oauthSessions, oauthConsents, folders, fs, tools)
	userDeleteJobRunner := NewJobRunner(userDeleteJob, 10*time.Second, tools)
	userDeleteJobRunner.FXRegister(lc)
}
