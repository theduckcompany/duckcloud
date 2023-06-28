package users

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func Test_CreateUserRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateUserRequest))
}

func Test_CreateUserRequest_Validate_success(t *testing.T) {
	err := CreateUserRequest{
		Username: "some-username",
		Email:    "some@email.com",
		Password: "myLittleSecret",
	}.Validate()

	assert.NoError(t, err)
}
