package browser

import (
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/browser"
)

type renameModalHandler struct {
	auth   *auth.Authenticator
	spaces spaces.Service
	html   html.Writer
	uuid   uuid.Service
	fs     dfs.Service
}

func newRenameModalHandler(
	auth *auth.Authenticator,
	spaces spaces.Service,
	html html.Writer,
	uuid uuid.Service,
	fs dfs.Service,
) *renameModalHandler {
	return &renameModalHandler{auth, spaces, html, uuid, fs}
}

func (h *renameModalHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/browser/rename", h.getRenameModal)
	r.Post("/browser/rename", h.handleRenameReq)
}

func (h *renameModalHandler) getRenameModal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	space, filePath, abort := h.getRenameParams(w, r)
	if abort {
		return
	}

	targetPath := dfs.NewPathCmd(space, filePath)

	_, err := h.fs.Get(ctx, targetPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the file %q: %w", filePath, err))
	}

	h.renderRenameModal(w, r, &browser.RenameTemplate{
		Error:               nil,
		Target:              targetPath,
		FieldValue:          path.Base(filePath),
		FieldValueSelection: 0, // Filled by renderRenameModal
	})
}

func (h *renameModalHandler) renderRenameModal(w http.ResponseWriter, r *http.Request, cmd *browser.RenameTemplate) {
	status := http.StatusOK
	if cmd.Error != nil {
		status = http.StatusUnprocessableEntity
	}

	value := cmd.FieldValue

	for i := len(value) - 1; i >= 0 && value[i] != '/'; i-- {
		if value[i] == '.' {
			cmd.FieldValueSelection = i
			break
		}
	}

	if cmd.FieldValueSelection == 0 {
		cmd.FieldValueSelection = len(value)
	}

	h.html.WriteHTMLTemplate(w, r, status, cmd)
}

func (h *renameModalHandler) handleRenameReq(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	space, filePath, abort := h.getRenameParams(w, r)
	if abort {
		return
	}

	targetPath := dfs.NewPathCmd(space, filePath)

	inode, err := h.fs.Get(ctx, targetPath)
	if errors.Is(err, errs.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err = h.fs.Rename(ctx, inode, r.FormValue("name"))
	if errors.Is(err, errs.ErrValidation) {
		h.renderRenameModal(w, r, &browser.RenameTemplate{
			Error:               ptr.To(err.Error()),
			Target:              targetPath,
			FieldValue:          r.FormValue("name"),
			FieldValueSelection: 0, // Filled by renderRenameModal
		})
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to rename the file: %w", err))
	}

	w.Header().Add("HX-Trigger", "refreshFolder")
	w.Header().Add("HX-Reswap", "none")
	w.WriteHeader(http.StatusOK)
}

func (h *renameModalHandler) getRenameParams(w http.ResponseWriter, r *http.Request) (*spaces.Space, string, bool) {
	path := r.FormValue("path")
	if len(path) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return nil, "", true
	}

	spaceID, err := h.uuid.Parse(r.FormValue("spaceID"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, "", true
	}

	space, err := h.spaces.GetByID(r.Context(), spaceID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the space: %w", err))
		return nil, "", true
	}

	return space, path, false
}
