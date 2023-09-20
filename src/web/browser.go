package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/fs"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/router"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
	"github.com/theduckcompany/duckcloud/src/web/html"
)

const MaxMemoryCache = 20 * 1024 * 1024 // 20MB

var ErrInvalidFolderID = errors.New("invalid folderID")

type browserHandler struct {
	html    html.Writer
	folders folders.Service
	inodes  inodes.Service
	files   files.Service
	uuid    uuid.Service
	auth    *Authenticator
}

func newBrowserHandler(
	tools tools.Tools,
	html html.Writer,
	folders folders.Service,
	inodes inodes.Service,
	files files.Service,
	auth *Authenticator,
) *browserHandler {
	return &browserHandler{
		html:    html,
		folders: folders,
		inodes:  inodes,
		files:   files,
		uuid:    tools.UUID(),
		auth:    auth,
	}
}

func (h *browserHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/browser", h.getBrowserHome)
	r.Post("/browser/upload", h.upload)
	r.Get("/browser/*", h.getBrowserContent)
}

func (h *browserHandler) String() string {
	return "web.browser"
}

func (h *browserHandler) getBrowserHome(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	folders, err := h.folders.GetAllUserFolders(r.Context(), user.ID(), nil)
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/home.tmpl", map[string]interface{}{
		"folders": folders,
	})
}

func (h *browserHandler) getBrowserContent(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	// For the path "/browser/{{folderID}}/foo/bar/baz" the elems variable will have for content:
	// []string{"", "browser", "{{folderID}}", "/foo/bar/baz"}
	elems := strings.SplitN(r.URL.Path, "/", 4)

	// no need to check elems len as the url format force a len of 3 minimum
	folderID, err := h.uuid.Parse(elems[2])
	if err != nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
		return
	}

	folder, err := h.folders.GetByID(r.Context(), folderID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to folders.GetByID: %w", err))
		return
	}

	if folder == nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
		return
	}

	fullPath := "/"
	if len(elems) == 4 {
		fullPath = path.Clean(elems[3])
	}

	inode, err := h.inodes.Get(r.Context(), &inodes.PathCmd{
		Root:     folder.RootFS(),
		FullName: fullPath,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to inodes.Get: %w", err))
		return
	}

	if inode == nil {
		w.Header().Set("Location", path.Join("/browser/", string(folder.ID())))
		w.WriteHeader(http.StatusFound)
		return
	}

	var dirContent []inodes.INode
	if inode.IsDir() {
		dirContent, err = h.inodes.Readdir(r.Context(), &inodes.PathCmd{
			Root:     folder.RootFS(),
			FullName: fullPath,
		}, nil)
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to inodes.Readdir: %w", err))
			return
		}
	}

	folders, err := h.folders.GetAllUserFolders(r.Context(), user.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetallUserFolders: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
		"fullPath":   fullPath,
		"folder":     folder,
		"breadcrumb": generateBreadCrumb(folder, fullPath),
		"folders":    folders,
		"inodes":     dirContent,
	})
}

func (h *browserHandler) upload(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var name, rawFolderID, rootPath, relPath []byte

	for {
		p, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if p.FileName() != "" {
			folderID, err := h.uuid.Parse(string(rawFolderID))
			if err != nil {
				fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			cmd := lauchUploadCmd{
				user:       user,
				name:       string(name),
				folderID:   folderID,
				relPath:    string(relPath),
				rootPath:   string(rootPath),
				fileReader: p,
			}

			defer p.Close()
			err = h.lauchUpload(r.Context(), &cmd)
			if err != nil {
				fmt.Printf("failed to upload: %s -> %#v\n\n\n", err, cmd)
			}

			return
		}

		switch p.FormName() {
		case "name":
			name, err = io.ReadAll(p)
		// case "type":
		// 	ftype, err = io.ReadAll(p)
		case "rootPath":
			rootPath, err = io.ReadAll(p)
		case "folderID":
			rawFolderID, err = io.ReadAll(p)
		case "relativePath":
			relPath, err = io.ReadAll(p)
		}
		if err != nil {
			fmt.Fprintf(w, `<div class="alert alert-danger role="alert">%s</div>`, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
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

type lauchUploadCmd struct {
	user       *users.User
	name       string
	folderID   uuid.UUID
	rootPath   string
	relPath    string
	fileReader io.Reader
}

func (h *browserHandler) lauchUpload(ctx context.Context, cmd *lauchUploadCmd) error {
	folder, err := h.folders.GetByID(ctx, cmd.folderID)
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

	if !slices.Contains[[]uuid.UUID, uuid.UUID](folder.Owners(), cmd.user.ID()) {
		return ErrInvalidFolderID
	}

	fs := fs.NewFSService(h.inodes, h.files, folder, h.folders)

	var fullPath string
	if cmd.relPath == "null" {
		fullPath = path.Join(cmd.rootPath, cmd.name)
	} else {
		fullPath = path.Join(cmd.rootPath, cmd.relPath)

		_, err = h.inodes.MkdirAll(ctx, &inodes.PathCmd{
			Root:     folder.RootFS(),
			FullName: path.Dir(fullPath),
		})
		if err != nil {
			return fmt.Errorf("failed to Mkdirall: %w", err)
		}
	}

	if fullPath[0] == '/' {
		fullPath = fullPath[1:]
	}

	file, err := fs.OpenFile(ctx, fullPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to OpenFile: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, cmd.fileReader)
	if err != nil {
		return fmt.Errorf("failed to Copy: %w", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("failed to Close file: %w", err)
	}

	return nil
}
