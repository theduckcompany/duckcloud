package oauthcodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

//go:generate mockery --name storage
type storage interface {
	Save(ctx context.Context, code *Code) error
	RemoveByCode(ctx context.Context, code secret.Text) error
	GetByCode(ctx context.Context, code secret.Text) (*Code, error)
}

// service handling all the logic.
type service struct {
	storage storage
	clock   clock.Clock
}

// newService create a new code service.
func newService(tools tools.Tools, storage storage) *service {
	return &service{storage, tools.Clock()}
}

// create and store the new code information
func (t *service) Create(ctx context.Context, input *CreateCmd) error {
	err := input.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	now := t.clock.Now()

	if input.ExpiresAt.Before(now) {
		return errs.BadRequest(ErrInvalidExpirationDate, "invalid expiration date")
	}

	err = t.storage.Save(ctx, &Code{
		code:            input.Code,
		createdAt:       now,
		expiresAt:       input.ExpiresAt,
		clientID:        input.ClientID,
		userID:          input.UserID,
		redirectURI:     input.RedirectURI,
		scope:           input.Scope,
		challenge:       input.Challenge,
		challengeMethod: input.ChallengeMethod,
	})
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to Save: %w", err))
	}

	return nil
}

// delete the authorization code
func (t *service) RemoveByCode(ctx context.Context, code secret.Text) error {
	err := t.storage.RemoveByCode(ctx, code)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

// use the authorization code for code information data
func (t *service) GetByCode(ctx context.Context, code secret.Text) (*Code, error) {
	res, err := t.storage.GetByCode(ctx, code)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(errNotFound)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}
