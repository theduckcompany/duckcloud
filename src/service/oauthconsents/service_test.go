package oauthconsents

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/src/service/oauthclients"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func Test_OauthConsents_Service(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("01ce56b3-5ab9-4265-b1d2-e0347dcd4158")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storage.On("Save", mock.Anything, &ExampleAliceConsent).Return(nil).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID:       uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			SessionToken: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			ClientID:     "alice-oauth-client",
			Scopes:       []string{"scopeA", "scopeB"},
		})
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceConsent, res)
	})

	t.Run("Create with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		res, err := service.Create(ctx, &CreateCmd{
			UserID:       uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			SessionToken: "some invalid id",
			ClientID:     "alice-oauth-client",
			Scopes:       []string{"scopeA", "scopeB"},
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: SessionToken: must be a valid UUID v4.")
	})

	t.Run("Create with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("01ce56b3-5ab9-4265-b1d2-e0347dcd4158")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storage.On("Save", mock.Anything, &ExampleAliceConsent).Return(errors.New("some-error")).Once()

		res, err := service.Create(ctx, &CreateCmd{
			UserID:       uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			SessionToken: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			ClientID:     "alice-oauth-client",
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

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(&ExampleAliceConsent, nil).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.WebSessionExample)
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

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.WebSessionExample)
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

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.WebSessionExample)
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

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.WebSessionExample)
		assert.EqualError(t, err, "consent not found")
	})

	t.Run("Check with an invalid client_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		consent := Consent{
			id:           ExampleAliceConsent.id,
			userID:       ExampleAliceConsent.userID,
			sessionToken: ExampleAliceConsent.sessionToken,
			clientID:     "some-other-client-id", // invalid client id
			scopes:       ExampleAliceConsent.scopes,
			createdAt:    now,
		}

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(&consent, nil).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.WebSessionExample)
		assert.EqualError(t, err, "bad request: consent clientID doesn't match with the given client")
	})

	t.Run("Check with an invalid websession_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storage := NewMockStorage(t)
		service := NewService(storage, tools)

		ExampleAliceConsent := Consent{
			id:     uuid.UUID("some-ExampleAliceConsent-id"),
			userID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			// The sessionToken doesn't match the one inside the session
			sessionToken: "some-other-token",
			clientID:     "alice-oauth-client",
			scopes:       []string{"scopeA", "scopeB"},
			createdAt:    now,
		}

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		storage.On("GetByID", mock.Anything, "84a871a1-e8f1-4041-83b3-530d013737cb").Return(&ExampleAliceConsent, nil).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.WebSessionExample)
		assert.EqualError(t, err, "bad request: consent session token doesn't match with the given session")
	})
}
