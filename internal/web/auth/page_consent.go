package auth

import (
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/auth"
)

type consentPage struct {
	html         html.Writer
	auth         *Authenticator
	uuid         uuid.Service
	clients      oauthclients.Service
	oauthConsent oauthconsents.Service
}

func newConsentPage(
	html html.Writer,
	auth *Authenticator,
	clients oauthclients.Service,
	oauthConsent oauthconsents.Service,
	tools tools.Tools,
) *consentPage {
	return &consentPage{
		html:         html,
		auth:         auth,
		clients:      clients,
		oauthConsent: oauthConsent,
		uuid:         tools.UUID(),
	}
}

func (h *consentPage) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.HandleFunc("/consent", h.printPage)
}

func (h *consentPage) printPage(w http.ResponseWriter, r *http.Request) {
	user, session, abort := h.auth.GetUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	reqID, ok := r.Context().Value(middleware.RequestIDKey).(string)
	if !ok {
		reqID = "????"
	}

	clientID, err := h.uuid.Parse(r.FormValue("client_id"))
	if err != nil {
		h.html.WriteHTMLTemplate(w, r, http.StatusBadRequest, &auth.ErrorPageTmpl{
			ErrorMsg:  "invalid client_id",
			RequestID: reqID,
		})
		return
	}

	client, err := h.clients.GetByID(r.Context(), clientID)
	if errors.Is(err, errs.ErrNotFound) {
		h.html.WriteHTMLTemplate(w, r, http.StatusBadRequest, &auth.ErrorPageTmpl{
			ErrorMsg:  "invalid client_id",
			RequestID: reqID,
		})
		return
	}
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, err)
		return
	}

	if r.Method == http.MethodPost {
		consent, err := h.oauthConsent.Create(r.Context(), &oauthconsents.CreateCmd{
			UserID:       user.ID(),
			SessionToken: session.Token().Raw(),
			ClientID:     client.GetID(),
			Scopes:       strings.Split(r.FormValue("scope"), ","),
		})
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, err)
			return
		}

		r.Form.Add("consent_id", string(consent.ID()))
		http.Redirect(w, r, "/auth/authorize?"+r.Form.Encode(), http.StatusFound)
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.ConsentPageTmpl{
		ClientName: client.Name(),
		Username:   user.Username(),
		Scopes:     strings.Split(r.FormValue("scope"), ","),
		Redirect:   template.URL("/consent?" + r.Form.Encode()),
	})
}
