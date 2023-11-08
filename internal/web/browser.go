package web

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/html"
)

const (
	MaxMemoryCache = 20 * 1024 * 1024 // 20MB
	PageSize       = 50
)

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
	r.Get("/download/*", h.download)

	r.Get("/browser/create-dir", h.getCreateDirModal)
	r.Post("/browser/create-dir", h.handleCreateDirReq)

	r.Delete("/browser/*", h.deleteAll)
}

func (h *browserHandler) String() string {
	return "web.browser"
}

func (h *browserHandler) getCreateDirModal(w http.ResponseWriter, r *http.Request) {
	_, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	dir := r.URL.Query().Get("dir")
	if dir == "" {
		h.html.WriteHTMLErrorPage(w, r, errors.New("failed to get the dir path from the url query"))
		return
	}

	folderID := r.URL.Query().Get("folder")
	if folderID == "" {
		h.html.WriteHTMLErrorPage(w, r, errors.New("failed to get the folder id from the url query"))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/create-dir.tmpl", map[string]interface{}{
		"directory": dir,
		"folderID":  folderID,
	})
}

func (h *browserHandler) handleCreateDirReq(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	dir := r.FormValue("dirPath")
	name := r.FormValue("name")
	folderID, err := h.uuid.Parse(r.FormValue("folderID"))
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, errors.New("invalid folder id param"))
		return
	}

	if name == "" {
		h.html.WriteHTML(w, r, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]interface{}{
			"directory": dir,
			"folderID":  folderID,
			"error":     "Must not be empty",
		})
		return
	}

	folder, err := h.folders.GetUserFolder(r.Context(), user.ID(), folderID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, errs.BadRequest(fmt.Errorf("failed to GetUserFolder: %w", err)))
		return
	}

	fs := h.fs.GetFolderFS(folder)

	existingDir, err := fs.Get(r.Context(), path.Join(dir, name))
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the directory: %w", err))
		return
	}

	if existingDir != nil {
		h.html.WriteHTML(w, r, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]interface{}{
			"directory": dir,
			"folderID":  folderID,
			"error":     "Already exists",
		})
		return
	}

	_, err = fs.CreateDir(r.Context(), path.Join(dir, name))
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to create the directory: %w", err))
		return
	}

	h.renderBrowserContent(w, r, user, fs, dir)
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

	lastElem := r.URL.Query().Get("last")
	if lastElem == "" {
		h.renderBrowserContent(w, r, user, fs, fullPath)
		return
	}

	h.renderMoreDirContent(w, r, folder, fullPath, lastElem)
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
				logger.LogEntrySetError(r, fmt.Errorf("upload error: %w", err))
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
	err := fs.Remove(r.Context(), fullPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Remove: %w", err))
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
	}

	if fullPath[0] == '/' {
		fullPath = fullPath[1:]
	}

	dirPath := path.Dir(fullPath)
	_, err = ffs.CreateDir(ctx, dirPath)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("failed to create the directory %q: %w", dirPath, err)
	}

	err = ffs.Upload(ctx, fullPath, cmd.fileReader)
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

func (h *browserHandler) renderBrowserContent(w http.ResponseWriter, r *http.Request, user *users.User, ffs dfs.FS, fullPath string) {
	inode, err := ffs.Get(r.Context(), fullPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	folder := ffs.Folder()

	if inode == nil {
		w.Header().Set("Location", path.Join("/browser/", string(folder.ID())))
		w.WriteHeader(http.StatusFound)
		return
	}

	dirContent := []dfs.INode{}
	if inode.IsDir() {
		dirContent, err = ffs.ListDir(r.Context(), fullPath, &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      PageSize,
		})
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to inodes.Readdir: %w", err))
			return
		}
	} else {
		fileMeta, _ := h.files.GetMetadata(r.Context(), *inode.FileID())
		file, err := ffs.Download(r.Context(), fullPath)
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Download: %w", err))
			return
		}
		defer file.Close()

		h.serveContent(w, r, inode, file, fileMeta)
		return
	}

	folders, err := h.folders.GetAllUserFolders(r.Context(), user.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetallUserFolders: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
		"host":       r.Host,
		"fullPath":   fullPath,
		"folder":     folder,
		"breadcrumb": generateBreadCrumb(folder, fullPath),
		"folders":    folders,
		"inodes":     dirContent,
	})
}

func (h *browserHandler) renderMoreDirContent(w http.ResponseWriter, r *http.Request, folder *folders.Folder, fullPath, lastElem string) {
	ffs := h.fs.GetFolderFS(folder)
	dirContent, err := ffs.ListDir(r.Context(), fullPath, &storage.PaginateCmd{
		StartAfter: map[string]string{"name": lastElem},
		Limit:      PageSize,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to ListDir: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/rows.tmpl", map[string]interface{}{
		"host":     r.Host,
		"fullPath": fullPath,
		"folder":   folder,
		"inodes":   dirContent,
	})
}

func (h *browserHandler) download(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	folder, fullPath, abort := h.getFolderAndPathFromURL(w, r, user)
	if abort {
		return
	}

	ffs := h.fs.GetFolderFS(folder)

	inode, err := ffs.Get(r.Context(), fullPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	if inode == nil {
		w.Header().Set("Location", path.Join("/browser/", string(folder.ID())))
		w.WriteHeader(http.StatusFound)
		return
	}

	if inode.IsDir() {
		h.serveFolderContent(w, r, ffs, fullPath)
	} else {
		fileMeta, _ := h.files.GetMetadata(r.Context(), *inode.FileID())

		file, err := ffs.Download(r.Context(), fullPath)
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Download: %w", err))
			return
		}
		defer file.Close()

		h.serveContent(w, r, inode, file, fileMeta)
		return
	}
}

func (h *browserHandler) serveFolderContent(w http.ResponseWriter, r *http.Request, ffs dfs.FS, root string) {
	var err error

	_, dir := path.Split(root)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", dir))
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	writer := zip.NewWriter(w)

	dfs.Walk(r.Context(), ffs, root, func(ctx context.Context, p string, i *dfs.INode) error {
		header := &zip.FileHeader{
			Method:             zip.Deflate,
			Comment:            "From DuckCloud with love",
			Name:               i.Name(),
			UncompressedSize64: i.Size(),
			Modified:           i.LastModifiedAt(),
		}

		if i.IsDir() {
			header.SetMode(0o755 | fs.ModeDir)
		} else {
			header.SetMode(0o644)
		}

		header.Name, err = filepath.Rel(root, p)
		if err != nil {
			return fmt.Errorf("failed to find the relative path: %w", err)
		}

		if i.IsDir() {
			header.Name += "/"
		}

		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create the zip header: %w", err)
		}

		if i.IsDir() {
			return nil
		}

		file, err := ffs.Download(ctx, path.Join(p, i.Name()))
		if err != nil {
			return fmt.Errorf("failed to download for zip: %w", err)
		}

		_, err = io.Copy(headerWriter, file)

		return err
	})

	err = writer.Close()
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Close the zip file: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *browserHandler) serveContent(w http.ResponseWriter, r *http.Request, inode *dfs.INode, file io.ReadSeeker, fileMeta *files.FileMeta) {
	if fileMeta != nil {
		w.Header().Set("Etag", fileMeta.Checksum())
		w.Header().Set("Content-Type", fileMeta.MimeType())
	}

	w.Header().Set("Expires", time.Now().Add(365*24*time.Hour).UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "max-age=31536000")

	http.ServeContent(w, r, inode.Name(), inode.LastModifiedAt(), file)
}
