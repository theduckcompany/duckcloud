package spaces

import (
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type ContentTemplate struct {
	IsAdmin bool
	Spaces  []spaces.Space
	Users   map[uuid.UUID]users.User
}

func (t *ContentTemplate) Template() string { return "settings/spaces/page" }

type CreateSpaceModal struct {
	IsAdmin   bool
	Selection UserSelectionTemplate
}

func (t *CreateSpaceModal) Template() string { return "settings/spaces/modal_create_space" }

type UserSelectionTemplate struct {
	UnselectedUsers []users.User
	SelectedUsers   []users.User
}

func (t *UserSelectionTemplate) Template() string { return "settings/spaces/user_selection" }
