// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package webdav provides a WebDAV server implementation.
package webdav // import "github.com/theduckcompany/duckcloud/internal/service/dav/webdav"

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

type webdavKeyCtx string

var SessionKeyCtx webdavKeyCtx = "user"

type Handler struct {
	// Prefix is the URL path prefix to strip from WebDAV resource paths.
	Prefix string
	// FileSystem is the virtual file system.
	FileSystem dfs.Service
	// Sessions handle the users sessions used for authentification.
	Sessions davsessions.Service
	Spaces   spaces.Service
	Users    users.Service
	Files    files.Service
	// Logger is an optional error logger. If non-nil, it will be called
	// for all HTTP requests.
	Logger func(*http.Request, error)
}

func (h *Handler) stripPrefix(p string) (string, int, error) {
	p = path.Clean(strings.TrimSuffix(p, "/"))

	var r string

	if h.Prefix != "" {
		r = strings.TrimPrefix(p, h.Prefix)
		if len(r) == len(p) {
			return p, http.StatusNotFound, errPrefixMismatch
		}
	}

	if p == "." || r == "" {
		r = "/"
	}

	return r, http.StatusOK, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Add("WWW-Authenticate", `Basic realm="fs"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	session, err := h.Sessions.Authenticate(r.Context(), username, secret.NewText(password))
	if errors.Is(err, davsessions.ErrInvalidCredentials) {
		w.Header().Add("WWW-Authenticate", `Basic realm="fs"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := h.Users.GetByID(r.Context(), session.UserID())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	space, err := h.Spaces.GetUserSpace(r.Context(), session.UserID(), session.SpaceID())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	status, err := http.StatusBadRequest, errUnsupportedMethod

	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		w.WriteHeader(status)
		w.Write([]byte(StatusText(status)))
		return
	}

	pathCmd := &dfs.PathCmd{Space: space, Path: reqPath}

	switch {
	case h.FileSystem == nil:
		status, err = http.StatusInternalServerError, errNoFileSystem
	default:
		switch r.Method {
		case "OPTIONS":
			status, err = h.handleOptions(w, r, space)
		case "GET", "HEAD", "POST":
			status, err = h.handleGetHeadPost(w, r, pathCmd)
		case "DELETE":
			status, err = h.handleDelete(w, r, space)
		case "PUT":
			status, err = h.handlePut(w, r, user, space)
		case "MKCOL":
			status, err = h.handleMkcol(w, r, user, space)
		case "COPY", "MOVE":
			status, err = h.handleCopyMove(w, r, user, space)
		case "PROPFIND":
			status, err = h.handlePropfind(w, r, pathCmd)
		case "PROPPATCH":
			status, err = h.handleProppatch(w, r, space)
		}
	}

	if status != 0 {
		w.WriteHeader(status)
		if status != http.StatusNoContent {
			w.Write([]byte(StatusText(status)))
		}
	}
	if h.Logger != nil {
		h.Logger(r, err)
	}
}

func (h *Handler) handleOptions(w http.ResponseWriter, r *http.Request, space *spaces.Space) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	ctx := r.Context()
	allow := "OPTIONS, PUT, MKCOL"
	if fi, err := h.FileSystem.Get(ctx, &dfs.PathCmd{Space: space, Path: reqPath}); err == nil {
		if fi.IsDir() {
			allow = "OPTIONS, DELETE, PROPPATCH, COPY, MOVE, PROPFIND"
		} else {
			allow = "OPTIONS, GET, HEAD, POST, DELETE, PROPPATCH, COPY, MOVE, PROPFIND, PUT"
		}
	}
	w.Header().Set("Allow", allow)
	// http://www.webdav.org/specs/rfc4918.html#dav.compliance.classes
	w.Header().Set("DAV", "1")
	// http://msdn.microsoft.com/en-au/library/cc250217.aspx
	w.Header().Set("MS-Author-Via", "DAV")
	return 0, nil
}

func (h *Handler) handleGetHeadPost(w http.ResponseWriter, r *http.Request, pathCmd *dfs.PathCmd) (status int, err error) {
	// TODO: check locks for read-only access??
	ctx := r.Context()
	info, err := h.FileSystem.Get(ctx, pathCmd)
	if err != nil {
		return http.StatusNotFound, err
	}

	if info.IsDir() {
		return http.StatusMethodNotAllowed, nil
	}

	f, err := h.FileSystem.Download(ctx, pathCmd)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer f.Close()

	fileMetas, err := h.Files.GetMetadata(ctx, *info.FileID())
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to get the file metadatas: %w", err)
	}

	w.Header().Set("ETag", fmt.Sprintf("W/%q", fileMetas.Checksum()))
	w.Header().Set("Content-Type", fileMetas.MimeType())
	http.ServeContent(w, r, pathCmd.Path, info.LastModifiedAt(), f)
	return 0, nil
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request, space *spaces.Space) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}

	ctx := r.Context()

	// TODO: return MultiStatus where appropriate.

	// "godoc os RemoveAll" says that "If the path does not exist, RemoveAll
	// returns nil (no error)." WebDAV semantics are that it should return a
	// "404 Not Found". We therefore have to Stat before we RemoveAll.
	pathCmd := &dfs.PathCmd{Space: space, Path: reqPath}

	if _, err := h.FileSystem.Get(ctx, pathCmd); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return http.StatusNotFound, err
		}
		return http.StatusMethodNotAllowed, err
	}
	if err := h.FileSystem.Remove(ctx, pathCmd); err != nil {
		return http.StatusMethodNotAllowed, err
	}
	return http.StatusNoContent, nil
}

func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request, user *users.User, space *spaces.Space) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}

	// TODO(rost): Support the If-Match, If-None-Match headers? See bradfitz'
	// comments in http.checkEtag.
	ctx := r.Context()

	err = h.FileSystem.Upload(ctx, &dfs.UploadCmd{
		Space:      space,
		FilePath:   reqPath,
		Content:    r.Body,
		UploadedBy: user,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// TODO(peltoche): Put back Etag again?

	return http.StatusCreated, nil
}

func (h *Handler) handleMkcol(w http.ResponseWriter, r *http.Request, user *users.User, space *spaces.Space) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}

	ctx := r.Context()

	if r.ContentLength > 0 {
		return http.StatusUnsupportedMediaType, nil
	}

	// No file or directory with the same name should already exists
	res, err := h.FileSystem.Get(ctx, &dfs.PathCmd{Space: space, Path: reqPath})
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return http.StatusInternalServerError, err
	}

	if res != nil {
		return http.StatusMethodNotAllowed, nil
	}

	// All the parents must exists
	if reqPath != "/" {
		parent, err := h.FileSystem.Get(ctx, &dfs.PathCmd{Space: space, Path: path.Dir(reqPath)})
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			return http.StatusInternalServerError, err
		}

		if parent == nil {
			return http.StatusConflict, nil
		}
	}

	if _, err := h.FileSystem.CreateDir(ctx, &dfs.CreateDirCmd{
		Space:     space,
		FilePath:  reqPath,
		CreatedBy: user,
	}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return http.StatusConflict, err
		}
		return http.StatusMethodNotAllowed, err
	}
	return http.StatusCreated, nil
}

func (h *Handler) handleCopyMove(w http.ResponseWriter, r *http.Request, user *users.User, space *spaces.Space) (status int, err error) {
	hdr := r.Header.Get("Destination")
	if hdr == "" {
		return http.StatusBadRequest, errInvalidDestination
	}
	u, err := url.Parse(hdr)
	if err != nil {
		return http.StatusBadRequest, errInvalidDestination
	}
	if u.Host != "" && u.Host != r.Host {
		return http.StatusBadGateway, errInvalidDestination
	}

	src, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}
	srcPath := &dfs.PathCmd{Space: space, Path: src}

	dst, status, err := h.stripPrefix(u.Path)
	if err != nil {
		return status, err
	}
	dstPath := &dfs.PathCmd{Space: space, Path: dst}

	if dst == "" {
		return http.StatusBadGateway, errInvalidDestination
	}
	if dst == src {
		return http.StatusForbidden, errDestinationEqualsSource
	}

	ctx := r.Context()

	_, err = h.FileSystem.Get(ctx, srcPath)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return http.StatusConflict, err
		}
		return http.StatusInternalServerError, err
	}

	if r.Method == "COPY" {
		// Section 9.8.3 says that "The COPY method on a collection without a Depth
		// header must act as if a Depth header with value "infinity" was included".
		depth := infiniteDepth
		if hdr := r.Header.Get("Depth"); hdr != "" {
			depth = parseDepth(hdr)
			if depth != 0 && depth != infiniteDepth {
				// Section 9.8.3 says that "A client may submit a Depth header on a
				// COPY on a collection with a value of "0" or "infinity"."
				return http.StatusBadRequest, errInvalidDepth
			}
		}
		return copyFiles(ctx, user, h.FileSystem, srcPath, dstPath, r.Header.Get("Overwrite") != "F", depth, 0)
	}

	// Section 9.9.2 says that "The MOVE method on a collection must act as if
	// a "Depth: infinity" header was used on it. A client must not submit a
	// Depth header on a MOVE on a collection with any value but "infinity"."
	if hdr := r.Header.Get("Depth"); hdr != "" {
		if parseDepth(hdr) != infiniteDepth {
			return http.StatusBadRequest, errInvalidDepth
		}
	}

	dstInfo, err := h.FileSystem.Get(ctx, &dfs.PathCmd{Space: space, Path: dst})
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return http.StatusInternalServerError, err
	}

	if r.Header.Get("Overwrite") == "F" && dstInfo != nil {
		return http.StatusPreconditionFailed, nil
	}

	err = h.FileSystem.Move(ctx, &dfs.MoveCmd{
		Src: &dfs.PathCmd{
			Space: space,
			Path:  src,
		},
		Dst: &dfs.PathCmd{
			Space: space,
			Path:  dst,
		},
		MovedBy: user,
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if dstInfo != nil {
		return http.StatusNoContent, nil
	}

	return http.StatusCreated, nil
}

func (h *Handler) handlePropfind(w http.ResponseWriter, r *http.Request, cmd *dfs.PathCmd) (status int, err error) {
	ctx := r.Context()
	fi, err := h.FileSystem.Get(ctx, cmd)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return http.StatusNotFound, err
		}
		return http.StatusMethodNotAllowed, err
	}
	depth := infiniteDepth
	if hdr := r.Header.Get("Depth"); hdr != "" {
		depth = parseDepth(hdr)
		if depth == invalidDepth {
			return http.StatusBadRequest, errInvalidDepth
		}
	}
	pf, status, err := readPropfind(r.Body)
	if err != nil {
		return status, err
	}

	mw := multistatusWriter{w: w}

	walkFn := func(cmd *dfs.PathCmd, info *dfs.INode, err error) error {
		if err != nil {
			return handlePropfindError(err, info)
		}

		var fileMeta *files.FileMeta
		if info.FileID() != nil {
			// TODO: Log the error?
			fileMeta, _ = h.Files.GetMetadata(ctx, *info.FileID())
		}

		var pstats []Propstat
		switch {
		case pf.Propname != nil:
			pnames, err := propnames(ctx, info, cmd)
			if err != nil {
				return handlePropfindError(err, info)
			}
			pstat := Propstat{Status: http.StatusOK}
			for _, xmlname := range pnames {
				pstat.Props = append(pstat.Props, Property{XMLName: xmlname})
			}
			pstats = append(pstats, pstat)
		case pf.Allprop != nil:
			pstats, err = allprop(ctx, info, fileMeta, cmd, pf.Prop)
		default:
			pstats, err = props(ctx, info, fileMeta, cmd, pf.Prop)
		}
		if err != nil {
			return handlePropfindError(err, info)
		}
		href := path.Join(h.Prefix, cmd.Path)
		if href != "/" && info.IsDir() {
			href += "/"
		}
		return mw.write(makePropstatResponse(href, pstats))
	}

	walkErr := walkFS(ctx, h.FileSystem, depth, cmd, fi, walkFn)
	closeErr := mw.close()
	if walkErr != nil {
		return http.StatusInternalServerError, walkErr
	}
	if closeErr != nil {
		return http.StatusInternalServerError, closeErr
	}
	return 0, nil
}

func (h *Handler) handleProppatch(w http.ResponseWriter, r *http.Request, space *spaces.Space) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path)
	if err != nil {
		return status, err
	}

	ctx := r.Context()

	if _, err := h.FileSystem.Get(ctx, &dfs.PathCmd{Space: space, Path: reqPath}); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return http.StatusNotFound, err
		}
		return http.StatusMethodNotAllowed, err
	}
	patches, status, err := readProppatch(r.Body)
	if err != nil {
		return status, err
	}
	pstats, err := patch(ctx, h.FileSystem, reqPath, patches)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	mw := multistatusWriter{w: w}
	writeErr := mw.write(makePropstatResponse(r.URL.Path, pstats))
	closeErr := mw.close()
	if writeErr != nil {
		return http.StatusInternalServerError, writeErr
	}
	if closeErr != nil {
		return http.StatusInternalServerError, closeErr
	}
	return 0, nil
}

func makePropstatResponse(href string, pstats []Propstat) *response {
	resp := response{
		Href:     []string{(&url.URL{Path: href}).EscapedPath()},
		Propstat: make([]propstat, 0, len(pstats)),
	}
	for _, p := range pstats {
		var xmlErr *xmlError
		if p.XMLError != "" {
			xmlErr = &xmlError{InnerXML: []byte(p.XMLError)}
		}
		resp.Propstat = append(resp.Propstat, propstat{
			Status:              fmt.Sprintf("HTTP/1.1 %d %s", p.Status, StatusText(p.Status)),
			Prop:                p.Props,
			ResponseDescription: p.ResponseDescription,
			Error:               xmlErr,
		})
	}
	return &resp
}

func handlePropfindError(err error, info *dfs.INode) error {
	var skipResp error = nil
	if info != nil && info.IsDir() {
		skipResp = filepath.SkipDir
	}

	if errors.Is(err, os.ErrPermission) {
		// If the server cannot recurse into a directory because it is not allowed,
		// then there is nothing more to say about it. Just skip sending anything.
		return skipResp
	}

	if _, ok := err.(*os.PathError); ok {
		// If the file is just bad, it couldn't be a proper WebDAV resource. Skip it.
		return skipResp
	}

	// We need to be careful with other errors: there is no way to abort the xml stream
	// part way through while returning a valid PROPFIND response. Returning only half
	// the data would be misleading, but so would be returning results tainted by errors.
	// The current behaviour by returning an error here leads to the stream being aborted,
	// and the parent http server complaining about writing a spurious header. We should
	// consider further enhancing this error handling to more gracefully fail, or perhaps
	// buffer the entire response until we've walked the tree.
	return err
}

const (
	infiniteDepth = -1
	invalidDepth  = -2
)

// parseDepth maps the strings "0", "1" and "infinity" to 0, 1 and
// infiniteDepth. Parsing any other string returns invalidDepth.
//
// Different WebDAV methods have further constraints on valid depths:
//   - PROPFIND has no further restrictions, as per section 9.1.
//   - COPY accepts only "0" or "infinity", as per section 9.8.3.
//   - MOVE accepts only "infinity", as per section 9.9.2.
//   - LOCK accepts only "0" or "infinity", as per section 9.10.3.
//
// These constraints are enforced by the handleXxx methods.
func parseDepth(s string) int {
	switch s {
	case "0":
		return 0
	case "1":
		return 1
	case "infinity":
		return infiniteDepth
	}
	return invalidDepth
}

// http://www.webdav.org/specs/rfc4918.html#status.code.extensions.to.http11
const (
	StatusMulti               = 207
	StatusUnprocessableEntity = 422
	StatusLocked              = 423
	StatusFailedDependency    = 424
	StatusInsufficientStorage = 507
)

func StatusText(code int) string {
	switch code {
	case StatusMulti:
		return "Multi-Status"
	case StatusUnprocessableEntity:
		return "Unprocessable Entity"
	case StatusLocked:
		return "Locked"
	case StatusFailedDependency:
		return "Failed Dependency"
	case StatusInsufficientStorage:
		return "Insufficient Storage"
	}
	return http.StatusText(code)
}

var (
	errDestinationEqualsSource = errors.New("webdav: destination equals source")
	errInvalidDepth            = errors.New("webdav: invalid depth")
	errInvalidDestination      = errors.New("webdav: invalid destination")
	errInvalidLockInfo         = errors.New("webdav: invalid lock info")
	errInvalidPropfind         = errors.New("webdav: invalid propfind")
	errInvalidProppatch        = errors.New("webdav: invalid proppatch")
	errInvalidResponse         = errors.New("webdav: invalid response")
	errNoFileSystem            = errors.New("webdav: no file system")
	errPrefixMismatch          = errors.New("webdav: prefix mismatch")
	errRecursionTooDeep        = errors.New("webdav: recursion too deep")
	errUnsupportedLockInfo     = errors.New("webdav: unsupported lock info")
	errUnsupportedMethod       = errors.New("webdav: unsupported method")
)

// slashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

type WalkFunc func(cmd *dfs.PathCmd, info *dfs.INode, err error) error

// walkFS traverses filesystem fs starting at name up to depth levels.
//
// Allowed values for depth are 0, 1 or infiniteDepth. For each visited node,
// walkFS calls walkFn. If a visited file system node is a directory and
// walkFn returns filepath.SkipDir, walkFS will skip traversal of this node.
func walkFS(ctx context.Context, fs dfs.Service, depth int, cmd *dfs.PathCmd, info *dfs.INode, walkFn WalkFunc) error {
	// This implementation is based on Walk's code in the standard path/filepath package.
	err := walkFn(cmd, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsDir() || depth == 0 {
		return nil
	}
	if depth == 1 {
		depth = 0
	}

	// Read directory names.
	fileInfos, err := fs.ListDir(ctx, cmd, nil)
	if err != nil {
		return walkFn(cmd, info, err)
	}

	for _, fileInfo := range fileInfos {
		newPath := &dfs.PathCmd{Space: cmd.Space, Path: path.Join(cmd.Path, fileInfo.Name())}
		err = walkFS(ctx, fs, depth, newPath, &fileInfo, walkFn)
		if err != nil {
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}
