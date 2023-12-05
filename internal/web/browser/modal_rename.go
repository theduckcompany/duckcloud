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
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

type renameModalCmd struct {
	ErrorMsg string
	Space    *spaces.Space
	Value    string
	Path     string
}

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

	fs := h.fs.GetSpaceFS(space)

	_, err := fs.Get(ctx, &dfs.PathCmd{Space: space, Path: filePath})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the file %q: %w", filePath, err))
	}

	h.renderRenameModal(w, r, &renameModalCmd{
		ErrorMsg: "",
		Space:    space,
		Path:     filePath,
		Value:    path.Base(filePath),
	})
}

func (h *renameModalHandler) renderRenameModal(w http.ResponseWriter, r *http.Request, cmd *renameModalCmd) {
	status := http.StatusOK
	if cmd.ErrorMsg != "" {
		status = http.StatusUnprocessableEntity
	}

	value := cmd.Value
	// endSelection indicate to the js when to stop the text selection.
	// For the files we want to select only the name, without the extension in
	// order to allow a quick name change without impacting the extension.
	var endSelection int

	for i := len(value) - 1; i >= 0 && value[i] != '/'; i-- {
		if value[i] == '.' {
			endSelection = i
			break
		}
	}

	if endSelection == 0 {
		endSelection = len(value)
	}

	h.html.WriteHTML(w, r, status, "browser/rename-form.tmpl", map[string]interface{}{
		"error":        cmd.ErrorMsg,
		"path":         cmd.Path,
		"value":        value,
		"spaceID":      cmd.Space.ID(),
		"endSelection": endSelection,
	})
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

	fs := h.fs.GetSpaceFS(space)

	inode, err := fs.Get(ctx, &dfs.PathCmd{Space: space, Path: filePath})
	if errors.Is(err, errs.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err = fs.Rename(ctx, inode, r.FormValue("name"))
	if errors.Is(err, errs.ErrValidation) {
		h.renderRenameModal(w, r, &renameModalCmd{
			ErrorMsg: err.Error(),
			Space:    space,
			Value:    r.FormValue("name"),
			Path:     filePath,
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
