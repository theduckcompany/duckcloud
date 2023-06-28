package jwt

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Peltoche/neurone/src/tools/response"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	ierr := fmt.Errorf("some stuff: %w", &Error{ErrInvalidAccessToken})

	var rerr *response.Error
	assert.True(t, errors.As(ierr, &rerr))
	assert.Equal(t, http.StatusUnauthorized, rerr.Code)
	assert.Equal(t, "invalid access token", rerr.Message)
}
