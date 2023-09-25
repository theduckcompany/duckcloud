package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultClock(t *testing.T) {
	t.Run("Now returns now", func(t *testing.T) {
		now := NewDefault().Now()
		assert.WithinDuration(t, time.Now(), now, 2*time.Millisecond)
	})
}
