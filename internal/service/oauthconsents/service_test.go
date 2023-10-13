package oauthconsents

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_OauthConsents_Service(t *testing.T) {
	ctx := context.Background()

	t.Run("Create success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("01ce56b3-5ab9-4265-b1d2-e0347dcd4158")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAliceConsent).Return(nil).Once()

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
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		res, err := service.Create(ctx, &CreateCmd{
			UserID:       uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			SessionToken: "some invalid id",
			ClientID:     "alice-oauth-client",
			Scopes:       []string{"scopeA", "scopeB"},
		})
		assert.Nil(t, res)
		assert.EqualError(t, err, "validation: SessionToken: must be a valid UUID v4.")
	})

	t.Run("Create with a storageMockerror", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		tools.UUIDMock.On("New").Return(uuid.UUID("01ce56b3-5ab9-4265-b1d2-e0347dcd4158")).Once()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Save", mock.Anything, &ExampleAliceConsent).Return(errors.New("some-error")).Once()

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
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		tools.UUIDMock.On("Parse", "84a871a1-e8f1-4041-83b3-530d013737cb").Return(uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb"), nil).Once()
		storageMock.On("GetByID", mock.Anything, uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb")).Return(&ExampleAliceConsent, nil).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.AliceWebSessionExample)
		assert.NoError(t, err)
	})

	t.Run("Check with an invalid consent_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "invalid format")
		req.URL.RawQuery = query.Encode()

		tools.UUIDMock.On("Parse", "invalid format").Return(uuid.UUID(""), errors.New("must be a valid UUID v4")).Once()
		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.AliceWebSessionExample)
		assert.EqualError(t, err, "validation: must be a valid UUID v4")
	})

	t.Run("Check with a storageMockerror", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		tools.UUIDMock.On("Parse", "84a871a1-e8f1-4041-83b3-530d013737cb").Return(uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb"), nil).Once()
		storageMock.On("GetByID", mock.Anything, uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb")).Return(nil, errors.New("some-error")).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.AliceWebSessionExample)
		assert.EqualError(t, err, "fail to fetch the consent from storage: some-error")
	})

	t.Run("Check with the consent not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		tools.UUIDMock.On("Parse", "84a871a1-e8f1-4041-83b3-530d013737cb").Return(uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb"), nil).Once()
		storageMock.On("GetByID", mock.Anything, uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb")).Return(nil, nil).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.AliceWebSessionExample)
		assert.EqualError(t, err, "consent not found")
	})

	t.Run("Check with an invalid client_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

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

		tools.UUIDMock.On("Parse", "84a871a1-e8f1-4041-83b3-530d013737cb").Return(uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb"), nil).Once()
		storageMock.On("GetByID", mock.Anything, uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb")).Return(&consent, nil).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.AliceWebSessionExample)
		assert.EqualError(t, err, "bad request: consent clientID doesn't match with the given client")
	})

	t.Run("Check with an invalid websession_id", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		ExampleAliceConsent := Consent{
			id:     uuid.UUID("some-ExampleAliceConsent-id"),
			userID: uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
			// The sessionID doesn't match the one inside the session
			sessionToken: "some-other-token",
			clientID:     "alice-oauth-client",
			scopes:       []string{"scopeA", "scopeB"},
			createdAt:    now,
		}

		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		query := req.URL.Query()
		query.Add("consent_id", "84a871a1-e8f1-4041-83b3-530d013737cb")
		req.URL.RawQuery = query.Encode()

		tools.UUIDMock.On("Parse", "84a871a1-e8f1-4041-83b3-530d013737cb").Return(uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb"), nil).Once()
		storageMock.On("GetByID", mock.Anything, uuid.UUID("84a871a1-e8f1-4041-83b3-530d013737cb")).Return(&ExampleAliceConsent, nil).Once()

		err := service.Check(req, &oauthclients.ExampleAliceClient, &websessions.AliceWebSessionExample)
		assert.EqualError(t, err, "bad request: consent session token doesn't match with the given session")
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceConsent.UserID(), (*storage.PaginateCmd)(nil)).Return([]Consent{ExampleAliceConsent}, nil).Once()

		res, err := service.GetAll(ctx, ExampleAliceConsent.UserID(), nil)
		assert.NoError(t, err)
		assert.Equal(t, []Consent{ExampleAliceConsent}, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("Delete", mock.Anything, ExampleAliceConsent.ID()).Return(nil).Once()

		service.Delete(ctx, ExampleAliceConsent.ID())
	})

	t.Run("DeleteAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceConsent.UserID(), (*storage.PaginateCmd)(nil)).Return([]Consent{ExampleAliceConsent}, nil).Once()
		storageMock.On("Delete", mock.Anything, ExampleAliceConsent.ID()).Return(nil).Once()

		err := service.DeleteAll(ctx, ExampleAliceConsent.UserID())
		assert.NoError(t, err)
	})

	t.Run("DeleteAll with a GetAll error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceConsent.UserID(), (*storage.PaginateCmd)(nil)).Return(nil, fmt.Errorf("some-error")).Once()

		err := service.DeleteAll(ctx, ExampleAliceConsent.UserID())
		assert.EqualError(t, err, "failed to GetAllForUser: some-error")
	})

	t.Run("DeleteAll with a revoke error stop directly", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(storageMock, tools)

		storageMock.On("GetAllForUser", mock.Anything, ExampleAliceConsent.UserID(), (*storage.PaginateCmd)(nil)).Return([]Consent{ExampleAliceConsent, ExampleAliceConsent}, nil).Once()
		storageMock.On("Delete", mock.Anything, ExampleAliceConsent.ID()).Return(fmt.Errorf("some-error")).Once()
		// Do not call GetByID and DeleteByID a second time

		err := service.DeleteAll(ctx, ExampleAliceConsent.UserID())
		assert.EqualError(t, err, fmt.Sprintf("failed to Delete an oauth consent %q: some-error", ExampleAliceConsent.ID()))
	})
}
