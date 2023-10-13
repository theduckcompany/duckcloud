package errs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrBadRequest   = fmt.Errorf("bad request")  // HTTP code: 400
	ErrUnauthorized = fmt.Errorf("unauthorized") // HTTP code: 401
	ErrNotFound     = fmt.Errorf("not found")    // HTTP code: 404
	ErrValidation   = fmt.Errorf("validation")   // HTTP code: 422
	ErrUnhandled    = fmt.Errorf("unhandled")    // HTTP code: 500
	ErrInternal     = fmt.Errorf("internal")     // HTTP code: 500
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

func BadRequest(err error, msgAndArgs ...any) error {
	return &Error{code: ErrBadRequest, err: err, msg: messageFromMsgAndArgs("bad request", msgAndArgs...)}
}

func Validation(err error) error {
	return &Error{code: ErrValidation, err: err, msg: err.Error()}
}

func NotFound(err error, msgAndArgs ...any) error {
	return &Error{code: ErrNotFound, err: err, msg: messageFromMsgAndArgs("not found", msgAndArgs...)}
}

func Unauthorized(err error, msgAndArgs ...any) error {
	return &Error{code: ErrUnauthorized, err: err, msg: messageFromMsgAndArgs("unauthorized", msgAndArgs...)}
}

func Internal(err error) error {
	return &Error{code: ErrInternal, err: err, msg: "internal error"}
}

func Unhandled(err error) error {
	return &Error{code: ErrUnhandled, err: err, msg: "internal error"}
}

func messageFromMsgAndArgs(defaultMsg string, msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return defaultMsg
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
