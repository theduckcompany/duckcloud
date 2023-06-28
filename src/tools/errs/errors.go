package errs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrBadRequest   = fmt.Errorf("bad request")      // 400
	ErrUnauthorized = fmt.Errorf("unauthorized")     // 401
	ErrNotFound     = fmt.Errorf("not found")        // 404
	ErrValidation   = fmt.Errorf("validation error") // 422
	ErrUnhandled    = fmt.Errorf("unhandled error")  // 500
)

type errResponse struct {
	Message string `json:"message"`
}

// Error error
type Error struct {
	code error
	err  error
	msg  string
}

func (t *Error) Code() int {
	switch t.code {
	case ErrBadRequest:
		return http.StatusBadRequest
	case ErrUnauthorized:
		return http.StatusUnauthorized
	case ErrNotFound:
		return http.StatusNotFound
	case ErrValidation:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func (t *Error) Error() string {
	return strings.Join([]string{t.code.Error(), t.err.Error()}, ": ")
}

func (t *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&errResponse{Message: t.msg})
}

func (t *Error) Unwrap() error {
	return t.err
}

func (t *Error) Is(err error) bool {
	return err == t.code || err == t.err
}

func BadRequest(err error, msg string) error {
	return &Error{code: ErrBadRequest, err: err, msg: msg}
}

func ValidationError(err error) error {
	return &Error{code: ErrValidation, err: err, msg: err.Error()}
}

func NotFound(err error, msg string) error {
	return &Error{code: ErrNotFound, err: err, msg: msg}
}

func Unauthorized(err error, msg string) error {
	return &Error{code: ErrUnauthorized, err: err, msg: msg}
}

func Unhandled(err error) error {
	return &Error{code: ErrUnhandled, err: err, msg: "internal error"}
}
