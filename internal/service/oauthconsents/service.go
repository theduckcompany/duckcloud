package oauthconsents

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
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
		return nil, errs.Validation(err)
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
		return nil, errs.Internal(fmt.Errorf("failed to Save: %w", err))
	}

	return &consent, nil
}

func (s *OauthConsentsService) Check(r *http.Request, client *oauthclients.Client, session *websessions.Session) error {
	rawConsentID := r.FormValue("consent_id")

	consentID, err := s.uuid.Parse(rawConsentID)
	if err != nil {
		return errs.Validation(is.ErrUUIDv4)
	}

	consent, err := s.storage.GetByID(r.Context(), consentID)
	if errors.Is(err, errNotFound) {
		return errs.NotFound(ErrConsentNotFound)
	}
	if err != nil {
		return errs.Internal(fmt.Errorf("fail to GetByID: %w", err))
	}

	if consent.ClientID() != client.GetID() {
		return errs.BadRequest(errors.New("consent clientID doesn't match with the given client"), "invalid request")
	}

	if consent.SessionToken() != session.Token().Raw() {
		return errs.BadRequest(errors.New("consent session token doesn't match with the given session"), "invalid request")
	}

	return nil
}

func (s *OauthConsentsService) GetAll(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Consent, error) {
	res, err := s.storage.GetAllForUser(ctx, userID, cmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *OauthConsentsService) Delete(ctx context.Context, consentID uuid.UUID) error {
	err := s.storage.Delete(ctx, consentID)
	if err != nil {
		return errs.Internal(err)
	}

	return nil
}

func (s *OauthConsentsService) DeleteAll(ctx context.Context, userID uuid.UUID) error {
	consents, err := s.GetAll(ctx, userID, nil)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to GetAllForUser: %w", err))
	}

	for _, consent := range consents {
		err = s.Delete(ctx, consent.ID())
		if err != nil {
			return errs.Internal(fmt.Errorf("failed to Delete an oauth consent %q: %w", consent.ID(), err))
		}
	}

	return nil
}
