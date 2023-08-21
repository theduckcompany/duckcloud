package oauthconsents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestOauthConsnet_Types(t *testing.T) {
	assert.Equal(t, uuid.UUID("01ce56b3-5ab9-4265-b1d2-e0347dcd4158"), ExampleAliceConsent.ID())
	assert.Equal(t, uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"), ExampleAliceConsent.UserID())
	assert.Equal(t, "3a708fc5-dc10-4655-8fc2-33b08a4b33a5", ExampleAliceConsent.SessionToken())
	assert.Equal(t, "alice-oauth-client", ExampleAliceConsent.ClientID())
	assert.Equal(t, []string{"scopeA", "scopeB"}, ExampleAliceConsent.Scopes())
	assert.Equal(t, now, ExampleAliceConsent.CreatedAt())
}
