package spaces

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_Space_Getters(t *testing.T) {
	assert.Equal(t, ExampleAlicePersonalSpace.ID(), ExampleAlicePersonalSpace.id)
	assert.Equal(t, ExampleAlicePersonalSpace.Name(), ExampleAlicePersonalSpace.name)
	assert.Equal(t, ExampleAlicePersonalSpace.Owners(), ExampleAlicePersonalSpace.owners)
	assert.Equal(t, ExampleAlicePersonalSpace.CreatedAt(), ExampleAlicePersonalSpace.createdAt)
	assert.Equal(t, ExampleAlicePersonalSpace.CreatedBy(), ExampleAlicePersonalSpace.createdBy)
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

		require.NoError(t, err)
		assert.Equal(t, owners, res)
	})

	t.Run("Scan error", func(t *testing.T) {
		res := Owners{}
		err := res.Scan(nil)

		require.EqualError(t, err, "not a string")
		assert.Empty(t, res)
	})

	t.Run("Value", func(t *testing.T) {
		val, err := owners.Value()
		assert.Equal(t, owners.String(), val)
		require.NoError(t, err)
	})
}

func Test_CreateCmd_Validate(t *testing.T) {
	require.EqualError(t, CreateCmd{
		User:   &users.ExampleAlice,
		Name:   "My space",
		Owners: []uuid.UUID{"some-invalid-uuid"},
	}.Validate(), "Owners: (0: must be a valid UUID v4.).")
}

func Test_AddOwnerCmd_Validate(t *testing.T) {
	require.EqualError(t, AddOwnerCmd{
		User:    &users.ExampleAlice,
		Owner:   &users.ExampleAlice,
		SpaceID: "",
	}.Validate(), "SpaceID: cannot be blank.")
}

func Test_RemoveOwnerCmd_Validate(t *testing.T) {
	require.EqualError(t, RemoveOwnerCmd{
		User:    &users.ExampleAlice,
		Owner:   &users.ExampleAlice,
		SpaceID: "",
	}.Validate(), "SpaceID: cannot be blank.")
}
