package websessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	now                    = time.Now()
	AliceWebSessionExample = Session{
		token:     secret.NewText("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
		userID:    uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		ip:        "192.168.1.1",
		device:    "Android - Chrome",
		createdAt: now,
	}

	BobWebSessionExample = Session{
		token:     secret.NewText("b9d8fc98-d71f-4f76-a23a-3411a48ef34e"),
		userID:    uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
		ip:        "192.168.1.1",
		device:    "Android - Chrome",
		createdAt: now,
	}
)
