package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserImplementations(t *testing.T) {
	assert.Implements(t, (*Parser)(nil), new(Default))
	assert.Implements(t, (*Parser)(nil), new(MockParser))
}
