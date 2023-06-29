package response

import "net/http"

type Writer interface {
	WriteJSON(w http.ResponseWriter, statusCode int, res any)
	WriteJSONError(w http.ResponseWriter, err error)
	WriteHTML(w http.ResponseWriter, status int, template string, args any)
}
