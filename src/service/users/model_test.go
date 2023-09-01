package users

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func Test_User_Getters(t *testing.T) {
	assert.Equal(t, ExampleAlice.id, ExampleAlice.ID())
	assert.Equal(t, ExampleAlice.username, ExampleAlice.Username())
	assert.Equal(t, ExampleAlice.isAdmin, ExampleAlice.IsAdmin())
	assert.Equal(t, ExampleAlice.createdAt, ExampleAlice.CreatedAt())
	assert.Equal(t, ExampleAlice.fsRoot, ExampleAlice.RootFS())
	assert.Equal(t, ExampleAlice.status, ExampleAlice.Status())
}

func Test_CreateUserRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateCmd))
}

func Test_CreateUserRequest_Validate_success(t *testing.T) {
	err := CreateCmd{
		Username: "some-username",
		Password: "myLittleSecret",
		IsAdmin:  true,
	}.Validate()

	assert.NoError(t, err)
}
