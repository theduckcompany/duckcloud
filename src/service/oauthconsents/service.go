package oauthconsents

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var ErrConsentNotFound = errors.New("consent not found")

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, consent *Consent) error
	GetByID(ctx context.Context, id uuid.UUID) (*Consent, error)
	Delete(ctx context.Context, consentID uuid.UUID) error
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Consent, error)
}

type OauthConsentsService struct {
	storage Storage
	uuid    uuid.Service
	clock   clock.Clock
}

func NewService(storage Storage, tools tools.Tools) *OauthConsentsService {
	return &OauthConsentsService{storage, tools.UUID(), tools.Clock()}
}

func (s *OauthConsentsService) Create(ctx context.Context, cmd *CreateCmd) (*Consent, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	consent := Consent{
		id:           s.uuid.New(),
		userID:       cmd.UserID,
		sessionToken: cmd.SessionToken,
		clientID:     cmd.ClientID,
		scopes:       cmd.Scopes,
		createdAt:    s.clock.Now(),
	}

	err = s.storage.Save(ctx, &consent)
	if err != nil {
		return nil, fmt.Errorf("failed to save the consent: %w", err)
	}

	return &consent, nil
}

func (s *OauthConsentsService) Check(r *http.Request, client *oauthclients.Client, session *websessions.Session) error {
	rawConsentID := r.FormValue("consent_id")

	consentID, err := s.uuid.Parse(rawConsentID)
	if err != nil {
		return errs.ValidationError(is.ErrUUIDv4)
	}

	consent, err := s.storage.GetByID(r.Context(), consentID)
	if err != nil {
		return fmt.Errorf("fail to fetch the consent from storage: %w", err)
	}

	if consent == nil {
		return ErrConsentNotFound
	}

	if consent.ClientID() != client.GetID() {
		return errs.BadRequest(errors.New("consent clientID doesn't match with the given client"), "invalid request")
	}

	if consent.SessionToken() != session.Token() {
		return errs.BadRequest(errors.New("consent session token doesn't match with the given session"), "invalid request")
	}

	return nil
}

func (s *OauthConsentsService) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Consent, error) {
	return s.storage.GetAllForUser(ctx, userID, cmd)
}

func (s *OauthConsentsService) Delete(ctx context.Context, consentID uuid.UUID) error {
	return s.storage.Delete(ctx, consentID)
}
