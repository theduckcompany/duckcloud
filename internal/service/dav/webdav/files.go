package webdav

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

// copyFiles copies files and/or directories from src to dst.
//
// See section 9.8.5 for when various HTTP status codes apply.
func copyFiles(ctx context.Context, fs dfs.FS, src, dst string, overwrite bool, depth int, recursion int) (status int, err error) {
	if recursion == 1000 {
		return http.StatusInternalServerError, errRecursionTooDeep
	}
	recursion++

	// TODO: section 9.8.3 says that "Note that an infinite-depth COPY of /A/
	// into /A/B/ could lead to infinite recursion if not handled correctly."

	srcStat, err := fs.Get(ctx, src)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return http.StatusNotFound, err
		}
		return http.StatusInternalServerError, err
	}

	created := false
	if _, err := fs.Get(ctx, dst); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			created = true
		} else {
			return http.StatusForbidden, err
		}

		_, err := fs.Get(ctx, path.Dir(dst))
		if err != nil && errors.Is(err, errs.ErrNotFound) && !overwrite {
			return http.StatusConflict, nil
		}
	} else {
		if !overwrite {
			return http.StatusPreconditionFailed, os.ErrExist
		}
		if err := fs.Remove(ctx, dst); err != nil && !errors.Is(err, errs.ErrNotFound) {
			return http.StatusForbidden, err
		}
	}

	if srcStat.IsDir() {
		if _, err := fs.CreateDir(ctx, dst); err != nil {
			return http.StatusForbidden, err
		}
		if depth == infiniteDepth {
			children, err := fs.ListDir(ctx, src, nil)
			if err != nil {
				return http.StatusForbidden, err
			}
			for _, c := range children {
				name := c.Name()
				s := path.Join(src, name)
				d := path.Join(dst, name)
				cStatus, cErr := copyFiles(ctx, fs, s, d, overwrite, depth, recursion)
				if cErr != nil {
					// TODO: MultiStatus.
					return cStatus, cErr
				}
			}
		}
	} else {
		reader, err := fs.Download(ctx, src)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				return http.StatusConflict, err
			}

			return http.StatusInternalServerError, err
		}
		defer reader.Close()

		err = fs.Upload(ctx, dst, reader)
		if err != nil {
			return http.StatusForbidden, err
		}
	}

	if created {
		return http.StatusCreated, nil
	}
	return http.StatusNoContent, nil
}