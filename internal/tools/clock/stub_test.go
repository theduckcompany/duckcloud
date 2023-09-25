package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStub(t *testing.T) {
	t.Run("Now return the given date", func(t *testing.T) {
		someDate := time.Now().Add(4 * time.Hour)

		stub := &Stub{Time: someDate}
		now := stub.Now()

		assert.Equal(t, someDate, now)
	})
}
