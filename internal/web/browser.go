package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

const MaxMemoryCache = 20 * 1024 * 1024 // 20MB

var ErrInvalidFolderID = errors.New("invalid folderID")

type browserHandler struct {
	html    html.Writer
	folders folders.Service
	files   files.Service
	uuid    uuid.Service
	auth    *Authenticator
	fs      dfs.Service
}

func newBrowserHandler(
	tools tools.Tools,
	html html.Writer,
	folders folders.Service,
	files files.Service,
	auth *Authenticator,
	fs dfs.Service,
) *browserHandler {
	return &browserHandler{
		html:    html,
		folders: folders,
		files:   files,
		uuid:    tools.UUID(),
		auth:    auth,
		fs:      fs,
	}
}

func (h *browserHandler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/browser", h.redirectDefaultBrowser)
	r.Post("/browser/upload", h.upload)
	r.Get("/browser/*", h.getBrowserContent)
	r.Delete("/browser/*", h.deleteAll)
}

func (h *browserHandler) String() string {
	return "web.browser"
}

func (h *browserHandler) redirectDefaultBrowser(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	w.Header().Set("Location", path.Join("/browser/", string(user.DefaultFolder())))
	w.WriteHeader(http.StatusFound)
}

func (h *browserHandler) getBrowserContent(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	folder, fullPath, abort := h.getFolderAndPathFromURL(w, r, user)
	if abort {
		return
	}
	fs := h.fs.GetFolderFS(folder)

	h.renderBrowserContent(w, r, user, fs, fullPath)
}

func (h *browserHandler) upload(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get mutlipart reader: %w", err))
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
				h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("invalid folder id %q", rawFolderID))
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
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
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
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to parse form: %w", err))
			return
		}
	}
}

func (h *browserHandler) deleteAll(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	folder, fullPath, abort := h.getFolderAndPathFromURL(w, r, user)
	if abort {
		return
	}

	fs := h.fs.GetFolderFS(folder)
	err := fs.RemoveAll(r.Context(), fullPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.RemoveAll: %w", err))
		return
	}

	h.renderBrowserContent(w, r, user, fs, path.Dir(fullPath))
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

	if fullPath == "." {
		res[0].Current = true
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
	folder, err := h.folders.GetUserFolder(ctx, cmd.user.ID(), cmd.folderID)
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

	ffs := h.fs.GetFolderFS(folder)

	var fullPath string
	if cmd.relPath == "null" || cmd.relPath == "" {
		fullPath = path.Join(cmd.rootPath, cmd.name)
	} else {
		fullPath = path.Join(cmd.rootPath, cmd.relPath)

		_, err = ffs.CreateDir(ctx, path.Dir(fullPath))
		if err != nil && !errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("failed to CreateDir: %w", err)
		}
	}

	if fullPath[0] == '/' {
		fullPath = fullPath[1:]
	}

	file, err := ffs.CreateFile(ctx, fullPath)
	if err != nil {
		return fmt.Errorf("failed to CreateFile: %w", err)
	}

	err = ffs.Upload(ctx, file, cmd.fileReader)
	if err != nil {
		return fmt.Errorf("failed to Upload file: %w", err)
	}

	return nil
}

func (h browserHandler) getFolderAndPathFromURL(w http.ResponseWriter, r *http.Request, user *users.User) (*folders.Folder, string, bool) {
	// For the path "/browser/{{folderID}}/foo/bar/baz" the elems variable will have for content:
	// []string{"", "browser", "{{folderID}}", "/foo/bar/baz"}
	elems := strings.SplitN(r.URL.Path, "/", 4)

	// no need to check elems len as the url format force a len of 3 minimum
	folderID, err := h.uuid.Parse(elems[2])
	if err != nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
		return nil, "", true
	}

	folder, err := h.folders.GetUserFolder(r.Context(), user.ID(), folderID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to folders.GetByID: %w", err))
		return nil, "", true
	}

	if folder == nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
		return nil, "", true
	}

	fullPath := "."
	if len(elems) == 4 {
		fullPath = path.Clean(elems[3])
	}

	return folder, fullPath, false
}

func (h *browserHandler) renderBrowserContent(w http.ResponseWriter, r *http.Request, user *users.User, fs dfs.FS, fullPath string) {
	inode, err := fs.Get(r.Context(), fullPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	folder := fs.Folder()

	if inode == nil {
		w.Header().Set("Location", path.Join("/browser/", string(folder.ID())))
		w.WriteHeader(http.StatusFound)
		return
	}

	dirContent := []inodes.INode{}
	if inode.IsDir() {
		dirContent, err = fs.ListDir(r.Context(), fullPath, &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      20,
		})
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to inodes.Readdir: %w", err))
			return
		}
	} else {
		file, err := h.files.Open(r.Context(), inode.ID())
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to files.Open: %w", err))
			return
		}
		defer file.Close()

		h.serveContent(w, r, inode, file)
		return
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

func (h *browserHandler) serveContent(w http.ResponseWriter, r *http.Request, inode *inodes.INode, file io.ReadSeeker) {
	w.Header().Add("Etag", inode.Checksum())
	w.Header().Add("Expires", time.Now().Add(365*24*time.Hour).UTC().Format(http.TimeFormat))
	w.Header().Add("Cache-Control", "max-age=31536000")
	http.ServeContent(w, r, inode.Name(), inode.ModTime(), file)
}
