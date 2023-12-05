package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

const WalkBatchSize = 100

type WalkDirFunc func(ctx context.Context, path string, i *INode) error

func Walk(ctx context.Context, ffs Service, cmd *PathCmd, fn WalkDirFunc) error {
	err := cmd.Validate()
	if err != nil {
		return errs.Validation(err)
	}

	inode, err := ffs.Get(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to Get a file: %w", err)
	}

	if !inode.IsDir() {
		return fn(ctx, cmd.Path, inode)
	}

	err = fn(ctx, cmd.Path, inode)
	if err != nil {
		return err
	}

	lastOffset := ""
	for {
		dirContent, err := ffs.ListDir(ctx, cmd, &storage.PaginateCmd{
			StartAfter: map[string]string{"name": lastOffset},
			Limit:      WalkBatchSize,
		})
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to ListDir %q: %w", cmd.Path, err)
		}

		for _, elem := range dirContent {
			err = Walk(ctx, ffs, &PathCmd{Space: cmd.Space, Path: path.Join(cmd.Path, elem.Name())}, fn)
			if err != nil {
				return err
			}
		}

		if len(dirContent) > 0 {
			lastOffset = dirContent[len(dirContent)-1].Name()
		}

		if len(dirContent) < WalkBatchSize {
			break
		}
	}

	return nil
}

// CleanPath is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func CleanPath(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
