package websessions

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	now               = time.Now()
	WebSessionExample = Session{
		token:     "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
		userID:    uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
		ip:        "192.168.1.1",
		device:    "Android - Chrome",
		createdAt: now,
	}
)
