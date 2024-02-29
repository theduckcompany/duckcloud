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

type MasterRegisterPasswordPage struct {
	html      html.Writer
	masterkey masterkey.Service
}

func NewRegisterMasterPasswordPage(html html.Writer, masterkey masterkey.Service) *MasterRegisterPasswordPage {
	return &MasterRegisterPasswordPage{
		html:      html,
		masterkey: masterkey,
	}
}

func (h *MasterRegisterPasswordPage) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.Defaults()...)
	}

	r.Get("/master-password/register", h.printPage)
	r.Post("/master-password/register", h.postForm)
}

func (h *MasterRegisterPasswordPage) printPage(w http.ResponseWriter, r *http.Request) {
	if h.masterkey.IsMasterKeyLoaded() {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.RegisterMasterPasswordPageTmpl{})
}

func (h *MasterRegisterPasswordPage) postForm(w http.ResponseWriter, r *http.Request) {
	password := secret.NewText(r.FormValue("password"))
	confirm := secret.NewText(r.FormValue("confirm"))

	if confirm != password {
		h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.RegisterMasterPasswordPageTmpl{
			PasswordError: "",
			ConfirmError:  "not identical",
		})
		return
	}

	if len(password.Raw()) < 8 {
		h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.RegisterMasterPasswordPageTmpl{
			PasswordError: "too short",
			ConfirmError:  "",
		})
		return
	}

	err := h.masterkey.GenerateMasterKey(r.Context(), &password)
	if errors.Is(err, errs.ErrBadRequest) {
		h.html.WriteHTMLTemplate(w, r, http.StatusOK, &auth.RegisterMasterPasswordPageTmpl{
			PasswordError: "invalid password",
			ConfirmError:  "",
		})
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to load the master key from password: %w", err))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
