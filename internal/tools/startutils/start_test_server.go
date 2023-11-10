package startutils

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/files"
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
	ConfigSvc        config.Service
	FoldersSvc       folders.Service
	SchedulerSvc     scheduler.Service
	DavSessionsSvc   davsessions.Service
	WebSessionsSvc   websessions.Service
	OauthSessionsSvc oauthsessions.Service
	OauthConsentsSvc oauthconsents.Service
	DFSSvc           dfs.Service
	Files            files.Service
	UsersSvc         users.Service
	RunnerSvc        runner.Service

	User *users.User
}

func NewServer(t *testing.T) *Server {
	t.Helper()

	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)
	afs := afero.NewMemMapFs()

	configSvc := config.Init(db)
	foldersSvc := folders.Init(tools, db)
	schedulerSvc := scheduler.Init(db, tools)
	webSessionsSvc := websessions.Init(tools, db)
	davSessionsSvc := davsessions.Init(db, foldersSvc, tools)
	oauthSessionsSvc := oauthsessions.Init(tools, db)
	oauthConsentsSvc := oauthconsents.Init(tools, db)

	filesInit, err := files.Init("/", afs, tools, db)
	require.NoError(t, err)

	dfsInit, err := dfs.Init(db, foldersSvc, filesInit.Service, schedulerSvc, tools)
	require.NoError(t, err)

	usersInit := users.Init(tools, db, schedulerSvc, foldersSvc, dfsInit.Service, webSessionsSvc,
		davSessionsSvc, oauthSessionsSvc, oauthConsentsSvc)

	runnerSvc := runner.Init(
		[]runner.TaskRunner{
			dfsInit.FSGCTask,
			dfsInit.FSMoveTask,
			dfsInit.FSRefreshSizeTask,
			dfsInit.FSRemoveDuplicateFilesRunner,
			usersInit.UserCreateTask,
			usersInit.UserDeleteTask,
		}, tools, db)

	return &Server{
		Tools: tools,
		DB:    db,
		FS:    afs,

		// Services
		ConfigSvc:        configSvc,
		FoldersSvc:       foldersSvc,
		SchedulerSvc:     schedulerSvc,
		DavSessionsSvc:   davSessionsSvc,
		WebSessionsSvc:   webSessionsSvc,
		OauthSessionsSvc: oauthSessionsSvc,
		OauthConsentsSvc: oauthConsentsSvc,

		Files:     filesInit.Service,
		DFSSvc:    dfsInit.Service,
		UsersSvc:  usersInit.Service,
		RunnerSvc: runnerSvc,
	}
}

func (s *Server) Bootstrap(t *testing.T) {
	ctx := context.Background()

	err := s.ConfigSvc.EnableDevMode(ctx)
	require.NoError(t, err)

	err = s.ConfigSvc.SetAddrs(ctx, []string{httptest.DefaultRemoteAddr, "localhost"}, 6890)
	require.NoError(t, err)

	err = s.ConfigSvc.DisableTLS(ctx)
	require.NoError(t, err)

	user, err := s.UsersSvc.Create(ctx, &users.CreateCmd{
		Username: "admin",
		Password: "my little secret",
		IsAdmin:  true,
	})
	require.NoError(t, err)

	err = s.RunnerSvc.Run(ctx)
	require.NoError(t, err)

	// Fetch again the user in order to have the values changed by
	// the runner jobs.
	user, err = s.UsersSvc.GetByID(ctx, user.ID())
	require.NoError(t, err)

	s.User = user
}
