package tasks

import (
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"go.uber.org/fx"
)

type Result struct {
	fx.Out
	UserDeleteTask runner.TaskRunner `group:"tasks"`
	UserCreateTask runner.TaskRunner `group:"tasks"`
}

func Init(
	fs dfs.Service,
	spaces spaces.Service,
	users users.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
) Result {
	return Result{
		UserCreateTask: NewUserCreateTaskRunner(users, spaces, fs),
		UserDeleteTask: NewUserDeleteTaskRunner(users, webSessions, davSessions, oauthSessions, oauthConsents, spaces, fs),
	}
}
