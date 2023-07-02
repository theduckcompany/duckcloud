package oauthclients

import (
	"context"
	"fmt"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

const (
	WebAppClientID     = "neurone-web-ui"
	WebAppClientSecret = "6e1302fb-c68c-4240-8ad1-b84b26e96723"
)

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

func (t *OauthClientService) BootstrapWebApp(ctx context.Context) error {
	res, err := t.GetByID(ctx, WebAppClientID)
	if err != nil {
		return fmt.Errorf("failed to get by id: %w", err)
	}

	if res != nil {
		// Already setup
		return nil
	}

	err = t.storage.Save(ctx, &Client{
		ID:             WebAppClientID,
		Secret:         WebAppClientSecret,
		RedirectURI:    "http://localhost:8080/oauth_callback",
		UserID:         nil,
		CreatedAt:      t.clock.Now(),
		Scopes:         Scopes{"app"},
		IsPublic:       true,
		SkipValidation: true,
	})
	if err != nil {
		return fmt.Errorf("failed to save the client: %w", err)
	}

	return nil
}

func (t *OauthClientService) GetByID(ctx context.Context, clientID string) (*Client, error) {
	client, err := t.storage.GetByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get by ID: %w", err)
	}

	return client, nil
}
