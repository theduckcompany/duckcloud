package websessions

import (
	"net/http"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSessionTypes(t *testing.T) {
	now := time.Now()
	session := Session{
		token:     "some-token",
		userID:    uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"),
		ip:        "192.168.1.1",
		clientID:  "some-client-id",
		device:    "Android - Chrome",
		createdAt: now,
	}

	assert.Equal(t, "some-token", session.Token())
	assert.Equal(t, uuid.UUID("3a708fc5-dc10-4655-8fc2-33b08a4b33a5"), session.UserID())
	assert.Equal(t, "192.168.1.1", session.IP())
	assert.Equal(t, "some-client-id", session.ClientID())
	assert.Equal(t, "Android - Chrome", session.Device())
	assert.Equal(t, now, session.CreatedAt())
}

func Test_CreateCmd_Validate(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/foo", nil)

	t.Run("success", func(t *testing.T) {
		cmd := CreateCmd{
			UserID:   "3a708fc5-dc10-4655-8fc2-33b08a4b33a5",
			ClientID: "some-client-id",
			Req:      req,
		}

		assert.NoError(t, cmd.Validate())
	})

	t.Run("with an error", func(t *testing.T) {
		cmd := CreateCmd{
			UserID:   "some-invalid-id",
			ClientID: "some-client-id",
			Req:      req,
		}

		assert.EqualError(t, cmd.Validate(), "UserID: must be a valid UUID v4.")
	})
}
