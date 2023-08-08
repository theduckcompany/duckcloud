package websessions

import (
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
)

var (
	now               = time.Now()
	WebSessionExample = Session{
		token:     "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
		userID:    uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
		ip:        "192.168.1.1",
		clientID:  "some-client-id",
		device:    "Android - Chrome",
		createdAt: now,
	}
)
