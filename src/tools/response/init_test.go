package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriterImplementations(t *testing.T) {
	assert.Implements(t, (*Writer)(nil), new(Default))
	assert.Implements(t, (*Writer)(nil), new(MockWriter))
}
