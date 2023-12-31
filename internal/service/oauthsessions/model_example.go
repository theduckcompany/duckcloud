package oauthsessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var nowData = time.Now().UTC()

var ExampleAliceSession = Session{
	accessToken:      secret.NewText("some-access-token"),
	accessCreatedAt:  nowData,
	accessExpiresAt:  nowData.Add(time.Hour),
	refreshToken:     secret.NewText("some-refresh-token"),
	refreshCreatedAt: nowData,
	refreshExpiresAt: nowData.Add(10 * time.Hour),
	clientID:         "some-client-id",
	userID:           uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	scope:            "some-scope",
}
