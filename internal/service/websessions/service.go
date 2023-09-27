package websessions

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	ua "github.com/mileusna/useragent"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var ErrUserIDNotMatching = errors.New("user ids are not matching")

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, session *Session) error
	GetByToken(ctx context.Context, token string) (*Session, error)
	RemoveByToken(ctx context.Context, token string) error
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error)
}

type WebSessionsService struct {
	clock   clock.Clock
	storage Storage
	uuid    uuid.Service
}

func NewService(storage Storage, tools tools.Tools) *WebSessionsService {
	return &WebSessionsService{
		clock:   tools.Clock(),
		uuid:    tools.UUID(),
		storage: storage,
	}
}

func (s *WebSessionsService) Create(ctx context.Context, cmd *CreateCmd) (*Session, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	uaRes := ua.Parse(cmd.Req.Header.Get("User-Agent"))

	session := &Session{
		token:     string(s.uuid.New()),
		userID:    uuid.UUID(cmd.UserID),
		ip:        cmd.Req.RemoteAddr,
		device:    fmt.Sprintf("%s - %s", uaRes.OS, uaRes.Name),
		createdAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to save the session: %w", err)
	}

	return session, nil
}

func (s *WebSessionsService) Delete(ctx context.Context, cmd *DeleteCmd) error {
	err := cmd.Validate()
	if err != nil {
		return errs.ValidationError(err)
	}

	session, err := s.storage.GetByToken(ctx, cmd.Token)
	if err != nil {
		return fmt.Errorf("failed to GetByToken: %w", err)
	}

	if session == nil {
		return nil
	}

	if session.UserID() != cmd.UserID {
		return errs.NotFound(ErrUserIDNotMatching, "not found")
	}

	err = s.storage.RemoveByToken(ctx, session.Token())
	if err != nil {
		return fmt.Errorf("failed to RemoveByToken: %w", err)
	}

	return nil
}

func (s *WebSessionsService) GetByToken(ctx context.Context, token string) (*Session, error) {
	session, err := s.storage.GetByToken(ctx, token)

	// TODO: Handle session expiration

	return session, err
}

func (s *WebSessionsService) GetFromReq(r *http.Request) (*Session, error) {
	c, err := r.Cookie("session_token")
	if errors.Is(err, http.ErrNoCookie) {
		return nil, errs.BadRequest(ErrMissingSessionToken, "invalid_request")
	}

	session, err := s.GetByToken(r.Context(), c.Value)
	if err != nil {
		return nil, errs.Unhandled(err)
	}

	if session == nil {
		return nil, errs.BadRequest(ErrSessionNotFound, "session not found")
	}

	return session, nil
}

func (s *WebSessionsService) Logout(r *http.Request, w http.ResponseWriter) error {
	c, err := r.Cookie("session_token")
	if errors.Is(err, http.ErrNoCookie) {
		// There is not session and so nothing to do.
		return nil
	}

	err = s.storage.RemoveByToken(r.Context(), c.Value)
	if err != nil {
		return fmt.Errorf("failed to remove the token: %w", err)
	}

	// Remove to cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})

	w.Header().Set("Location", "/login")
	w.WriteHeader(http.StatusFound)

	return nil
}

func (s *WebSessionsService) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error) {
	return s.storage.GetAllForUser(ctx, userID, cmd)
}

func (s *WebSessionsService) DeleteAll(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.GetAllForUser(ctx, userID, nil)
	if err != nil {
		return fmt.Errorf("failed to GetAllForUser: %w", err)
	}

	for _, session := range sessions {
		err = s.Delete(ctx, &DeleteCmd{
			UserID: userID,
			Token:  session.Token(),
		})
		if err != nil {
			return fmt.Errorf("failed to Delete web session %q: %w", session.Token(), err)
		}
	}

	return nil
}