package oauthconsents

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var now = time.Now()

var ExampleAliceConsent = Consent{
	id:           uuid.UUID("01ce56b3-5ab9-4265-b1d2-e0347dcd4158"),
	userID:       uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	sessionToken: "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
	clientID:     "alice-oauth-client",
	scopes:       []string{"scopeA", "scopeB"},
	createdAt:    now,
}
