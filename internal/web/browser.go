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
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
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

var ErrInvalidSpaceID = errors.New("invalid spaceID")

type browserHandler struct {
	html   html.Writer
	spaces spaces.Service
	files  files.Service
	uuid   uuid.Service
	auth   *Authenticator
	fs     dfs.Service
}

func newBrowserHandler(
	tools tools.Tools,
	html html.Writer,
	spaces spaces.Service,
	files files.Service,
	auth *Authenticator,
	fs dfs.Service,
) *browserHandler {
	return &browserHandler{
		html:   html,
		spaces: spaces,
		files:  files,
		uuid:   tools.UUID(),
		auth:   auth,
		fs:     fs,
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

	r.Get("/browser/rename", h.getRenameModal)
	r.Post("/browser/rename", h.handleRenameReq)

	r.Get("/browser/create-dir", h.getCreateDirModal)
	r.Post("/browser/create-dir", h.handleCreateDirReq)

	r.Delete("/browser/*", h.deleteAll)
}

func (h *browserHandler) String() string {
	return "web.browser"
}

func (h *browserHandler) handleRenameReq(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	space, filePath, abort := h.getRenameParams(w, r)
	if abort {
		return
	}

	fs := h.fs.GetSpaceFS(space)

	inode, err := fs.Get(ctx, filePath)
	if errors.Is(err, errs.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err = fs.Rename(ctx, inode, r.FormValue("name"))
	if errors.Is(err, errs.ErrValidation) {
		h.renderRenameModal(w, r, &renameCmd{
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

	h.renderBrowserContent(w, r, user, fs, path.Dir(filePath))
}

func (h *browserHandler) getRenameParams(w http.ResponseWriter, r *http.Request) (*spaces.Space, string, bool) {
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

func (h *browserHandler) getRenameModal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	space, filePath, abort := h.getRenameParams(w, r)
	if abort {
		return
	}

	fs := h.fs.GetSpaceFS(space)

	_, err := fs.Get(ctx, filePath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the file %q: %w", filePath, err))
	}

	h.renderRenameModal(w, r, &renameCmd{
		ErrorMsg: "",
		Space:    space,
		Path:     filePath,
		Value:    path.Base(filePath),
	})
}

type renameCmd struct {
	ErrorMsg string
	Space    *spaces.Space
	Value    string
	Path     string
}

func (h *browserHandler) renderRenameModal(w http.ResponseWriter, r *http.Request, cmd *renameCmd) {
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

	spaceID := r.URL.Query().Get("space")
	if spaceID == "" {
		h.html.WriteHTMLErrorPage(w, r, errors.New("failed to get the space id from the url query"))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/create-dir.tmpl", map[string]interface{}{
		"directory": dir,
		"spaceID":   spaceID,
	})
}

func (h *browserHandler) handleCreateDirReq(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
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
		h.html.WriteHTML(w, r, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]interface{}{
			"directory": dir,
			"spaceID":   spaceID,
			"error":     "Must not be empty",
		})
		return
	}

	space, err := h.spaces.GetUserSpace(r.Context(), user.ID(), spaceID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, errs.BadRequest(fmt.Errorf("failed to GetUserSpace: %w", err)))
		return
	}

	fs := h.fs.GetSpaceFS(space)

	existingDir, err := fs.Get(r.Context(), path.Join(dir, name))
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the directory: %w", err))
		return
	}

	if existingDir != nil {
		h.html.WriteHTML(w, r, http.StatusUnprocessableEntity, "browser/create-dir.tmpl", map[string]interface{}{
			"directory": dir,
			"spaceID":   spaceID,
			"error":     "Already exists",
		})
		return
	}

	_, err = fs.CreateDir(r.Context(), &dfs.CreateDirCmd{
		FilePath:  path.Join(dir, name),
		CreatedBy: user,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to create the directory: %w", err))
		return
	}

	h.renderBrowserContent(w, r, user, fs, dir)
}

func (h *browserHandler) redirectDefaultBrowser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	spaces, err := h.spaces.GetAllUserSpaces(ctx, user.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to get the user spaces: %w", err))
	}

	if len(spaces) == 0 {
		h.html.WriteHTMLErrorPage(w, r, errors.New("user have zero spaces"))
		return
	}

	spaceID := spaces[0].ID()

	w.Header().Set("Location", path.Join("/browser/", string(spaceID)))
	w.WriteHeader(http.StatusFound)
}

func (h *browserHandler) getBrowserContent(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	space, fullPath, abort := h.getSpaceAndPathFromURL(w, r, user, r.URL.Path)
	if abort {
		return
	}
	fs := h.fs.GetSpaceFS(space)

	lastElem := r.URL.Query().Get("last")
	if lastElem == "" {
		h.renderBrowserContent(w, r, user, fs, fullPath)
		return
	}

	h.renderMoreDirContent(w, r, space, fullPath, lastElem)
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

	var name, rawSpaceID, rootPath, relPath []byte

	for {
		p, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if p.FileName() != "" {
			spaceID, err := h.uuid.Parse(string(rawSpaceID))
			if err != nil {
				h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("invalid space id %q", rawSpaceID))
				return
			}

			cmd := lauchUploadCmd{
				user:       user,
				name:       string(name),
				spaceID:    spaceID,
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
		case "spaceID":
			rawSpaceID, err = io.ReadAll(p)
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

	space, fullPath, abort := h.getSpaceAndPathFromURL(w, r, user, r.URL.Path)
	if abort {
		return
	}

	fs := h.fs.GetSpaceFS(space)
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

func generateBreadCrumb(space *spaces.Space, fullPath string) []breadCrumbElement {
	basePath := path.Join("/browser/", string(space.ID()))

	res := []breadCrumbElement{{
		Name:    space.Name(),
		Href:    basePath,
		Current: false,
	}}

	fullPath = strings.TrimPrefix(fullPath, "/")

	if fullPath == "" {
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
	spaceID    uuid.UUID
	rootPath   string
	relPath    string
	fileReader io.Reader
}

func (h *browserHandler) lauchUpload(ctx context.Context, cmd *lauchUploadCmd) error {
	space, err := h.spaces.GetUserSpace(ctx, cmd.user.ID(), cmd.spaceID)
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

	ffs := h.fs.GetSpaceFS(space)

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
	_, err = ffs.CreateDir(ctx, &dfs.CreateDirCmd{
		FilePath:  dirPath,
		CreatedBy: cmd.user,
	})
	if err != nil && !errors.Is(err, dfs.ErrAlreadyExists) {
		return fmt.Errorf("failed to create the directory %q: %w", dirPath, err)
	}

	err = ffs.Upload(ctx, &dfs.UploadCmd{
		FilePath:   fullPath,
		Content:    cmd.fileReader,
		UploadedBy: cmd.user,
	})
	if err != nil {
		return fmt.Errorf("failed to Upload file: %w", err)
	}

	return nil
}

func (h browserHandler) getSpaceAndPathFromURL(w http.ResponseWriter, r *http.Request, user *users.User, pathStr string) (*spaces.Space, string, bool) {
	pathStr = strings.TrimPrefix(pathStr, "/")        // Trim for the urls like: /space-id/foo/bar
	pathStr = strings.TrimPrefix(pathStr, "browser/") // Trim for the urls like: /browser/space-id/foo/bar

	// For the path "{{spaceID}}/foo/bar/baz" the elems variable will have for content:
	// []string{"{{spaceID}}", "/foo/bar/baz"}
	elems := strings.SplitN(pathStr, "/", 2)

	// no need to check elems len as the url format force a len of 3 minimum
	spaceID, err := h.uuid.Parse(elems[0])
	if err != nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
		return nil, "", true
	}

	space, err := h.spaces.GetUserSpace(r.Context(), user.ID(), spaceID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to spaces.GetByID: %w", err))
		return nil, "", true
	}

	if space == nil {
		w.Header().Set("Location", "/browser")
		w.WriteHeader(http.StatusFound)
		return nil, "", true
	}

	fullPath := "/"
	if len(elems) == 2 {
		fullPath = path.Clean("/" + elems[1])
	}

	return space, fullPath, false
}

func (h *browserHandler) renderBrowserContent(w http.ResponseWriter, r *http.Request, user *users.User, ffs dfs.FS, fullPath string) {
	inode, err := ffs.Get(r.Context(), fullPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	space := ffs.Space()

	if inode == nil {
		w.Header().Set("Location", path.Join("/browser/", string(space.ID())))
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

	spaces, err := h.spaces.GetAllUserSpaces(r.Context(), user.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetallUserSpaces: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
		"host":       r.Host,
		"fullPath":   fullPath,
		"space":      space,
		"breadcrumb": generateBreadCrumb(space, fullPath),
		"spaces":     spaces,
		"inodes":     dirContent,
	})
}

func (h *browserHandler) renderMoreDirContent(w http.ResponseWriter, r *http.Request, space *spaces.Space, fullPath, lastElem string) {
	ffs := h.fs.GetSpaceFS(space)
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
		"space":    space,
		"inodes":   dirContent,
	})
}

func (h *browserHandler) download(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.getUserAndSession(w, r, AnyUser)
	if abort {
		return
	}

	space, fullPath, abort := h.getSpaceAndPathFromURL(w, r, user, r.URL.Path)
	if abort {
		return
	}

	ffs := h.fs.GetSpaceFS(space)

	inode, err := ffs.Get(r.Context(), fullPath)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	if inode == nil {
		w.Header().Set("Location", path.Join("/browser/", string(space.ID())))
		w.WriteHeader(http.StatusFound)
		return
	}

	if inode.IsDir() {
		h.serveSpaceContent(w, r, ffs, fullPath)
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

func (h *browserHandler) serveSpaceContent(w http.ResponseWriter, r *http.Request, ffs dfs.FS, root string) {
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
