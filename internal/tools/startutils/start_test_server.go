package startutils

import (
	"context"
	"database/sql"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner/fileupload"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner/usercreate"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
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
	FilesSvc       files.Service
	InodesSvc      inodes.Service
	FoldersSvc     folders.Service
	SchedulerSvc   scheduler.Service
	DavSessionsSvc davsessions.Service
	DFSSvc         dfs.Service
	UsersSvc       users.Service
	RunnerSvc      runner.Service

	User *users.User
}

func NewServer(t *testing.T) *Server {
	tools := tools.NewToolbox(tools.Config{Log: logger.Config{}})
	db := storage.NewTestStorage(t)
	afs := afero.NewMemMapFs()

	filesSvc, err := files.NewFSService(afs, "/", tools)
	require.NoError(t, err)

	inodesSvc := inodes.Init(tools, db)
	foldersSvc := folders.Init(tools, db)
	schedulerSvc := scheduler.Init(db, tools)
	dfsSvc := dfs.NewFSService(inodesSvc, filesSvc, foldersSvc, schedulerSvc, tools)
	usersSvc := users.Init(tools, db, schedulerSvc)
	davSessionsSvc := davsessions.Init(db, foldersSvc, tools)

	fileUploadTask := fileupload.NewTaskRunner(foldersSvc, filesSvc, inodesSvc)
	userCreateTask := usercreate.NewTaskRunner(usersSvc, foldersSvc, inodesSvc)

	runnerSvc := runner.Init([]runner.TaskRunner{
		userCreateTask,
		fileUploadTask,
	}, nil, tools, db)

	return &Server{
		Tools: tools,
		DB:    db,
		FS:    afs,

		// Services
		FilesSvc:       filesSvc,
		InodesSvc:      inodesSvc,
		FoldersSvc:     foldersSvc,
		SchedulerSvc:   schedulerSvc,
		DFSSvc:         dfsSvc,
		DavSessionsSvc: davSessionsSvc,
		UsersSvc:       usersSvc,
		RunnerSvc:      runnerSvc,
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

	s.User = user

	err = s.RunnerSvc.RunSingleJob(ctx)
	require.NoError(t, err)
}
