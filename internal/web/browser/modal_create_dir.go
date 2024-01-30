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

type createDirModalHandler struct {
	auth   *auth.Authenticator
	spaces spaces.Service
	html   html.Writer
	uuid   uuid.Service
	fs     dfs.Service
}

func newCreateDirModalHandler(
	auth *auth.Authenticator,
	spaces spaces.Service,
	html html.Writer,
	uuid uuid.Service,
	fs dfs.Service,
) *createDirModalHandler {
	return &createDirModalHandler{auth, spaces, html, uuid, fs}
}

func (h *createDirModalHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/browser/create-dir", h.getCreateDirModal)
	r.Post("/browser/create-dir", h.handleCreateDirReq)
}

func (h *createDirModalHandler) getCreateDirModal(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	dir := r.URL.Query().Get("dir")
	if dir == "" {
		h.html.WriteHTMLErrorPage(w, r, errors.New("failed to get the dir path from the url query"))
		return
	}

	spaceID, err := h.uuid.Parse(r.URL.Query().Get("space"))
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, errors.New("failed to get the space id from the url query"))
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &browser.CreateDirTemplate{
		DirPath: dir,
		SpaceID: spaceID,
		Error:   nil,
	})
}

func (h *createDirModalHandler) handleCreateDirReq(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	dir := r.FormValue("dirPath")
	name := r.FormValue("name")
	spaceID, err := h.uuid.Parse(r.FormValue("spaceID"))
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, errors.New("invalid space id param"))
		return
	}

	if name == "" {
		h.html.WriteHTMLTemplate(w, r, http.StatusUnprocessableEntity, &browser.CreateDirTemplate{
			DirPath: dir,
			SpaceID: spaceID,
			Error:   ptr.To("Must not be empty"),
		})
		return
	}

	space, err := h.spaces.GetUserSpace(r.Context(), user.ID(), spaceID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, errs.BadRequest(fmt.Errorf("failed to GetUserSpace: %w", err)))
		return
	}

	existingDir, err := h.fs.Get(r.Context(), &dfs.PathCmd{Space: space, Path: path.Join(dir, name)})
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the directory: %w", err))
		return
	}

	if existingDir != nil {
		h.html.WriteHTMLTemplate(w, r, http.StatusUnprocessableEntity, &browser.CreateDirTemplate{
			DirPath: dir,
			SpaceID: spaceID,
			Error:   ptr.To("Already exists"),
		})
		return
	}

	_, err = h.fs.CreateDir(r.Context(), &dfs.CreateDirCmd{
		Path:      &dfs.PathCmd{Space: space, Path: path.Join(dir, name)},
		CreatedBy: user,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to create the directory: %w", err))
		return
	}

	w.Header().Add("HX-Trigger", "refreshFolder")
	w.Header().Add("HX-Reswap", "none")
	w.WriteHeader(http.StatusCreated)
}
