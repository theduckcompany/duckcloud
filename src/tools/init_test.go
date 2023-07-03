package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolsImplementations(t *testing.T) {
	assert.Implements(t, (*Tools)(nil), new(Prod))
	assert.Implements(t, (*Tools)(nil), new(Mock))
}
