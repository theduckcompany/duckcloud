package clock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClockImplementations(t *testing.T) {
	assert.Implements(t, (*Clock)(nil), new(Default))
	assert.Implements(t, (*Clock)(nil), new(Stub))
	assert.Implements(t, (*Clock)(nil), new(MockClock))
}
