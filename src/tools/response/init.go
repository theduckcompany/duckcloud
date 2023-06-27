package response

import "net/http"

type Error struct {
	Internal error  `json:"-"`
	Code     int    `json:"-"`
	Message  string `json:"message"`
	Details  string `json:"details"`
}

func (e *Error) Error() string {
	return e.Internal.Error()
}

func (e *Error) Unwrap() error {
	return e.Internal
}

type Writer interface {
	Write(w http.ResponseWriter, res any, statusCode int)
	WriteError(err error, w http.ResponseWriter)
}
