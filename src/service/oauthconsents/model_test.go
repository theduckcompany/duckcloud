package oauthconsents

import (
	"testing"
	"time"

	"github.com/myminicloud/myminicloud/src/tools/uuid"
	"github.com/stretchr/testify/assert"
)

func TestConsentTypes(t *testing.T) {
	now := time.Now()

	consent := Consent{
		id:           uuid.UUID("some-consent-id"),
		userID:       uuid.UUID("some-user-id"),
		sessionToken: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
		clientID:     "some-other-client-id",
		scopes:       []string{"scopeA", "scopeB"},
		createdAt:    now,
	}

	assert.Equal(t, uuid.UUID("some-consent-id"), consent.ID())
	assert.Equal(t, uuid.UUID("some-user-id"), consent.UserID())
	assert.Equal(t, "3a708fc5-dc10-4655-8fc2-33b08a4b33a5", consent.SessionToken())
	assert.Equal(t, "some-other-client-id", consent.ClientID())
	assert.Equal(t, []string{"scopeA", "scopeB"}, consent.Scopes())
	assert.Equal(t, now, consent.CreatedAt())
}
