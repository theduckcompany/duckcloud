package response

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	v "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/unrolled/render"
)

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		Name          string
		Input         error
		ExpectedCode  int
		ExpectedJSON  string
		ExpectedError string
	}{
		{
			Name:          "BadRequest",
			Input:         errs.BadRequest(errors.New("some detailed error"), "invalid stuff"),
			ExpectedCode:  http.StatusBadRequest,
			ExpectedJSON:  `{ "message": "invalid stuff" }`,
			ExpectedError: "bad request: some detailed error",
		},
		{
			Name:          "Unauthorized",
			Input:         errs.Unauthorized(errors.New("some detailed error"), "don't have permissions"),
			ExpectedCode:  http.StatusUnauthorized,
			ExpectedJSON:  `{ "message": "don't have permissions" }`,
			ExpectedError: "unauthorized: some detailed error",
		},
		{
			Name:          "NotFound",
			Input:         errs.NotFound(errors.New("some detailed error"), "don't exists"),
			ExpectedCode:  http.StatusNotFound,
			ExpectedJSON:  `{ "message": "don't exists" }`,
			ExpectedError: "not found: some detailed error",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			resWriter := New(render.New())

			r := httptest.NewRequest(http.MethodGet, "/foo", nil)
			w := httptest.NewRecorder()

			resWriter.WriteJSONError(w, r, test.Input)

			res := w.Result()
			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			assert.EqualError(t, test.Input, test.ExpectedError)
			assert.Equal(t, test.ExpectedCode, res.StatusCode)
			assert.JSONEq(t, test.ExpectedJSON, string(body))
		})
	}
}

type someCmd struct {
	Username string
	Email    string
}

func (t someCmd) Validate() error {
	return v.ValidateStruct(&t,
		v.Field(&t.Username, v.Required, v.Length(2, 5)),
		v.Field(&t.Email, v.Required, is.EmailFormat, v.Length(1, 1000)),
	)
}

func TestWriteJSONError_validation(t *testing.T) {
	resWriter := New(render.New())

	r := httptest.NewRequest(http.MethodGet, "/foo", nil)
	w := httptest.NewRecorder()

	cmd := someCmd{Username: "valid", Email: "invalid-input"}
	err := errs.ValidationError(cmd.Validate())
	resWriter.WriteJSONError(w, r, err)

	res := w.Result()
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.EqualError(t, err, "validation error: Email: must be a valid email address.")
	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	assert.JSONEq(t, `{"message": "Email: must be a valid email address."}`, string(body))
}

func TestWriteUnhandledError(t *testing.T) {
	resWriter := New(render.New())

	r := httptest.NewRequest(http.MethodGet, "/foo", nil)
	w := httptest.NewRecorder()

	err := errors.New("some unknown error")
	resWriter.WriteJSONError(w, r, err)

	res := w.Result()
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.JSONEq(t, `{ "message": "internal error" }`, string(body))
}
