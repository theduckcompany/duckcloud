package masterkey

import (
	"net/http"
)

type HTTPMiddleware struct {
	masterkey Service
}

func NewHTTPMiddleware(masterkey Service) *HTTPMiddleware {
	return &HTTPMiddleware{
		masterkey: masterkey,
	}
}

func (m *HTTPMiddleware) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !m.masterkey.IsMasterKeyLoaded() {
			http.Redirect(w, r, "/master-password/ask", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
