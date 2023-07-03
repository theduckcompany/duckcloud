package oauthcodes

import (
	"context"
	"fmt"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/errs"
)

var ErrInvalidExpirationDate = fmt.Errorf("invalid expiration date")

type (
	Storage interface {
		Save(ctx context.Context, code *Code) error
		RemoveByCode(ctx context.Context, code string) error
		GetByCode(ctx context.Context, code string) (*Code, error)
	}

	// OauthCodeService handling all the logic.
	OauthCodeService struct {
		storage Storage
		clock   clock.Clock
	}
)

// NewService create a new code service.
func NewService(tools tools.Tools, storage Storage) *OauthCodeService {
	return &OauthCodeService{storage, tools.Clock()}
}

// create and store the new code information
func (t *OauthCodeService) CreateCode(ctx context.Context, input *CreateCodeRequest) error {
	err := input.Validate()
	if err != nil {
		return errs.ValidationError(err)
	}

	now := time.Now()

	if input.ExpiresAt.Before(now) {
		return errs.BadRequest(ErrInvalidExpirationDate, "invalid expiration date")
	}

	err = t.storage.Save(ctx, &Code{
		Code:            input.Code,
		CreatedAt:       now,
		ExpiresAt:       input.ExpiresAt,
		ClientID:        input.ClientID,
		UserID:          input.UserID,
		RedirectURI:     input.RedirectURI,
		Scope:           input.Scope,
		Challenge:       input.Challenge,
		ChallengeMethod: input.ChallengeMethod,
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
