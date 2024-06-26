package startutils

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/config"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/stats"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tasks"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

type Server struct {
	// Main tools
	Tools *tools.Toolbox
	DB    sqlstorage.Querier
	FS    afero.Fs

	// Services
	ConfigSvc        config.Service
	SpacesSvc        spaces.Service
	SchedulerSvc     scheduler.Service
	DavSessionsSvc   davsessions.Service
	WebSessionsSvc   websessions.Service
	OauthSessionsSvc oauthsessions.Service
	OauthConsentsSvc oauthconsents.Service
	DFSSvc           dfs.Service
	Files            files.Service
	UsersSvc         users.Service
	RunnerSvc        runner.Service
	MasterKeySvc     masterkey.Service
	StatsSvc         stats.Service

	User *users.User
}

func NewServer(t *testing.T) *Server {
	t.Helper()

	masterPassword := secret.NewText("1superStrongPa$$word!")

	ctx := context.Background()

	tools := tools.NewToolboxForTest(t)
	db := sqlstorage.NewTestStorage(t)
	afs := afero.NewMemMapFs()

	configSvc := config.Init(db)
	schedulerSvc := scheduler.Init(db, tools)
	spacesSvc := spaces.Init(tools, db, schedulerSvc)
	webSessionsSvc := websessions.Init(tools, db)
	davSessionsSvc := davsessions.Init(db, spacesSvc, tools)
	oauthSessionsSvc := oauthsessions.Init(tools, db)
	oauthConsentsSvc := oauthconsents.Init(tools, db)
	usersSvc := users.Init(tools, db, schedulerSvc)
	statsSvc := stats.Init(db)

	masterKeySvc, err := masterkey.Init(ctx, configSvc, afs, tools)
	require.NoError(t, err)

	filesInit, err := files.Init(masterKeySvc, "/", afs, tools, db)
	require.NoError(t, err)

	dfsInit, err := dfs.Init(db, spacesSvc, filesInit.Service, schedulerSvc, usersSvc, tools, statsSvc)
	require.NoError(t, err)

	tasks := tasks.Init(dfsInit.Service, spacesSvc, usersSvc, webSessionsSvc, davSessionsSvc, oauthSessionsSvc, oauthConsentsSvc)

	runnerSvc := runner.Init(
		[]runner.TaskRunner{
			dfsInit.FSGCTask,
			dfsInit.FSMoveTask,
			dfsInit.FSRefreshSizeTask,
			dfsInit.FSRemoveDuplicateFilesRunner,
			tasks.UserCreateTask,
			tasks.UserDeleteTask,
			tasks.SpaceCreateTask,
		}, tools, db)

	masterKeySvc.GenerateMasterKey(ctx, &masterPassword)
	require.NoError(t, err)

	err = runnerSvc.Run(ctx)
	require.NoError(t, err)

	user, err := usersSvc.Bootstrap(ctx)
	require.NoError(t, err)

	err = spacesSvc.Bootstrap(ctx, user)
	require.NoError(t, err)

	err = runnerSvc.Run(ctx)
	require.NoError(t, err)

	return &Server{
		Tools: tools,
		DB:    db,
		FS:    afs,

		// Services
		ConfigSvc:        configSvc,
		SpacesSvc:        spacesSvc,
		SchedulerSvc:     schedulerSvc,
		DavSessionsSvc:   davSessionsSvc,
		WebSessionsSvc:   webSessionsSvc,
		OauthSessionsSvc: oauthSessionsSvc,
		OauthConsentsSvc: oauthConsentsSvc,
		MasterKeySvc:     masterKeySvc,
		StatsSvc:         statsSvc,

		Files:     filesInit.Service,
		DFSSvc:    dfsInit.Service,
		UsersSvc:  usersSvc,
		RunnerSvc: runnerSvc,
		User:      user,
	}
}
