package oauthsessions

import (
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
)

func Test_CreateSessionRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateSessionRequest))
}

func Test_CreateSessionRequest_Validate_success(t *testing.T) {
	err := CreateSessionRequest{
		AccessToken:      "some-access-session",
		AccessExpiresAt:  time.Now(),
		RefreshToken:     "some-refresh-session",
		RefreshExpiresAt: time.Now(),
		UserID:           "1b51ce74-2f89-47de-bfb4-ee9e12ca814e",
		Scope:            "some-scope",
	}.Validate()

	assert.NoError(t, err)
}
