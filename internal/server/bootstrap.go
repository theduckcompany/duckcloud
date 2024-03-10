package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

func bootstrap(ctx context.Context, usersSvc users.Service, spacesSvc spaces.Service) error {
	res, err := usersSvc.GetAll(ctx, &storage.PaginateCmd{Limit: 4})
	if err != nil {
		return fmt.Errorf("failed to GetAll users: %w", err)
	}

	var bootstrapUser *users.User
	switch len(res) {
	case 0:
		bootstrapUser, err = usersSvc.Bootstrap(ctx)
		if err != nil {
			return fmt.Errorf("failed to create the first user: %w", err)
		}
	default:
		for _, user := range res {
			u := user
			if user.IsAdmin() {
				bootstrapUser = &u
			}
		}
	}

	if bootstrapUser == nil {
		return errs.Internal(errors.New("no admin found"))
	}

	err = spacesSvc.Bootstrap(ctx, bootstrapUser)
	if err != nil {
		return fmt.Errorf("failed to create the first space: %w", err)
	}

	return nil
}
