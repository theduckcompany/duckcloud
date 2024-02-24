package webdav

import (
	"bytes"
	"context"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
)

type TestContext struct {
	SpacesSvc      spaces.Service
	UsersSvc       users.Service
	DavSessionsSvc davsessions.Service

	FSService dfs.Service
	Scheduler scheduler.Service
	Files     files.Service
	Runner    runner.Service
	FS        dfs.Service

	User  *users.User
	Space *spaces.Space
}

func buildTestFS(t *testing.T, buildfs []string) *TestContext {
	ctx := context.Background()

	serv := startutils.NewServer(t)

	spaces, err := serv.SpacesSvc.GetAllUserSpaces(ctx, serv.User.ID(), nil)
	require.NoError(t, err, "failed to get the user default space")
	require.NotEmpty(t, spaces)

	space := &spaces[0]

	fsSvc := serv.DFSSvc

	for _, b := range buildfs {
		op := strings.Split(b, " ")
		switch op[0] {
		case "mkdir":
			_, err := fsSvc.CreateDir(ctx, &dfs.CreateDirCmd{
				Path:      dfs.NewPathCmd(space, op[1]),
				CreatedBy: serv.User,
			})
			require.NoError(t, err)
		case "touch":
			err := fsSvc.Upload(ctx, &dfs.UploadCmd{
				Path:       dfs.NewPathCmd(space, op[1]),
				Content:    http.NoBody,
				UploadedBy: serv.User,
			})
			require.NoError(t, err)
		case "write":
			buf := bytes.NewBuffer(nil)
			buf.Write([]byte(op[2]))

			err := fsSvc.Upload(ctx, &dfs.UploadCmd{
				Path:       dfs.NewPathCmd(space, op[1]),
				Content:    buf,
				UploadedBy: serv.User,
			})
			require.NoError(t, err)
		default:
			t.Fatalf("unknown file operation %q", op[0])
		}

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)
	}

	return &TestContext{
		SpacesSvc:      serv.SpacesSvc,
		UsersSvc:       serv.UsersSvc,
		DavSessionsSvc: serv.DavSessionsSvc,

		FSService: serv.DFSSvc,
		Scheduler: serv.SchedulerSvc,
		Runner:    serv.RunnerSvc,
		Files:     serv.Files,
		FS:        fsSvc,

		User:  serv.User,
		Space: &spaces[0],
	}
}

// find appends to ss the names of the named file and its children. It is
// analogous to the Unix find command.
//
// The returned strings are not guaranteed to be in any particular order.
func find(ctx context.Context, ss []string, fs dfs.Service, cmd *dfs.PathCmd) ([]string, error) {
	stat, err := fs.Get(ctx, cmd)
	if err != nil {
		return nil, err
	}
	ss = append(ss, cmd.Path())
	if stat.IsDir() {
		children, err := fs.ListDir(ctx, cmd, nil)
		if err != nil {
			return nil, err
		}
		for _, c := range children {
			ss, err = find(ctx, ss, fs, dfs.NewPathCmd(cmd.Space(), path.Join(cmd.Path(), c.Name())))
			if err != nil {
				return nil, err
			}
		}
	}
	return ss, nil
}
