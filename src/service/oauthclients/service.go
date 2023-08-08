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
	WebAppClientID = "web"
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

func NewService(tools tools.Tools, storage Storage) *OauthClientService {
	return &OauthClientService{storage, tools.Clock(), tools.UUID()}
}

func (s *OauthClientService) Create(ctx context.Context, cmd *CreateCmd) (*Client, error) {
	existingClient, err := s.storage.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get by id: %w", err)
	}

	if existingClient != nil {
		return nil, ErrClientIDTaken
	}

	client := Client{
		id:             cmd.ID,
		name:           cmd.Name,
		secret:         string(s.uuid.New()),
		redirectURI:    cmd.RedirectURI,
		userID:         cmd.UserID,
		createdAt:      s.clock.Now(),
		scopes:         cmd.Scopes,
		public:         cmd.Public,
		skipValidation: cmd.SkipValidation,
	}

	err = s.storage.Save(ctx, &client)
	if err != nil {
		return nil, fmt.Errorf("failed to save the client: %w", err)
	}

	return &client, nil
}

func (s *OauthClientService) GetByID(ctx context.Context, clientID string) (*Client, error) {
	client, err := s.storage.GetByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get by ID: %w", err)
	}

	return client, nil
}
