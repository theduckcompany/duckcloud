package oauthsessions

import (
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

func TestSessionGetter(t *testing.T) {
	assert.Equal(t, ExampleAliceSession.AccessToken(), ExampleAliceSession.accessToken)
	assert.Equal(t, ExampleAliceSession.AccessCreatedAt(), ExampleAliceSession.accessCreatedAt)
	assert.Equal(t, ExampleAliceSession.AccessExpiresAt(), ExampleAliceSession.accessExpiresAt)
	assert.Equal(t, ExampleAliceSession.RefreshToken(), ExampleAliceSession.refreshToken)
	assert.Equal(t, ExampleAliceSession.RefreshCreatedAt(), ExampleAliceSession.refreshCreatedAt)
	assert.Equal(t, ExampleAliceSession.RefreshExpiresAt(), ExampleAliceSession.refreshExpiresAt)

	assert.Equal(t, ExampleAliceSession.ClientID(), ExampleAliceSession.clientID)
	assert.Equal(t, ExampleAliceSession.UserID(), ExampleAliceSession.userID)
	assert.Equal(t, ExampleAliceSession.Scope(), ExampleAliceSession.scope)
}

func Test_CreateCmd_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateCmd))
}

func Test_CreateCmd_Validate_success(t *testing.T) {
	err := CreateCmd{
		AccessToken:      secret.NewText("some-access-session"),
		AccessExpiresAt:  time.Now(),
		RefreshToken:     secret.NewText("some-refresh-session"),
		RefreshExpiresAt: time.Now(),
		UserID:           "1b51ce74-2f89-47de-bfb4-ee9e12ca814e",
		Scope:            "some-scope",
	}.Validate()

	require.NoError(t, err)
}
