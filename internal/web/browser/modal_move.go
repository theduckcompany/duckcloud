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
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/browser"
)

type moveModalCmd struct {
	ErrorMsg string
	Src      *dfs.PathCmd
	Dst      *dfs.PathCmd
}

type moveModalHandler struct {
	auth   *auth.Authenticator
	spaces spaces.Service
	html   html.Writer
	uuid   uuid.Service
	fs     dfs.Service
}

func newMoveModalHandler(
	auth *auth.Authenticator,
	spaces spaces.Service,
	html html.Writer,
	uuid uuid.Service,
	fs dfs.Service,
) *moveModalHandler {
	return &moveModalHandler{auth, spaces, html, uuid, fs}
}

func (h *moveModalHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/browser/move", h.getMoveModal)
	r.Post("/browser/move", h.handleMoveReq)
}

func (h *moveModalHandler) getMoveModal(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	srcPath, dstPath, abort := h.getMoveParams(w, r)
	if abort {
		return
	}

	lastElem := r.URL.Query().Get("last")
	if lastElem != "" {
		h.renderMoreContent(w, r, &moveModalCmd{
			ErrorMsg: "",
			Src:      srcPath,
			Dst:      dstPath,
		}, lastElem)
		return
	}

	h.renderMoveModal(w, r, &moveModalCmd{
		ErrorMsg: "",
		Src:      srcPath,
		Dst:      dstPath,
	})
}

func (h *moveModalHandler) renderMoveModal(w http.ResponseWriter, r *http.Request, cmd *moveModalCmd) {
	status := http.StatusOK
	if cmd.ErrorMsg != "" {
		status = http.StatusUnprocessableEntity
	}

	childs, err := h.fs.ListDir(r.Context(), cmd.Dst, &storage.PaginateCmd{
		StartAfter: map[string]string{"name": ""},
		Limit:      PageSize,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to list dir for elem %s: %w", cmd.Dst.Path(), err))
		return
	}

	srcInode, err := h.fs.Get(r.Context(), cmd.Src)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the source file %s: %w", cmd.Src.Path(), err))
		return
	}

	folderContent := make(map[dfs.PathCmd]dfs.INode, len(childs))
	for _, child := range childs {
		folderContent[*dfs.NewPathCmd(cmd.Dst.Space(), path.Join(cmd.Dst.Path(), child.Name()))] = child
	}

	h.html.WriteHTMLTemplate(w, r, status, &browser.MoveTemplate{
		SrcPath:       cmd.Src,
		SrcInode:      srcInode,
		DstPath:       cmd.Dst,
		FolderContent: folderContent,
		PageSize:      PageSize,
	})
}

func (h *moveModalHandler) renderMoreContent(w http.ResponseWriter, r *http.Request, cmd *moveModalCmd, lastElem string) {
	childs, err := h.fs.ListDir(r.Context(), cmd.Dst, &storage.PaginateCmd{
		StartAfter: map[string]string{"name": lastElem},
		Limit:      PageSize,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to ListDir: %w", err))
		return
	}

	folderContent := make(map[dfs.PathCmd]dfs.INode, len(childs))
	for _, child := range childs {
		folderContent[*dfs.NewPathCmd(cmd.Dst.Space(), path.Join(cmd.Dst.Path(), child.Name()))] = child
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &browser.MoveRowsTemplate{
		SrcPath:       cmd.Src,
		DstPath:       cmd.Dst,
		FolderContent: folderContent,
		PageSize:      PageSize,
	})
}

func (h *moveModalHandler) handleMoveReq(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	srcPath, dstPath, abort := h.getMoveParams(w, r)
	if abort {
		return
	}

	err := h.fs.Move(ctx, &dfs.MoveCmd{
		Src:     srcPath,
		Dst:     dfs.NewPathCmd(dstPath.Space(), path.Join(dstPath.Path(), path.Base(srcPath.Path()))),
		MovedBy: user,
	})
	if errors.Is(err, errs.ErrValidation) {
		h.renderMoveModal(w, r, &moveModalCmd{
			ErrorMsg: err.Error(),
			Src:      srcPath,
			Dst:      dstPath,
		})
		return
	}

	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to move the file: %w", err))
		return
	}

	w.Header().Add("HX-Trigger", "refreshFolder")
	w.Header().Add("HX-Reswap", "none")
	w.WriteHeader(http.StatusOK)
}

func (h *moveModalHandler) getMoveParams(w http.ResponseWriter, r *http.Request) (*dfs.PathCmd, *dfs.PathCmd, bool) {
	srcPath := r.FormValue("srcPath")
	if len(srcPath) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, true
	}

	dstPath := r.FormValue("dstPath")
	if len(dstPath) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, true
	}

	spaceID, err := h.uuid.Parse(r.FormValue("spaceID"))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, nil, true
	}

	space, err := h.spaces.GetByID(r.Context(), spaceID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the space: %w", err))
		return nil, nil, true
	}

	srcCmd := dfs.NewPathCmd(space, srcPath)
	dstCmd := dfs.NewPathCmd(space, dstPath)

	return srcCmd, dstCmd, false
}
