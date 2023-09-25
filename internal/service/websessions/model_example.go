package websessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	now                    = time.Now()
	AliceWebSessionExample = Session{
		token:     "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
		userID:    uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
		ip:        "192.168.1.1",
		device:    "Android - Chrome",
		createdAt: now,
	}
)
