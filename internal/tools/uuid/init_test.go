package uuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceImplementations(t *testing.T) {
	assert.Implements(t, (*Service)(nil), new(Default))
	assert.Implements(t, (*Service)(nil), new(Stub))
	assert.Implements(t, (*Service)(nil), new(MockService))
}
