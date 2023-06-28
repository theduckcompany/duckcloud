package response

import "net/http"

type Writer interface {
	Write(w http.ResponseWriter, r *http.Request, res any, statusCode int)
	WriteError(err error, w http.ResponseWriter, r *http.Request)
}
