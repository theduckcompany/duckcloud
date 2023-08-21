package davsessions

import (
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func Test_CreateUserRequest_is_validatable(t *testing.T) {
	assert.Implements(t, (*validation.Validatable)(nil), new(CreateCmd))
}

func Test_CreateUserRequest_Validate_success(t *testing.T) {
	err := CreateCmd{
		UserID: uuid.UUID("2c6b2615-6204-4817-a126-b6c13074afdf"),
		FSRoot: uuid.UUID("d43afe5b-5c3c-4ba4-a08c-031d701f2aef"),
	}.Validate()

	assert.NoError(t, err)
}

func TestDavSession_Getters(t *testing.T) {
}
