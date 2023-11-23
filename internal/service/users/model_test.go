package users

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func Test_User_Getters(t *testing.T) {
	assert.Equal(t, ExampleAlice.id, ExampleAlice.ID())
	assert.Equal(t, ExampleAlice.username, ExampleAlice.Username())
	assert.Equal(t, ExampleAlice.isAdmin, ExampleAlice.IsAdmin())
	assert.Equal(t, ExampleAlice.createdAt, ExampleAlice.CreatedAt())
	assert.Equal(t, ExampleAlice.status, ExampleAlice.Status())
}

func Test_CreateUserRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateCmd))
}

func Test_CreateUserRequest_Validate_success(t *testing.T) {
	err := CreateCmd{
		Username: "some-username",
		Password: secret.NewText("myLittleSecret"),
		IsAdmin:  true,
	}.Validate()

	assert.NoError(t, err)
}

func Test_UpdatePasswordCmd(t *testing.T) {
	err := UpdatePasswordCmd{
		UserID:      uuid.UUID("some-invalid-id"),
		NewPassword: secret.NewText("foobar1234"),
	}.Validate()

	assert.EqualError(t, err, "UserID: must be a valid UUID v4.")
}
