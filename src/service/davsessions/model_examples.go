package davsessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var now time.Time = time.Now()

var ExampleAliceSession = DavSession{
	id:        uuid.UUID("d43afe5b-5c3c-4ba4-a08c-031d701f2aef"),
	userID:    users.ExampleAlice.ID(),
	username:  users.ExampleAlice.Username(),
	password:  "f0ce9d6e7315534d2f3603d11f496dafcda25f2f5bc2b4f8292a8ee34fe7735b", // sha256 of "some-password"
	fsRoot:    inodes.ExampleAliceRoot.ID(),
	createdAt: now,
}
