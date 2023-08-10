package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"golang.org/x/exp/slog"
)

const gcBatchSize = 10

type GCService struct {
	inodes inodes.Service
	log    *slog.Logger
	cancel context.CancelFunc
	quit   chan struct{}
}

func NewGCService(inodes inodes.Service, tools tools.Tools) *GCService {
	return &GCService{inodes, tools.Logger(), nil, make(chan struct{})}
}

func (s *GCService) Start(pauseDuration time.Duration) {
	ticker := time.NewTicker(pauseDuration)
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		for {
			select {
			case <-ticker.C:
				err := s.run(ctx)
				if err != nil {
					s.log.Error("fs gc error", slog.String("error", err.Error()))
				}
			case <-s.quit:
				ticker.Stop()
				cancel()
			}
		}
	}()
}

func (s *GCService) Stop() {
	close(s.quit)

	if s.cancel != nil {
		s.cancel()
	}
}

func (s *GCService) run(ctx context.Context) error {
	for {
		toDelete, err := s.inodes.GetDeletedINodes(ctx, gcBatchSize)
		if err != nil {
			return fmt.Errorf("failed to GetDeletedINodes: %w", err)
		}

		for _, inode := range toDelete {
			err = s.deleteINode(ctx, &inode)
			if err != nil {
				return fmt.Errorf("failed to delete inode %q: %w", inode.ID(), err)
			}

			s.log.DebugCtx(ctx, "inode successfully removed", slog.String("inode", string(inode.ID())))
		}

		if len(toDelete) < gcBatchSize {
			return nil
		}
	}
}

func (s *GCService) deleteINode(ctx context.Context, inode *inodes.INode) error {
	if inode.Mode().IsDir() {
		childs, err := s.inodes.Readdir(ctx, &inodes.PathCmd{
			Root:     inode.ID(),
			UserID:   inode.UserID(),
			FullName: "/",
		}, &storage.PaginateCmd{Limit: 10})
		if err != nil {
			return fmt.Errorf("failed to Readdir: %w", err)
		}

		for _, child := range childs {
			err = s.deleteINode(ctx, &child)
			if err != nil {
				return fmt.Errorf("failed to deleteINode %q: %w", child.ID(), err)
			}
		}
	}

	return s.inodes.HardDelete(ctx, inode.ID())
}
