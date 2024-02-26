package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/masterkey"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	masterkeytmpl "github.com/theduckcompany/duckcloud/internal/web/html/templates/masterkey"
)

type MasterKeyPage struct {
	html      html.Writer
	masterkey masterkey.Service
}

func NewMasterKeyPage(html html.Writer, masterkey masterkey.Service) *MasterKeyPage {
	return &MasterKeyPage{
		html:      html,
		masterkey: masterkey,
	}
}

func (h *MasterKeyPage) Register(r chi.Router, mids *router.Middlewares) {
	r.Get("/masterkey-password", h.printMasterKeyPasswordPage)
}

func (h *MasterKeyPage) printMasterKeyPasswordPage(w http.ResponseWriter, r *http.Request) {
	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &masterkeytmpl.ContentTemplate{})
}
