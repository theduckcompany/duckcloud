package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidationError_match_ErrValidation(t *testing.T) {
	err := BadRequest(fmt.Errorf("some-error"), "super message")

	assert.True(t, errors.Is(err, ErrBadRequest))
	assert.EqualError(t, err, "bad request: some-error")
}
