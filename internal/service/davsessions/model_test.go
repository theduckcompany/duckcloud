package davsessions

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDavSession_Getters(t *testing.T) {
	assert.Equal(t, ExampleAliceSession.id, ExampleAliceSession.ID())
	assert.Equal(t, ExampleAliceSession.userID, ExampleAliceSession.UserID())
	assert.Equal(t, ExampleAliceSession.name, ExampleAliceSession.Name())
	assert.Equal(t, ExampleAliceSession.folders, ExampleAliceSession.FoldersIDs())
	assert.Equal(t, ExampleAliceSession.username, ExampleAliceSession.Username())
	assert.Equal(t, ExampleAliceSession.createdAt, ExampleAliceSession.CreatedAt())
}

func Test_CreateUserRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateCmd))
}

func Test_CreateRequest_Validate_success(t *testing.T) {
	err := CreateCmd{
		Name:    ExampleAliceSession.Name(),
		UserID:  uuid.UUID("2c6b2615-6204-4817-a126-b6c13074afdf"),
		Folders: []uuid.UUID{folders.ExampleAlicePersonalFolder.ID()},
	}.Validate()

	assert.NoError(t, err)
}

func Test_CreateRequest_Validate_with_no_folders(t *testing.T) {
	err := CreateCmd{
		Name:    ExampleAliceSession.Name(),
		UserID:  uuid.UUID("2c6b2615-6204-4817-a126-b6c13074afdf"),
		Folders: []uuid.UUID{},
	}.Validate()

	assert.EqualError(t, err, "Folders: cannot be blank.")
}

func Test_DeleteRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(DeleteCmd))
}

func Test_DeleteRequest_Validate_success(t *testing.T) {
	err := DeleteCmd{
		UserID:    uuid.UUID("2c6b2615-6204-4817-a126-b6c13074afdf"),
		SessionID: uuid.UUID("d43afe5b-5c3c-4ba4-a08c-031d701f2aef"),
	}.Validate()

	assert.NoError(t, err)
}

func Test_Folders_Getters(t *testing.T) {
	var raw string
	owners := Folders([]uuid.UUID{"some-id-1", "some-id-2"})

	t.Run("String", func(t *testing.T) {
		raw = owners.String()
		assert.Equal(t, "some-id-1,some-id-2", raw)
	})

	t.Run("Scan", func(t *testing.T) {
		res := Folders{}
		err := res.Scan(raw)

		assert.NoError(t, err)
		assert.Equal(t, owners, res)
	})

	t.Run("Scan error", func(t *testing.T) {
		res := Folders{}
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