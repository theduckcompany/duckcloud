package startutils

import (
	"context"
	"database/sql"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type Server struct {
	// Main tools
	Tools *tools.Toolbox
	DB    *sql.DB
	FS    afero.Fs

	// Services
	FoldersSvc       folders.Service
	SchedulerSvc     scheduler.Service
	DavSessionsSvc   davsessions.Service
	WebSessionsSvc   websessions.Service
	OauthSessionsSvc oauthsessions.Service
	OauthConsentsSvc oauthconsents.Service
	DFSSvc           dfs.Service
	UsersSvc         users.Service
	RunnerSvc        runner.Service

	User *users.User
}

func NewServer(t *testing.T) *Server {
	t.Helper()

	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)
	afs := afero.NewMemMapFs()

	foldersSvc := folders.Init(tools, db)
	schedulerSvc := scheduler.Init(db, tools)
	webSessionsSvc := websessions.Init(tools, db)
	davSessionsSvc := davsessions.Init(db, foldersSvc, tools)
	oauthSessionsSvc := oauthsessions.Init(tools, db)
	oauthConsentsSvc := oauthconsents.Init(tools, db)

	dfsInit, err := dfs.Init(dfs.Config{Path: "/"}, afs, db, foldersSvc, schedulerSvc, tools)
	require.NoError(t, err)

	usersInit := users.Init(tools, db, schedulerSvc, foldersSvc, dfsInit.Service, webSessionsSvc,
		davSessionsSvc, oauthSessionsSvc, oauthConsentsSvc)

	runnerSvc := runner.Init(
		[]runner.TaskRunner{
			dfsInit.FSGCTask,
			dfsInit.FSMoveTask,
			dfsInit.FileUploadTask,
			usersInit.UserCreateTask,
			usersInit.UserDeleteTask,
		}, nil, tools, db)

	return &Server{
		Tools: tools,
		DB:    db,
		FS:    afs,

		// Services
		FoldersSvc:       foldersSvc,
		SchedulerSvc:     schedulerSvc,
		DavSessionsSvc:   davSessionsSvc,
		WebSessionsSvc:   webSessionsSvc,
		OauthSessionsSvc: oauthSessionsSvc,
		OauthConsentsSvc: oauthConsentsSvc,

		DFSSvc:    dfsInit.Service,
		UsersSvc:  usersInit.Service,
		RunnerSvc: runnerSvc,
	}
}

func (s *Server) Bootstrap(t *testing.T) {
	ctx := context.Background()

	user, err := s.UsersSvc.Create(ctx, &users.CreateCmd{
		Username: "admin",
		Password: "my little secret",
		IsAdmin:  true,
	})
	require.NoError(t, err)

	err = s.RunnerSvc.RunSingleJob(ctx)
	require.NoError(t, err)

	// Fetch again the user in order to have the values changed by
	// the runner jobs.
	user, err = s.UsersSvc.GetByID(ctx, user.ID())
	require.NoError(t, err)

	s.User = user
}
