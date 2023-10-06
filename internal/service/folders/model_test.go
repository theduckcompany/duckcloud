package folders

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_CreateCmd_Validate(t *testing.T) {
	assert.EqualError(t, CreatePersonalFolderCmd{
		Name:  "My folder",
		Owner: uuid.UUID("some-invalid-uuid"),
	}.Validate(), "Owner: must be a valid UUID v4.")
}

func Test_Folder_Getters(t *testing.T) {
	assert.Equal(t, ExampleAlicePersonalFolder.ID(), ExampleAlicePersonalFolder.id)
	assert.Equal(t, ExampleAlicePersonalFolder.Name(), ExampleAlicePersonalFolder.name)
	assert.Equal(t, ExampleAlicePersonalFolder.IsPublic(), ExampleAlicePersonalFolder.isPublic)
	assert.Equal(t, ExampleAlicePersonalFolder.Owners(), ExampleAlicePersonalFolder.owners)
	assert.Equal(t, ExampleAlicePersonalFolder.RootFS(), ExampleAlicePersonalFolder.rootFS)
	assert.Equal(t, ExampleAlicePersonalFolder.CreatedAt(), ExampleAlicePersonalFolder.createdAt)
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
