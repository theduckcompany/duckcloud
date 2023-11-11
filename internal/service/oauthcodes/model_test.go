package oauthcodes

import (
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

func Test_CreateCodeRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateCmd))
}

func Test_CreateCodeRequest_Validate_success(t *testing.T) {
	err := CreateCmd{
		Code:      secret.NewText("some-code"),
		ExpiresAt: time.Now(),
		UserID:    "1b51ce74-2f89-47de-bfb4-ee9e12ca814e",
		Scope:     "some-scope",
	}.Validate()

	assert.NoError(t, err)
}
