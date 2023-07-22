package websessions

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
