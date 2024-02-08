package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	err error
	msg string
}

func (t *Error) Code() int {
	switch {
	case errors.Is(t.err, ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(t.err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(t.err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(t.err, ErrValidation):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func (t *Error) Error() string {
	return t.err.Error()
}

func (t *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(&errResponse{Message: t.msg})
}

func (t *Error) Unwrap() error {
	return t.err
}

func (t *Error) Is(err error) bool {
	return err == t.err
}

func BadRequest(err error, msgAndArgs ...any) error {
	return &Error{err: fmt.Errorf("%w: %w", ErrBadRequest, err), msg: messageFromMsgAndArgs(ErrBadRequest, msgAndArgs...)}
}

func Validation(err error) error {
	return &Error{err: fmt.Errorf("%w: %w", ErrValidation, err), msg: err.Error()}
}

func NotFound(err error, msgAndArgs ...any) error {
	return &Error{err: fmt.Errorf("%w: %w", ErrNotFound, err), msg: messageFromMsgAndArgs(ErrNotFound, msgAndArgs...)}
}

func Unauthorized(err error, msgAndArgs ...any) error {
	return &Error{err: fmt.Errorf("%w: %w", ErrUnauthorized, err), msg: messageFromMsgAndArgs(ErrUnauthorized, msgAndArgs...)}
}

func Internal(err error) error {
	return &Error{err: fmt.Errorf("%w: %w", ErrInternal, err), msg: "internal error"}
}

func Unhandled(err error) error {
	return &Error{err: fmt.Errorf("%w: %w", ErrUnhandled, err), msg: "internal error"}
}

func messageFromMsgAndArgs(baseError error, msgAndArgs ...any) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return baseError.Error()
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
