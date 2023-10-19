package dfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"

	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

const WalkBatchSize = 100

type WalkDirFunc func(ctx context.Context, path string, i *INode) error

func Walk(ctx context.Context, ffs FS, root string, fn WalkDirFunc) error {
	if !fs.ValidPath(root) {
		return ErrInvalidPath
	}

	root = path.Clean(root)

	inode, err := ffs.Get(ctx, root)
	if err != nil {
		return fmt.Errorf("failed to Get a file: %w", err)
	}

	if !inode.IsDir() {
		return fn(ctx, root, inode)
	}

	err = fn(ctx, root, inode)
	if err != nil {
		return err
	}

	lastOffset := ""
	for {
		dirContent, err := ffs.ListDir(ctx, root, &storage.PaginateCmd{
			StartAfter: map[string]string{"name": lastOffset},
			Limit:      WalkBatchSize,
		})
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to ListDir %q: %w", root, err)
		}

		for _, elem := range dirContent {
			err = Walk(ctx, ffs, path.Join(root, elem.Name()), fn)
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
