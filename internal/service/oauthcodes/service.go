package oauthcodes

import (
	"context"
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
		return fmt.Errorf("failed to save the code: %w", err)
	}

	return nil
}

// delete the authorization code
func (t *OauthCodeService) RemoveByCode(ctx context.Context, code string) error {
	return t.storage.RemoveByCode(ctx, code)
}

// use the authorization code for code information data
func (t *OauthCodeService) GetByCode(ctx context.Context, code string) (*Code, error) {
	return t.storage.GetByCode(ctx, code)
}
