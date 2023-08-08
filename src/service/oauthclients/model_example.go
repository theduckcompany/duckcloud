package oauthclients

import "time"

var now = time.Now()

var ExampleClient = Client{
	id:             "some-client-id",
	name:           "some-name",
	secret:         "some-secret-uuid",
	redirectURI:    "http://some-url",
	userID:         "some-user-id",
	createdAt:      now,
	scopes:         Scopes{"scopeA", "scopeB"},
	public:         true,
	skipValidation: true,
}
