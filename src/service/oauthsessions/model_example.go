package oauthsessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var nowData = time.Now().UTC()

var ExampleAliceSession = Session{
	accessToken:      "some-access-token",
	accessCreatedAt:  nowData,
	accessExpiresAt:  nowData.Add(time.Hour),
	refreshToken:     "some-refresh-token",
	refreshCreatedAt: nowData,
	refreshExpiresAt: nowData.Add(10 * time.Hour),
	clientID:         "some-client-id",
	userID:           uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	scope:            "some-scope",
}
