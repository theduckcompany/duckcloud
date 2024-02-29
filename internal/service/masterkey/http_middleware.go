package masterkey

import (
	"fmt"
	"net/http"

	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type HTTPMiddleware struct {
	masterkey Service
	html      html.Writer
}

func NewHTTPMiddleware(masterkey Service, html html.Writer) *HTTPMiddleware {
	return &HTTPMiddleware{
		masterkey: masterkey,
		html:      html,
	}
}

func (m *HTTPMiddleware) Handle(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if !m.masterkey.IsMasterKeyLoaded() {

			// IsRegistered is call only if ww see that the key is not loaded because
			// it should append only for the first calls and it's way more costly.
			isRegistered, err := m.masterkey.IsMasterKeyRegistered(r.Context())
			if err != nil {
				m.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to check if the master key is registered: %w", err))
				return
			}

			switch {
			case isRegistered && r.URL.Path != "/master-password/ask": // Registered but not loaded -> ask for the password.
				http.Redirect(w, r, "/master-password/ask", http.StatusSeeOther)
			case !isRegistered && r.URL.Path != "/master-password/register": // Not registered -> ask for a new password and generate the key.
				http.Redirect(w, r, "/master-password/register", http.StatusSeeOther)
			default:
				next.ServeHTTP(w, r)
			}

			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
