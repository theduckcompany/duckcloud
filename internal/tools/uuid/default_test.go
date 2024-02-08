package uuid

import (
	"testing"

	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUUID(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		uuidSvc := NewProvider()

		id := uuidSvc.New()
		assert.NotEmpty(t, id)
		require.NoError(t, is.UUIDv4.Validate(id))
	})

	t.Run("parse success", func(t *testing.T) {
		uuidSvc := NewProvider()

		id, err := uuidSvc.Parse("bcb5425c-fa31-4b46-a6d3-1e4a35cacf93")
		require.NoError(t, err)
		assert.Equal(t, UUID("bcb5425c-fa31-4b46-a6d3-1e4a35cacf93"), id)
	})

	t.Run("parse error", func(t *testing.T) {
		uuidSvc := NewProvider()

		id, err := uuidSvc.Parse("some-invalid-id")
		assert.Empty(t, id)
		require.EqualError(t, err, "invalid UUID length: 15")
	})
}
