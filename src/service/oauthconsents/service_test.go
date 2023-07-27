package oauthconsents

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/websessions"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_Service(t *testing.T) {
	now := time.Now()
	ctx := context.Background()

	consent := Consent{
		ID:           uuid.UUID("some-consent-id"),
		UserID:       uuid.UUID("some-user-id"),
		SessionToken: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
		ClientID:     "some-client-id",
		Scopes:       []string{"scopeA", "scopeB"},
		CreatedAt:    now,
	}

	oauthClient := oauthclients.Client{
		ID:             "some-client-id",
		Name:           "some-name",
		Secret:         "some-secret-uuid",
		RedirectURI:    "http://some-url",
		UserID:         "some-user-id",
		CreatedAt:      now,
		Scopes:         oauthclients.Scopes{"scopeA", "scopeB"},
		Public:         true,
		SkipValidation: true,
	}

	webSession := websessions.Session{
		Token:     "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
		UserID:    uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
		IP:        "192.168.1.1",
		ClientID:  "some-client-id",
		Device:    "Android - Chrome",
		CreatedAt: now,
	}

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-consent-id")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storage.On("Save", mock.Anything, &consent).Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID:       uuid.UUID("some-user-id"),
			SessionToken: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			ClientID:     "some-client-id",
			Scopes:       []string{"scopeA", "scopeB"},
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &consent, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		res, err := service.Create(ctx, &CreateCmd{
			UserID:       uuid.UUID("some-user-id"),
			SessionToken: "some invalid id",
			ClientID:     "some-client-id",
			Scopes:       []string{"scopeA", "scopeB"},
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: SessionToken: must be a valid UUID v4.")
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("some-consent-id")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storage.On("Save", mock.Anything, &consent).Return(errors.New("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID:       uuid.UUID("some-user-id"),
			SessionToken: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			ClientID:     "some-client-id",
			Scopes:       []string{"scopeA", "scopeB"},
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "failed to save the consent: some-error")
	})

	t.Run("Check success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(&consent, nil).Once()

		err := service.Check(req, &oauthClient, &webSession)
		assert.NoError(t, err)
	})

	t.Run("Check with an invalid consent_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "invalid format")
		req.URL.RawQuery = query.Encode()

		err := service.Check(req, &oauthClient, &webSession)
		assert.EqualError(t, err, "validation error: must be a valid UUID v4")
	})

	t.Run("Check with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(nil, errors.New("some-error")).Once()

		err := service.Check(req, &oauthClient, &webSession)
		assert.EqualError(t, err, "fail to fetch the consent from storage: some-error")
	})

	t.Run("Check with the consent not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(nil, nil).Once()

		err := service.Check(req, &oauthClient, &webSession)
		assert.EqualError(t, err, "consent not found")
	})

	t.Run("Check with an invalid client_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		oauthClient := oauthclients.Client{
			// The oauthClient ID doesn't match the one inside the consent
			ID:             "some-othen-client-id",
			Name:           "some-name",
			Secret:         "some-secret-uuid",
			RedirectURI:    "http://some-url",
			UserID:         "some-user-id",
			CreatedAt:      now,
			Scopes:         oauthclients.Scopes{"scopeA", "scopeB"},
			Public:         true,
			SkipValidation: true,
		}

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(&consent, nil).Once()

		err := service.Check(req, &oauthClient, &webSession)
		assert.EqualError(t, err, "bad request: consent clientID doesn't match with the given client")
	})

	t.Run("Check with an invalid websession_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		webSession := websessions.Session{
			// The websession token doesn't match the one inside the consent
			Token:     "some-other-token",
			UserID:    uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
			IP:        "192.168.1.1",
			ClientID:  "some-client-id",
			Device:    "Android - Chrome",
			CreatedAt: now,
		}

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(&consent, nil).Once()

		err := service.Check(req, &oauthClient, &webSession)
		assert.EqualError(t, err, "bad request: consent session token doesn't match with the given session")
	})
}
