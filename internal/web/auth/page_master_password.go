package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/auth"
)

type MasterPasswordPage struct {
	html      html.Writer
	masterkey masterkey.Service
}

func NewMasterPasswordPage(html html.Writer, masterkey masterkey.Service) *MasterPasswordPage {
	return &MasterPasswordPage{
		html:      html,
		masterkey: masterkey,
	}
}

func (h *MasterPasswordPage) Register(r chi.Router, mids *router.Middlewares) {
	r.Get("/master-password", h.printMasterKeyPasswordPage)
	r.Post("/master-password", h.postForm)
}

func (h *MasterPasswordPage) printMasterKeyPasswordPage(w http.ResponseWriter, r *http.Request) {
	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.MasterPasswordPageTmpl{})
}

func (h *MasterPasswordPage) postForm(w http.ResponseWriter, r *http.Request) {
	password := secret.NewText(r.FormValue("password"))

	h.masterkey.Open(password)
}
