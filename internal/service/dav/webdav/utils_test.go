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
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/startutils"
)

type TestContext struct {
	FoldersSvc     folders.Service
	UsersSvc       users.Service
	DavSessionsSvc davsessions.Service

	FSService dfs.Service
	Scheduler scheduler.Service
	Files     files.Service
	Runner    runner.Service
	FS        dfs.FS

	User   *users.User
	Folder *folders.Folder
}

func buildTestFS(t *testing.T, buildfs []string) *TestContext {
	ctx := context.Background()

	serv := startutils.NewServer(t)

	folder, err := serv.FoldersSvc.GetByID(ctx, serv.User.DefaultFolder())
	require.NoError(t, err, "failed to get the user default folder")

	fs := serv.DFSSvc.GetFolderFS(folder)

	for _, b := range buildfs {
		op := strings.Split(b, " ")
		switch op[0] {
		case "mkdir":
			_, err := fs.CreateDir(ctx, op[1])
			require.NoError(t, err)
		case "touch":
			err := fs.Upload(ctx, op[1], http.NoBody)
			require.NoError(t, err)
		case "write":
			buf := bytes.NewBuffer(nil)
			buf.Write([]byte(op[2]))

			err := fs.Upload(ctx, op[1], buf)
			require.NoError(t, err)
		default:
			t.Fatalf("unknown file operation %q", op[0])
		}

		err = serv.RunnerSvc.Run(ctx)
		require.NoError(t, err)
	}

	return &TestContext{
		FoldersSvc:     serv.FoldersSvc,
		UsersSvc:       serv.UsersSvc,
		DavSessionsSvc: serv.DavSessionsSvc,

		FSService: serv.DFSSvc,
		Scheduler: serv.SchedulerSvc,
		Runner:    serv.RunnerSvc,
		Files:     serv.Files,
		FS:        fs,

		User:   serv.User,
		Folder: folder,
	}
}

// find appends to ss the names of the named file and its children. It is
// analogous to the Unix find command.
//
// The returned strings are not guaranteed to be in any particular order.
func find(ctx context.Context, ss []string, fs dfs.FS, name string) ([]string, error) {
	stat, err := fs.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	ss = append(ss, name)
	if stat.IsDir() {
		children, err := fs.ListDir(ctx, name, nil)
		if err != nil {
			return nil, err
		}
		for _, c := range children {
			ss, err = find(ctx, ss, fs, path.Join(name, c.Name()))
			if err != nil {
				return nil, err
			}
		}
	}
	return ss, nil
}
