package oauth2

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/Peltoche/neurone/src/service/oauthcodes"
	"github.com/Peltoche/neurone/src/service/oauthsessions"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/models"
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
			Code:            code,
			ExpiresAt:       info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()),
			ClientID:        info.GetClientID(),
			UserID:          info.GetUserID(),
			Scope:           info.GetScope(),
			Challenge:       info.GetCodeChallenge(),
			ChallengeMethod: info.GetCodeChallengeMethod().String(),
		})
		if stderrors.Is(err, errs.ErrValidation) {
			return errors.ErrInvalidRequest
		}
		if err != nil {
			return fmt.Errorf("failed to create a code token: %w", err)
		}
		return nil
	}

	err := t.session.Create(ctx, &oauthsessions.CreateCmd{
		AccessToken:      info.GetAccess(),
		AccessExpiresAt:  info.GetAccessCreateAt().Add(info.GetAccessExpiresIn()),
		RefreshToken:     info.GetRefresh(),
		RefreshExpiresAt: info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn()),
		ClientID:         info.GetClientID(),
		UserID:           info.GetUserID(),
		Scope:            info.GetScope(),
	})
	if err != nil {
		return fmt.Errorf("failed to create the access/refresh pair: %w", err)
	}

	return nil
}

// delete the authorization code
func (t *tokenStorage) RemoveByCode(ctx context.Context, code string) error {
	return t.code.RemoveByCode(ctx, code)
}

// use the access token to delete the token information
func (t *tokenStorage) RemoveByAccess(ctx context.Context, access string) error {
	return t.session.RemoveByAccessToken(ctx, access)
}

// use the refresh token to delete the token information
func (t *tokenStorage) RemoveByRefresh(ctx context.Context, refresh string) error {
	return t.session.RemoveByRefreshToken(ctx, refresh)
}

// use the authorization code for token information data
func (t *tokenStorage) GetByCode(ctx context.Context, input string) (oauth2.TokenInfo, error) {
	code, err := t.code.GetByCode(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the token by code: %w", err)
	}

	if code == nil {
		return nil, nil
	}

	return &models.Token{
		Code:                code.Code,
		CodeCreateAt:        code.CreatedAt,
		CodeExpiresIn:       time.Until(code.ExpiresAt),
		ClientID:            code.ClientID,
		UserID:              code.UserID,
		RedirectURI:         code.RedirectURI,
		Scope:               code.Scope,
		CodeChallenge:       code.Challenge,
		CodeChallengeMethod: code.ChallengeMethod,
	}, nil
}

func (t *tokenStorage) GetByAccess(ctx context.Context, input string) (oauth2.TokenInfo, error) {
	session, err := t.session.GetByAccessToken(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the session by access: %w", err)
	}

	if session == nil {
		return nil, nil
	}

	return t.sessionToToken(session), nil
}

// use the refresh token for token information data
func (t *tokenStorage) GetByRefresh(ctx context.Context, input string) (oauth2.TokenInfo, error) {
	session, err := t.session.GetByRefreshToken(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the session by refresh: %w", err)
	}

	if session == nil {
		return nil, nil
	}

	return t.sessionToToken(session), nil
}

func (t *tokenStorage) sessionToToken(session *oauthsessions.Session) *models.Token {
	return &models.Token{
		Access:           session.AccessToken,
		AccessCreateAt:   session.AccessCreatedAt,
		AccessExpiresIn:  time.Until(session.AccessExpiresAt),
		Refresh:          session.RefreshToken,
		RefreshCreateAt:  session.RefreshCreatedAt,
		RefreshExpiresIn: time.Until(session.RefreshExpiresAt),
		ClientID:         session.ClientID,
		UserID:           session.UserID,
		Scope:            session.Scope,
	}
}
