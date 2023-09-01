package davsessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var now time.Time = time.Now().UTC()

var ExampleAliceSession = DavSession{
	id:        uuid.UUID("d43afe5b-5c3c-4ba4-a08c-031d701f2aef"),
	name:      "My Computer",
	userID:    uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:  "Alice",
	password:  "f0ce9d6e7315534d2f3603d11f496dafcda25f2f5bc2b4f8292a8ee34fe7735b", // sha256 of "some-password"
	folders:   Folders{folders.ExampleAlicePersonalFolder.ID()},
	createdAt: now,
}
