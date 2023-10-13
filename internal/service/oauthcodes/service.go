package oauthcodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, code *Code) error
	RemoveByCode(ctx context.Context, code string) error
	GetByCode(ctx context.Context, code string) (*Code, error)
}

// OauthCodeService handling all the logic.
type OauthCodeService struct {
	storage Storage
	clock   clock.Clock
}

// NewService create a new code service.
func NewService(tools tools.Tools, storage Storage) *OauthCodeService {
	return &OauthCodeService{storage, tools.Clock()}
}

// create and store the new code information
func (t *OauthCodeService) Create(ctx context.Context, input *CreateCmd) error {
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
func (t *OauthCodeService) RemoveByCode(ctx context.Context, code string) error {
	err := t.storage.RemoveByCode(ctx, code)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

// use the authorization code for code information data
func (t *OauthCodeService) GetByCode(ctx context.Context, code string) (*Code, error) {
	res, err := t.storage.GetByCode(ctx, code)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(errNotFound)
	}

	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}
