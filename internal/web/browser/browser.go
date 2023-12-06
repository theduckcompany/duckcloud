package browser

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
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/browser"
)

const (
	MaxMemoryCache = 20 * 1024 * 1024 // 20MB
	PageSize       = 50
)

var ErrInvalidSpaceID = errors.New("invalid spaceID")

type Handler struct {
	html   html.Writer
	spaces spaces.Service
	files  files.Service
	uuid   uuid.Service
	auth   *auth.Authenticator
	fs     dfs.Service
}

func NewHandler(
	tools tools.Tools,
	html html.Writer,
	spaces spaces.Service,
	files files.Service,
	auth *auth.Authenticator,
	fs dfs.Service,
) *Handler {
	return &Handler{
		html:   html,
		spaces: spaces,
		files:  files,
		uuid:   tools.UUID(),
		auth:   auth,
		fs:     fs,
	}
}

func (h *Handler) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.RealIP, mids.StripSlashed, mids.Logger)
	}

	r.Get("/browser", h.redirectDefaultBrowser)
	r.Post("/browser/upload", h.upload)
	r.Get("/download/*", h.download)
	r.Get("/browser/*", h.getBrowserContent)

	r.Delete("/browser/*", h.deleteAll)

	newCreateDirModalHandler(h.auth, h.spaces, h.html, h.uuid, h.fs).Register(r, mids)
	newRenameModalHandler(h.auth, h.spaces, h.html, h.uuid, h.fs).Register(r, mids)
}

func (h *Handler) String() string {
	return "web.browser"
}

func (h *Handler) redirectDefaultBrowser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
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

func (h *Handler) getBrowserContent(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	space, fullPath, abort := h.getSpaceAndPathFromURL(w, r, user, r.URL.Path)
	if abort {
		return
	}

	lastElem := r.URL.Query().Get("last")
	if lastElem == "" {
		h.renderBrowserContent(w, r, user, &dfs.PathCmd{Space: space, Path: fullPath})
		return
	}

	h.renderMoreDirContent(w, r, space, fullPath, lastElem)
}

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
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

func (h *Handler) deleteAll(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	space, fullPath, abort := h.getSpaceAndPathFromURL(w, r, user, r.URL.Path)
	if abort {
		return
	}

	err := h.fs.Remove(r.Context(), &dfs.PathCmd{Space: space, Path: fullPath})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Remove: %w", err))
		return
	}

	w.Header().Add("HX-Trigger", "refreshFolder")
	w.Header().Add("HX-Reswap", "none")
	w.WriteHeader(http.StatusNoContent)
}

func generateBreadCrumb(cmd *dfs.PathCmd) *browser.BreadCrumbTemplate {
	basePath := path.Join("/browser/", string(cmd.Space.ID()))

	res := &browser.BreadCrumbTemplate{}

	res.Elements = []browser.BreadCrumbElement{{
		Name:    cmd.Space.Name(),
		Href:    basePath,
		Current: false,
	}}

	fullPath := strings.TrimPrefix(cmd.Path, "/")

	if fullPath == "" {
		res.Elements[0].Current = true
		return res
	}

	for _, elem := range strings.Split(fullPath, "/") {
		basePath = path.Join(basePath, elem)

		res.Elements = append(res.Elements, browser.BreadCrumbElement{
			Name:    elem,
			Href:    basePath,
			Current: false,
		})
	}

	res.Elements[len(res.Elements)-1].Current = true

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

func (h *Handler) lauchUpload(ctx context.Context, cmd *lauchUploadCmd) error {
	space, err := h.spaces.GetUserSpace(ctx, cmd.user.ID(), cmd.spaceID)
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

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
	_, err = h.fs.CreateDir(ctx, &dfs.CreateDirCmd{
		Space:     space,
		FilePath:  dirPath,
		CreatedBy: cmd.user,
	})
	if err != nil && !errors.Is(err, dfs.ErrAlreadyExists) {
		return fmt.Errorf("failed to create the directory %q: %w", dirPath, err)
	}

	err = h.fs.Upload(ctx, &dfs.UploadCmd{
		Space:      space,
		FilePath:   fullPath,
		Content:    cmd.fileReader,
		UploadedBy: cmd.user,
	})
	if err != nil {
		return fmt.Errorf("failed to Upload file: %w", err)
	}

	return nil
}

func (h Handler) getSpaceAndPathFromURL(w http.ResponseWriter, r *http.Request, user *users.User, pathStr string) (*spaces.Space, string, bool) {
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

func (h *Handler) renderBrowserContent(w http.ResponseWriter, r *http.Request, user *users.User, cmd *dfs.PathCmd) {
	inode, err := h.fs.Get(r.Context(), cmd)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	if inode == nil {
		w.Header().Set("Location", path.Join("/browser/", string(cmd.Space.ID())))
		w.WriteHeader(http.StatusFound)
		return
	}

	dirContent := []dfs.INode{}
	if inode.IsDir() {
		dirContent, err = h.fs.ListDir(r.Context(), cmd, &storage.PaginateCmd{
			StartAfter: map[string]string{"name": ""},
			Limit:      PageSize,
		})
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to inodes.Readdir: %w", err))
			return
		}
	} else {
		fileMeta, _ := h.files.GetMetadata(r.Context(), *inode.FileID())
		file, err := h.fs.Download(r.Context(), cmd)
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Download: %w", err))
			return
		}
		defer file.Close()

		serveContent(w, r, inode, file, fileMeta)
		return
	}

	spaces, err := h.spaces.GetAllUserSpaces(r.Context(), user.ID(), nil)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to GetallUserSpaces: %w", err))
		return
	}

	h.html.WriteHTML(w, r, http.StatusOK, "browser/content.tmpl", map[string]interface{}{
		"host":       r.Host,
		"fullPath":   cmd.Path,
		"space":      cmd.Space,
		"Breadcrumb": generateBreadCrumb(cmd),
		"spaces":     spaces,
		"RowsContent": &browser.RowsTemplate{
			Folder: cmd,
			Inodes: dirContent,
		},
	})
}

func (h *Handler) renderMoreDirContent(w http.ResponseWriter, r *http.Request, space *spaces.Space, fullPath, lastElem string) {
	dirContent, err := h.fs.ListDir(r.Context(), &dfs.PathCmd{Space: space, Path: fullPath}, &storage.PaginateCmd{
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

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	space, fullPath, abort := h.getSpaceAndPathFromURL(w, r, user, r.URL.Path)
	if abort {
		return
	}

	pathCmd := &dfs.PathCmd{Space: space, Path: fullPath}

	inode, err := h.fs.Get(r.Context(), pathCmd)
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
		h.serveFolderContent(w, r, &dfs.PathCmd{Space: space, Path: fullPath})
	} else {
		fileMeta, _ := h.files.GetMetadata(r.Context(), *inode.FileID())

		file, err := h.fs.Download(r.Context(), pathCmd)
		if err != nil {
			h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to Download: %w", err))
			return
		}
		defer file.Close()

		serveContent(w, r, inode, file, fileMeta)
		return
	}
}

func (h *Handler) serveFolderContent(w http.ResponseWriter, r *http.Request, cmd *dfs.PathCmd) {
	var err error

	_, dir := path.Split(cmd.Path)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", dir))
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

	writer := zip.NewWriter(w)

	dfs.Walk(r.Context(), h.fs, cmd, func(ctx context.Context, p string, i *dfs.INode) error {
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

		header.Name, err = filepath.Rel(cmd.Path, p)
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

		file, err := h.fs.Download(ctx, &dfs.PathCmd{Space: cmd.Space, Path: path.Join(p, i.Name())})
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
