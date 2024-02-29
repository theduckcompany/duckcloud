package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/auth"
)

type MasterAskPasswordPage struct {
	html      html.Writer
	masterkey masterkey.Service
}

func NewAskMasterPasswordPage(html html.Writer, masterkey masterkey.Service) *MasterAskPasswordPage {
	return &MasterAskPasswordPage{
		html:      html,
		masterkey: masterkey,
	}
}

func (h *MasterAskPasswordPage) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		// Remove the mids.Masterkey middleware to avoid an infinit redirection loop
		r = r.With(mids.Logger, mids.RealIP, mids.StripSlashed, mids.CORS)
	}

	r.Get("/master-password/ask", h.printMasterKeyPasswordPage)
	r.Post("/master-password/ask", h.postForm)
}

func (h *MasterAskPasswordPage) printMasterKeyPasswordPage(w http.ResponseWriter, r *http.Request) {
	if h.masterkey.IsMasterKeyLoaded() {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.MasterPasswordPageTmpl{})
}

func (h *MasterAskPasswordPage) postForm(w http.ResponseWriter, r *http.Request) {
	password := secret.NewText(r.FormValue("password"))

	err := h.masterkey.LoadMasterKeyFromPassword(r.Context(), &password)
	if errors.Is(err, errs.ErrBadRequest) {
		h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.MasterPasswordPageTmpl{
			ErrorMsg: "invalid password",
		})
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to load the master key from password: %w", err))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
