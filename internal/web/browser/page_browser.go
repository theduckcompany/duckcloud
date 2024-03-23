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
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/logger"
	"github.com/theduckcompany/duckcloud/internal/tools/router"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"github.com/theduckcompany/duckcloud/internal/web/auth"
	"github.com/theduckcompany/duckcloud/internal/web/html"
	"github.com/theduckcompany/duckcloud/internal/web/html/templates/browser"
	browsertmpl "github.com/theduckcompany/duckcloud/internal/web/html/templates/browser"
)

const (
	MaxMemoryCache = 20 * 1024 * 1024 // 20MB
	PageSize       = 30
)

var ErrInvalidSpaceID = errors.New("invalid spaceID")

type BrowserPage struct {
	html       html.Writer
	spaces     spaces.Service
	files      files.Service
	uuid       uuid.Service
	auth       *auth.Authenticator
	fs         dfs.Service
	uploadLock *sync.Mutex
}

func NewBrowserPage(
	tools tools.Tools,
	html html.Writer,
	spaces spaces.Service,
	files files.Service,
	auth *auth.Authenticator,
	fs dfs.Service,
) *BrowserPage {
	return &BrowserPage{
		html:       html,
		spaces:     spaces,
		files:      files,
		uuid:       tools.UUID(),
		auth:       auth,
		fs:         fs,
		uploadLock: new(sync.Mutex),
	}
}

func (h *BrowserPage) Register(r chi.Router, mids *router.Middlewares) {
	if mids != nil {
		r = r.With(mids.Defaults()...)
	}

	r.Get("/browser", h.redirectDefaultBrowser)
	r.Post("/browser/upload", h.upload)
	r.Get("/download/{spaceID}/*", h.download)
	r.Get("/browser/{spaceID}", h.getBrowserContent)
	r.Get("/browser/{spaceID}/*", h.getBrowserContent)
	r.Get("/browser/view/{spaceID}/*", h.showMedia)
	r.Delete("/browser/{spaceID}/*", h.deleteAll)

	newCreateDirModalHandler(h.auth, h.spaces, h.html, h.uuid, h.fs).Register(r, mids)
	newRenameModalHandler(h.auth, h.spaces, h.html, h.uuid, h.fs).Register(r, mids)
	newMoveModalHandler(h.auth, h.spaces, h.html, h.uuid, h.fs).Register(r, mids)
}

func (h *BrowserPage) redirectDefaultBrowser(w http.ResponseWriter, r *http.Request) {
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

	http.Redirect(w, r, path.Join("/browser/", string(spaceID)), http.StatusFound)
}

func (h *BrowserPage) getBrowserContent(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	path := h.getPathFromURL(w, r, user)
	if path == nil {
		return
	}

	lastElem := r.URL.Query().Get("last")
	if lastElem != "" {
		h.renderMoreDirContent(w, r, path, lastElem)
		return
	}

	h.renderBrowserContent(w, r, user, path)
}

func (h *BrowserPage) upload(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, io.EOF) {
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

func (h *BrowserPage) deleteAll(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	path := h.getPathFromURL(w, r, user)
	if path == nil {
		return
	}

	err := h.fs.Remove(r.Context(), path)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Remove: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

type lauchUploadCmd struct {
	fileReader io.Reader
	user       *users.User
	name       string
	spaceID    uuid.UUID
	rootPath   string
	relPath    string
}

func (h *BrowserPage) lauchUpload(ctx context.Context, cmd *lauchUploadCmd) error {
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

	// TODO: Replace this lock by a webdav lock implementation.
	h.uploadLock.Lock()

	dirPath := path.Dir(fullPath)
	_, err = h.fs.CreateDir(ctx, &dfs.CreateDirCmd{
		Path:      dfs.NewPathCmd(space, dirPath),
		CreatedBy: cmd.user,
	})
	if err != nil && !errors.Is(err, dfs.ErrAlreadyExists) {
		h.uploadLock.Unlock()
		return fmt.Errorf("failed to create the directory %q: %w", dirPath, err)
	}
	h.uploadLock.Unlock()

	err = h.fs.Upload(ctx, &dfs.UploadCmd{
		Path:       dfs.NewPathCmd(space, fullPath),
		Content:    cmd.fileReader,
		UploadedBy: cmd.user,
	})
	if err != nil {
		return fmt.Errorf("failed to Upload file: %w", err)
	}

	return nil
}

func (h BrowserPage) getPathFromURL(w http.ResponseWriter, r *http.Request, user *users.User) *dfs.PathCmd {
	// no need to check elems len as the url format force a len of 3 minimum
	spaceID, err := h.uuid.Parse(chi.URLParam(r, "spaceID"))
	if err != nil {
		http.Redirect(w, r, "/browser", http.StatusFound)
		return nil
	}

	space, err := h.spaces.GetUserSpace(r.Context(), user.ID(), spaceID)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to spaces.GetByID: %w", err))
		return nil
	}

	if space == nil {
		http.Redirect(w, r, "/browser", http.StatusFound)
		return nil
	}

	return dfs.NewPathCmd(space, chi.URLParam(r, "*"))
}

func (h *BrowserPage) renderBrowserContent(w http.ResponseWriter, r *http.Request, user *users.User, cmd *dfs.PathCmd) {
	inode, err := h.fs.Get(r.Context(), cmd)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	if inode == nil {
		http.Redirect(w, r, path.Join("/browser/", string(cmd.Space().ID())), http.StatusFound)
		return
	}

	var dirContent []dfs.INode
	if inode.IsDir() {
		dirContent, err = h.fs.ListDir(r.Context(), cmd, &sqlstorage.PaginateCmd{
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

	w.Header().Set("HX-Push-Url", path.Join("/browser", string(cmd.Space().ID()), cmd.Path()))

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &browser.ContentTemplate{
		Folder:        cmd,
		Inodes:        dirContent,
		CurrentSpace:  cmd.Space(),
		AllSpaces:     spaces,
		ContentTarget: "body",
	})
}

func (h *BrowserPage) renderMoreDirContent(w http.ResponseWriter, r *http.Request, folderPath *dfs.PathCmd, lastElem string) {
	dirContent, err := h.fs.ListDir(r.Context(), folderPath, &sqlstorage.PaginateCmd{
		StartAfter: map[string]string{"name": lastElem},
		Limit:      PageSize,
	})
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to ListDir: %w", err))
		return
	}

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &browser.RowsTemplate{
		Inodes:        dirContent,
		Folder:        folderPath,
		ContentTarget: "body",
	})
}

func (h *BrowserPage) download(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	pathCmd := h.getPathFromURL(w, r, user)
	if pathCmd == nil {
		return
	}

	inode, err := h.fs.Get(r.Context(), pathCmd)
	if err != nil {
		h.html.WriteHTMLErrorPage(w, r, fmt.Errorf("failed to fs.Get: %w", err))
		return
	}

	if inode == nil {
		http.Redirect(w, r, path.Join("/browser/", string(pathCmd.Space().ID())), http.StatusFound)
		return
	}

	if inode.IsDir() {
		h.serveFolderContent(w, r, pathCmd)
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

func (h *BrowserPage) serveFolderContent(w http.ResponseWriter, r *http.Request, cmd *dfs.PathCmd) {
	var err error

	_, dir := path.Split(cmd.Path())

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

		header.Name, err = filepath.Rel(cmd.Path(), p)
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

		file, err := h.fs.Download(ctx, dfs.NewPathCmd(cmd.Space(), path.Join(p, i.Name())))
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

func (h *BrowserPage) showMedia(w http.ResponseWriter, r *http.Request) {
	user, _, abort := h.auth.GetUserAndSession(w, r, auth.AnyUser)
	if abort {
		return
	}

	filePath := h.getPathFromURL(w, r, user)
	if filePath == nil {
		return
	}

	folder, fileName := path.Split(filePath.Path())

	h.html.WriteHTMLTemplate(w, r, http.StatusOK, &browsertmpl.MediaViewerModal{
		Path:     filePath,
		FileName: fileName,
		Folder:   strings.TrimSuffix(folder, "/"),
	})
}
