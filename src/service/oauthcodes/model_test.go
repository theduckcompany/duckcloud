package oauthcodes

import (
	"testing"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
)

func Test_CreateCodeRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateCodeRequest))
}

func Test_CreateCodeRequest_Validate_success(t *testing.T) {
	err := CreateCodeRequest{
		Code:      "some-code",
		ExpiresAt: time.Now(),
		UserID:    "1b51ce74-2f89-47de-bfb4-ee9e12ca814e",
		Scope:     "some-scope",
	}.Validate()

	assert.NoError(t, err)
}
