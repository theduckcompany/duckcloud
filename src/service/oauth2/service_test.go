package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImplements(t *testing.T) {
	assert.Implements(t, (*Service)(nil), new(Oauth2Service))
}
