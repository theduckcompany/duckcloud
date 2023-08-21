package oauthclients

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var now = time.Now()

var ExampleAliceClient = Client{
	id:             "alice-oauth-client",
	name:           "some-name",
	secret:         "some-secret-uuid",
	redirectURI:    "http://some-url",
	userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	createdAt:      now,
	scopes:         Scopes{"scopeA", "scopeB"},
	public:         true,
	skipValidation: true,
}
