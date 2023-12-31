package oauth2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	oautherrors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/theduckcompany/duckcloud/internal/service/oauthcodes"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type tokenStorage struct {
	uuid    uuid.Service
	code    oauthcodes.Service
	session oauthsessions.Service
}

// create and store the new token information
func (t *tokenStorage) Create(ctx context.Context, info oauth2.TokenInfo) error {
	if code := info.GetCode(); code != "" {
		err := t.code.Create(ctx, &oauthcodes.CreateCmd{
			Code:            secret.NewText(code),
			ExpiresAt:       info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()),
			ClientID:        info.GetClientID(),
			UserID:          info.GetUserID(),
			Scope:           info.GetScope(),
			Challenge:       secret.NewText(info.GetCodeChallenge()),
			ChallengeMethod: info.GetCodeChallengeMethod().String(),
		})
		if errors.Is(err, errs.ErrValidation) {
			return oautherrors.ErrInvalidRequest
		}
		if err != nil {
			return fmt.Errorf("failed to create a code token: %w", err)
		}
		return nil
	}

	userID, err := t.uuid.Parse(info.GetUserID())
	if err != nil {
		return fmt.Errorf("invalid userID: %w", err)
	}

	_, err = t.session.Create(ctx, &oauthsessions.CreateCmd{
		AccessToken:      secret.NewText(info.GetAccess()),
		AccessExpiresAt:  info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()),
		RefreshToken:     secret.NewText(info.GetRefresh()),
		RefreshExpiresAt: info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()),
		ClientID:         info.GetClientID(),
		UserID:           userID,
		Scope:            info.GetScope(),
	})
	if err != nil {
		return fmt.Errorf("failed to create the access/refresh pair: %w", err)
	}

	return nil
}

// delete the authorization code
func (t *tokenStorage) RemoveByCode(ctx context.Context, code string) error {
	return t.code.RemoveByCode(ctx, secret.NewText(code))
}

// use the access token to delete the token information
func (t *tokenStorage) RemoveByAccess(ctx context.Context, access string) error {
	return t.session.RemoveByAccessToken(ctx, secret.NewText(access))
}

// use the refresh token to delete the token information
func (t *tokenStorage) RemoveByRefresh(ctx context.Context, refresh string) error {
	return t.session.RemoveByRefreshToken(ctx, secret.NewText(refresh))
}

// use the authorization code for token information data
func (t *tokenStorage) GetByCode(ctx context.Context, input string) (oauth2.TokenInfo, error) {
	code, err := t.code.GetByCode(ctx, secret.NewText(input))
	if errors.Is(err, errs.ErrNotFound) {
		return nil, oautherrors.ErrInvalidAuthorizeCode
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the token by code: %w", err)
	}

	if code == nil {
		return nil, nil
	}

	return &models.Token{
		Code:                code.Code().Raw(),
		CodeCreateAt:        code.CreatedAt(),
		CodeExpiresIn:       time.Until(code.ExpiresAt()),
		ClientID:            code.ClientID(),
		UserID:              code.UserID(),
		RedirectURI:         code.RedirectURI(),
		Scope:               code.Scope(),
		CodeChallenge:       code.Challenge().Raw(),
		CodeChallengeMethod: code.ChallengeMethod(),
	}, nil
}

func (t *tokenStorage) GetByAccess(ctx context.Context, input string) (oauth2.TokenInfo, error) {
	session, err := t.session.GetByAccessToken(ctx, secret.NewText(input))
	if errors.Is(err, errs.ErrNotFound) {
		return nil, oautherrors.ErrInvalidAccessToken
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the session by access: %w", err)
	}

	return t.sessionToToken(session), nil
}

// use the refresh token for token information data
func (t *tokenStorage) GetByRefresh(ctx context.Context, input string) (oauth2.TokenInfo, error) {
	session, err := t.session.GetByRefreshToken(ctx, secret.NewText(input))
	if errors.Is(err, errs.ErrNotFound) {
		return nil, oautherrors.ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the session by refresh: %w", err)
	}

	return t.sessionToToken(session), nil
}

func (t *tokenStorage) sessionToToken(session *oauthsessions.Session) *models.Token {
	return &models.Token{
		Access:           session.AccessToken().Raw(),
		AccessCreateAt:   session.AccessCreatedAt(),
		AccessExpiresIn:  time.Until(session.AccessExpiresAt()),
		Refresh:          session.RefreshToken().Raw(),
		RefreshCreateAt:  session.RefreshCreatedAt(),
		RefreshExpiresIn: time.Until(session.RefreshExpiresAt()),
		ClientID:         session.ClientID(),
		UserID:           string(session.UserID()),
		Scope:            session.Scope(),
	}
}
