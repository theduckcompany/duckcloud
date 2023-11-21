package spaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_CreateCmd_Validate(t *testing.T) {
	assert.EqualError(t, CreateCmd{
		Name:   "My space",
		Owners: []uuid.UUID{"some-invalid-uuid"},
		RootFS: uuid.UUID("49d16286-2a29-44c3-8dc5-3f7e53b49a0b"),
	}.Validate(), "Owners: (0: must be a valid UUID v4.).")
}

func Test_Space_Getters(t *testing.T) {
	assert.Equal(t, ExampleAlicePersonalSpace.ID(), ExampleAlicePersonalSpace.id)
	assert.Equal(t, ExampleAlicePersonalSpace.Name(), ExampleAlicePersonalSpace.name)
	assert.Equal(t, ExampleAlicePersonalSpace.IsPublic(), ExampleAlicePersonalSpace.isPublic)
	assert.Equal(t, ExampleAlicePersonalSpace.Owners(), ExampleAlicePersonalSpace.owners)
	assert.Equal(t, ExampleAlicePersonalSpace.RootFS(), ExampleAlicePersonalSpace.rootFS)
	assert.Equal(t, ExampleAlicePersonalSpace.CreatedAt(), ExampleAlicePersonalSpace.createdAt)
}

func Test_Owners_Getters(t *testing.T) {
	var raw string
	owners := Owners{uuid.UUID("some-id-1"), "some-id-2"}

	t.Run("String", func(t *testing.T) {
		raw = owners.String()
		assert.Equal(t, "some-id-1,some-id-2", raw)
	})

	t.Run("Scan", func(t *testing.T) {
		res := Owners{}
		err := res.Scan(raw)

		assert.NoError(t, err)
		assert.Equal(t, owners, res)
	})

	t.Run("Scan error", func(t *testing.T) {
		res := Owners{}
		err := res.Scan(nil)

		assert.EqualError(t, err, "not a string")
		assert.Empty(t, res)
	})

	t.Run("Value", func(t *testing.T) {
		val, err := owners.Value()
		assert.Equal(t, owners.String(), val)
		assert.NoError(t, err)
	})
}
