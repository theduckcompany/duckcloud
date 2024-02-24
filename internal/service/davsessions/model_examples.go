package davsessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var now time.Time = time.Now().UTC()

var ExampleAliceSession = DavSession{
	id:        uuid.UUID("d43afe5b-5c3c-4ba4-a08c-031d701f2aef"),
	name:      "My Computer",
	userID:    uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:  "Alice",
	password:  secret.NewText("736f6d652d70617373776f7264"), // hex-encoding of "some-password"
	spaceID:   spaces.ExampleAlicePersonalSpace.ID(),
	createdAt: now,
}

var ExampleAliceSession2 = DavSession{
	id:        uuid.UUID("0c2f3980-3ee4-42dc-8c9e-17249a99203d"),
	name:      "My Computer",
	userID:    uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:  "Alice",
	password:  secret.NewText("736f6d652d70617373776f7264"), // hex-encoding of "some-password"
	spaceID:   spaces.ExampleAlicePersonalSpace.ID(),
	createdAt: now,
}
