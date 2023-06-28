package jwt

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Peltoche/neurone/src/tools/response"
)

var (
	ErrInvalidAccessToken = fmt.Errorf("invalid access token")
	ErrMissingAccessToken = fmt.Errorf("missing access token")
	ErrInvalidFormat      = fmt.Errorf("invalid format")
)

type Error struct {
	error
}

func (e *Error) Unwrap() error {
	return e.error
}

func (e *Error) As(target any) bool {
	rerr, ok := target.(**response.Error)
	if !ok {
		return false
	}

	switch {
	case errors.Is(e, ErrInvalidAccessToken),
		errors.Is(e, ErrInvalidFormat):
		(*rerr) = &response.Error{
			Internal: e,
			Code:     http.StatusUnauthorized,
			Message:  "invalid access token",
		}
		return true

	case errors.Is(e, ErrMissingAccessToken):
		(*rerr) = &response.Error{
			Internal: e,
			Code:     http.StatusUnauthorized,
			Message:  "missing access token",
		}
		return true

	default:
		return false
	}
}
