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
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/models"
)

type tokenStorage struct {
	uuid    uuid.Service
	code    oauthcodes.Service
	session oauthsessions.Service
}

// create and store the new token information
func (t *tokenStorage) Create(info oauth2.TokenInfo) error {
	if code := info.GetCode(); code != "" {
		err := t.code.CreateCode(context.Background(), &oauthcodes.CreateCodeRequest{
			Code:      code,
			ExpiresAt: info.GetCodeCreateAt().Add(info.GetCodeExpiresIn()),
			ClientID:  info.GetClientID(),
			UserID:    info.GetUserID(),
			Scope:     info.GetScope(),
		})
		if stderrors.Is(err, errs.ErrValidation) {
			return errors.ErrInvalidRequest
		}
		if err != nil {
			return fmt.Errorf("failed to create a code token: %w", err)
		}
		return nil
	}

	err := t.session.CreateSession(context.Background(), &oauthsessions.CreateSessionRequest{
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
func (t *tokenStorage) RemoveByCode(code string) error {
	return t.code.RemoveByCode(context.Background(), code)
}

// use the access token to delete the token information
func (t *tokenStorage) RemoveByAccess(access string) error {
	return t.session.RemoveByAccessToken(context.Background(), access)
}

// use the refresh token to delete the token information
func (t *tokenStorage) RemoveByRefresh(refresh string) error {
	return t.session.RemoveByRefreshToken(context.Background(), refresh)
}

// use the authorization code for token information data
func (t *tokenStorage) GetByCode(input string) (oauth2.TokenInfo, error) {
	code, err := t.code.GetByCode(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the token by code: %w", err)
	}

	if code == nil {
		return nil, nil
	}

	return &models.Token{
		Code:          code.Code,
		CodeCreateAt:  code.CreatedAt,
		CodeExpiresIn: time.Until(code.ExpiresAt),
		ClientID:      code.ClientID,
		UserID:        code.UserID,
		RedirectURI:   code.RedirectURI,
		Scope:         code.Scope,
	}, nil
}

func (t *tokenStorage) GetByAccess(input string) (oauth2.TokenInfo, error) {
	session, err := t.session.GetByAccessToken(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the session by access: %w", err)
	}

	if session == nil {
		return nil, nil
	}

	return t.sessionToToken(session), nil
}

// use the refresh token for token information data
func (t *tokenStorage) GetByRefresh(input string) (oauth2.TokenInfo, error) {
	session, err := t.session.GetByRefreshToken(context.Background(), input)
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
