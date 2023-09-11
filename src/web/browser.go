package web

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/service/websessions"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/response"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

type browserHandler struct {
	response    response.Writer
	webSessions websessions.Service
	folders     folders.Service
	users       users.Service
	inodes      inodes.Service
	uuid        uuid.Service
}

func newBrowserHandler(
	tools tools.Tools,
	webSessions websessions.Service,
	folders folders.Service,
	users users.Service,
	inodes inodes.Service,
	uuid uuid.Service,
) *browserHandler {
	return &browserHandler{
		response:    tools.ResWriter(),
		webSessions: webSessions,
		folders:     folders,
		users:       users,
		inodes:      inodes,
		uuid:        uuid,
	}
}

func (h *browserHandler) Register(r chi.Router, mids router.Middlewares) {
	browser := r.With(mids.RealIP, mids.StripSlashed, mids.Logger)

	browser.Get("/browser", h.getBrowserHome)
	browser.Get("/browser/*", h.getBrowserContent)
}

func (h *browserHandler) String() string {
	return "web.browser"
}

func (h *browserHandler) getBrowserHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	folders, err := h.folders.GetAllUserFolders(ctx, user.ID(), nil)
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		return
	}

	fullPage := r.Header.Get("HX-Boosted") == "" && r.Header.Get("HX-Request") == ""
	h.response.WriteHTML(w, http.StatusOK, "browser/home.tmpl", fullPage, map[string]interface{}{
		"folders": folders,
	})
}

func (h *browserHandler) getBrowserContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _ := h.getUserAndSession(w, r)
	if user == nil {
		return
	}

	folders, err := h.folders.GetAllUserFolders(ctx, user.ID(), nil)
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		return
	}

	// For the path "/browser/{{folderID}}/foo/bar/baz" the elems variable will have for content:
	// []string{"", "browser", "{{folderID}}", "/foo/bar/baz"}
	elems := strings.SplitN(r.URL.Path, "/", 4)

	if len(elems) < 3 {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
	}

	folderID, err := h.uuid.Parse(elems[2])
	if err != nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
		return
	}

	fullPath := "/"
	if len(elems) == 4 {
		fullPath = path.Clean(elems[3])
	}

	folder, err := h.folders.GetByID(ctx, folderID)
	if err != nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
	}

	if folder == nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">Folder not found</div>`)
		return
	}

	inode, err := h.inodes.Get(ctx, &inodes.PathCmd{
		Root:     folder.RootFS(),
		FullName: fullPath,
	})
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		return
	}

	if inode == nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">File/Dir not found</div>`)
		return
	}

	var dirContent []inodes.INode
	if inode.IsDir() {
		dirContent, err = h.inodes.Readdir(ctx, &inodes.PathCmd{
			Root:     folder.RootFS(),
			FullName: fullPath,
		}, nil)
		if err != nil {
			fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	fullPage := r.Header.Get("HX-Boosted") == "" && r.Header.Get("HX-Request") == ""
	h.response.WriteHTML(w, http.StatusOK, "browser/content.tmpl", fullPage, map[string]interface{}{
		"fullPath":   fullPath,
		"folder":     folder,
		"breadcrumb": generateBreadCrumb(folder, fullPath),
		"folders":    folders,
		"inodes":     dirContent,
	})
}

func (h *browserHandler) getUserAndSession(w http.ResponseWriter, r *http.Request) (*users.User, *websessions.Session) {
	ctx := r.Context()

	currentSession, err := h.webSessions.GetFromReq(r)
	if err != nil || currentSession == nil {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return nil, nil
	}

	user, err := h.users.GetByID(ctx, currentSession.UserID())
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil
	}

	if user == nil {
		_ = h.webSessions.Logout(r, w)
		return nil, nil
	}

	return user, currentSession
}

type breadCrumbElement struct {
	Name    string
	Href    string
	Current bool
}

func generateBreadCrumb(folder *folders.Folder, fullPath string) []breadCrumbElement {
	basePath := path.Join("/browser/", string(folder.ID()))

	res := []breadCrumbElement{{
		Name:    folder.Name(),
		Href:    basePath,
		Current: false,
	}}

	if fullPath == "/" {
		return res
	}

	for _, elem := range strings.Split(fullPath, "/") {
		basePath = path.Join(basePath, elem)

		res = append(res, breadCrumbElement{
			Name:    elem,
			Href:    basePath,
			Current: false,
		})
	}

	res[len(res)-1].Current = true

	return res
}
