package oauthclients

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_CreateCmd_Validate_success(t *testing.T) {
	err := CreateCmd{
		ID:             "some-ID",
		Name:           "some-name",
		RedirectURI:    "http://some-url",
		UserID:         uuid.UUID("fe424b54-17ec-4830-bdd8-0e3a49de7179"),
		Scopes:         Scopes{"foo", "bar"},
		Public:         true,
		SkipValidation: true,
	}.Validate()

	assert.NoError(t, err)
}
