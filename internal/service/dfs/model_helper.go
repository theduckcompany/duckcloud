package dfs

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

type FakeINodeBuilder struct {
	t     *testing.T
	inode *INode
}

func NewFakeINode(t *testing.T) *FakeINodeBuilder {
	t.Helper()

	uuidProvider := uuid.NewProvider()
	createdAt := gofakeit.DateRange(time.Now().Add(-time.Hour*1000), time.Now())

	return &FakeINodeBuilder{
		t: t,
		inode: &INode{
			id:             uuidProvider.New(),
			createdAt:      createdAt,
			lastModifiedAt: createdAt,
			parent:         ptr.To(uuidProvider.New()),
			fileID:         ptr.To(uuidProvider.New()),
			name:           gofakeit.BeerName(),
			spaceID:        uuidProvider.New(),
			createdBy:      uuidProvider.New(),
			size:           42,
		},
	}
}

func (f *FakeINodeBuilder) IsRootDirectory() *FakeINodeBuilder {
	f.inode.parent = nil
	f.inode.fileID = nil

	return f
}

// func (f *FakeINodeBuilder) IsDirectory() *FakeINodeBuilder {
// 	f.inode.fileID = nil
//
// 	return f
// }

func (f *FakeINodeBuilder) WithSpace(space *spaces.Space) *FakeINodeBuilder {
	f.inode.spaceID = space.ID()

	return f
}

func (f *FakeINodeBuilder) WithParent(inode *INode) *FakeINodeBuilder {
	f.inode.parent = ptr.To(inode.ID())

	return f
}

func (f *FakeINodeBuilder) WithFile(file *files.FileMeta) *FakeINodeBuilder {
	f.inode.size = file.Size()
	f.inode.fileID = ptr.To(file.ID())

	return f
}

func (f *FakeINodeBuilder) WithID(id string) *FakeINodeBuilder {
	f.inode.id = uuid.UUID(id)

	return f
}

func (f *FakeINodeBuilder) WithName(name string) *FakeINodeBuilder {
	f.inode.name = name

	return f
}

func (f *FakeINodeBuilder) CreatedBy(user *users.User) *FakeINodeBuilder {
	f.inode.createdBy = user.ID()

	return f
}

func (f *FakeINodeBuilder) CreatedAt(t time.Time) *FakeINodeBuilder {
	f.inode.createdAt = t
	f.inode.lastModifiedAt = t

	return f
}

func (f *FakeINodeBuilder) Build() *INode {
	return f.inode
}

func (f *FakeINodeBuilder) BuildAndStore(ctx context.Context, db sqlstorage.Querier) *INode {
	f.t.Helper()

	storage := newSqlStorage(db)

	err := storage.Save(ctx, f.inode)
	require.NoError(f.t, err)

	return f.inode
}
