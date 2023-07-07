package oauthclients

import (
	"context"
	"errors"
	"fmt"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

const (
	WebAppClientID = "neurone-web-ui"
)

var (
	ErrClientIDTaken = errors.New("clientID already exists")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, client *Client) error
	GetByID(ctx context.Context, id string) (*Client, error)
}

type OauthClientService struct {
	storage Storage
	clock   clock.Clock
	uuid    uuid.Service
}

func NewService(app tools.Tools, storage Storage) *OauthClientService {
	return &OauthClientService{storage, app.Clock(), app.UUID()}
}

func (s *OauthClientService) Create(ctx context.Context, cmd *CreateCmd) error {
	client, err := s.storage.GetByID(ctx, cmd.ID)
	if err != nil {
		return fmt.Errorf("failed to get by id: %w", err)
	}

	if client != nil {
		return ErrClientIDTaken
	}

	err = s.storage.Save(ctx, &Client{
		ID:             cmd.ID,
		Name:           cmd.Name,
		Secret:         string(s.uuid.New()),
		RedirectURI:    cmd.RedirectURI,
		UserID:         cmd.UserID,
		CreatedAt:      s.clock.Now(),
		Scopes:         cmd.Scopes,
		Public:         cmd.Public,
		SkipValidation: cmd.SkipValidation,
	})
	if err != nil {
		return fmt.Errorf("failed to save the client: %w", err)
	}

	return nil
}

func (s *OauthClientService) GetByID(ctx context.Context, clientID string) (*Client, error) {
	client, err := s.storage.GetByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get by ID: %w", err)
	}

	return client, nil
}
