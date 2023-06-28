package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidationError_match_ErrValidation(t *testing.T) {
	err := ValidationError(fmt.Errorf("some-error"), "super message")

	assert.True(t, errors.Is(err, ErrValidation))
	assert.EqualError(t, err, "validation error: some-error")
}
