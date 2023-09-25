package inodes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/ptr"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestINodes(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateDir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now()
		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "some-dir-name",
			parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			isDir:          true,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/some-dir-name",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("CreateDir success 2", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now()
		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "bar",
			parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			isDir:          true,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			isDir:  true,
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("CreateDir with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     "some-invalid-root",
			FullName: "/some-dir-name",
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: Root: must be a valid UUID v4.")
	})

	t.Run("CreateDir with a parent not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "unknown", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/unknown/some-dir-name", // invalid path
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "mkdir /unknown/some-dir-name: file does not exist")
	})

	t.Run("CreateDir with a file as child", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			isDir:  false, // File and not directory here <-,
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.EqualError(t, err, "mkdir /foo/bar: invalid argument")
		assert.Nil(t, res)
	})

	t.Run("CreateDir with a already existing file/directory", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now()

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "some-dir-name", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleBobRoot, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()

		res, err := service.CreateDir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/some-dir-name",
		})

		assert.EqualError(t, err, "mkdir /some-dir-name: file already exists")
		assert.Nil(t, res)
	})

	t.Run("GetByID success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()

		res, err := service.GetByID(ctx, ExampleAliceRoot.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceRoot, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		invalidID := uuid.UUID("f092f39a-1b5b-488c-8679-75607e798502")

		storageMock.On("GetByID", mock.Anything, invalidID).Return(nil, nil).Once()

		res, err := service.GetByID(ctx, invalidID)
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("RemoveAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now().UTC()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			isDir:  true,
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		storageMock.On("Patch", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"), map[string]any{"deleted_at": now}).Return(nil).Once()

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		})

		assert.NoError(t, err)
	})

	t.Run("RemoveAll with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     "some-invalid-id",
			FullName: "/foo",
		})

		assert.EqualError(t, err, "validation error: Root: must be a valid UUID v4.")
	})

	t.Run("RemoveAll with a file not found", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, nil).Once()

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		})

		assert.NoError(t, err)
	})

	t.Run("RemoveAll with a storage error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		now := time.Now().UTC()
		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			isDir:  true,
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		storageMock.On("Patch", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"), map[string]any{"deleted_at": now}).Return(errors.New("some-error")).Once()

		err := service.RemoveAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		})

		assert.EqualError(t, err, "some-error")
	})

	t.Run("Get success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		inode := INode{
			id:     uuid.UUID("eec51147-ec64-4640-b148-aceadbcb876e"),
			parent: ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			isDir:  false,
			name:   "bar",
			// some other unused fields
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			isDir:  true,
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "bar", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&inode, nil).Once()

		res, err := service.Get(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &inode, res)
	})

	t.Run("Get with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.Get(ctx, &PathCmd{
			Root:     "some-invalid-id",
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: Root: must be a valid UUID v4.")
	})

	t.Run("Get with an invalid root", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		res, err := service.Get(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "not found: root \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\" not found")
	})

	t.Run("GetAllDeleted success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetAllDeleted", mock.Anything, 10).Return([]INode{ExampleAliceRoot}, nil).Once()

		res, err := service.GetAllDeleted(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, ExampleAliceRoot, res[0])
	})

	t.Run("HardDelete success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("HardDelete", mock.Anything, ExampleAliceFile.ID()).Return(nil).Once()

		err := service.HardDelete(ctx, ExampleAliceFile.ID())
		assert.NoError(t, err)
	})

	t.Run("Readdir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		paginateCmd := storage.PaginateCmd{Limit: 2}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			isDir:  true,
			name:   "bar",
			// some other unused fields
		}, nil).Once()

		child1 := INode{
			id:     uuid.UUID("b3411c4b-acc3-4f79-a54e-f315a18ce6c7"),
			parent: ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			isDir:  true,
			name:   "some-dir",
		}
		child2 := INode{
			id:     uuid.UUID("0af1f541-454e-4c7d-a871-706d9c5ad2cc"),
			parent: ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
			isDir:  false,
			name:   "some-file",
		}

		storageMock.On("GetAllChildrens", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"), &paginateCmd).Return(
			[]INode{child1, child2}, nil).Once()

		res, err := service.Readdir(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo",
		}, &paginateCmd)

		assert.NoError(t, err)
		assert.Len(t, res, 2)
		assert.Equal(t, child1, res[0])
		assert.Equal(t, child2, res[1])
	})

	t.Run("Readdir with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.Readdir(ctx, &PathCmd{
			Root:     "some-invalid-id",
			FullName: "/some-dir-name",
		}, &storage.PaginateCmd{Limit: 10})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: Root: must be a valid UUID v4.")
	})

	t.Run("CreateFile success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		inode := INode{
			id:             uuid.UUID("some-id"),
			parent:         ptr.To(ExampleAliceRoot.ID()),
			name:           "foobar",
			isDir:          false,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()

		tools.ClockMock.On("Now").Return(now).Once()

		tools.UUIDMock.On("New").Return(uuid.UUID("some-id")).Once()

		storageMock.On("Save", mock.Anything, &inode).Return(nil).Once()

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Parent: ExampleAliceRoot.ID(),
			Name:   "foobar",
		})

		assert.NoError(t, err)
		assert.Equal(t, &inode, res)
	})

	t.Run("CreateFile with a validation error", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Parent: "some-invalid-id",
			Name:   "foobar",
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "validation error: Parent: must be a valid UUID v4.")
	})

	t.Run("CreateFile with a non existing parent", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(nil, nil).Once()

		res, err := service.CreateFile(ctx, &CreateFileCmd{
			Parent: ExampleAliceRoot.ID(),
			Name:   "foobar",
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "invalid parent: parent \"f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f\" not found")
	})

	t.Run("Get with an invalid path", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&INode{
			id:     uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
			parent: ptr.To(ExampleAliceRoot.ID()),
			isDir:  false, // Should be a directory with a "bar" as child
			name:   "foo",
			// some other unused fields
		}, nil).Once()

		res, err := service.Get(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "open /foo/bar: invalid argument")
	})

	t.Run("CreateRootDir success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		inode := &INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "",
			parent:         nil,
			isDir:          true,
			createdAt:      now,
			lastModifiedAt: now,
		}

		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")).Once()
		storageMock.On("Save", mock.Anything, inode).Return(nil).Once()

		res, err := service.CreateRootDir(ctx)

		assert.NoError(t, err)
		assert.EqualValues(t, inode, res)
	})

	t.Run("RegisterWrite success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		t.Cleanup(func() {
			ExampleAliceFile.lastModifiedAt = now2
		})

		now := time.Now()
		hash := sha256.New()
		n, err := hash.Write([]byte("some-content"))
		require.NoError(t, err)

		tools.ClockMock.On("Now").Return(now).Once()
		storageMock.On("Patch", mock.Anything, ExampleAliceFile.id, map[string]any{
			"checksum":         hex.EncodeToString(hash.Sum(nil)),
			"last_modified_at": now,
			"size":             ExampleAliceFile.size + uint64(n),
		}).Return(nil).Once()

		err = service.RegisterWrite(ctx, &ExampleAliceFile, n, hash)
		assert.NoError(t, err)
	})

	t.Run("GetINodeRoot success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, *ExampleAliceFile.parent).Return(&ExampleAliceRoot, nil).Once()

		res, err := service.GetINodeRoot(ctx, &ExampleAliceFile)
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceRoot, res)
	})

	t.Run("GetINodeRoot with a root", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		// There is not call to GetByID because there is no parent.

		res, err := service.GetINodeRoot(ctx, &ExampleAliceRoot)
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAliceRoot, res)
	})

	t.Run("MkdirAll success", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now().UTC()
		fooDir := INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "foo",
			parent:         ptr.To(ExampleAliceRoot.ID()),
			isDir:          true,
			createdAt:      now,
			lastModifiedAt: now,
		}
		barDir := INode{
			id:             uuid.UUID("1afc4ef3-d0e8-4efe-8e37-4d23acc5df9c"),
			name:           "bar",
			parent:         ptr.To(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")),
			isDir:          true,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		// Check if the folder /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(nil, nil).Once()

		// CreateDir("/foo") internal
		// Check if the folder already exists
		storageMock.On("GetByID", mock.Anything, ExampleAliceRoot.ID()).Return(&ExampleAliceRoot, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(nil, nil).Once()
		// Generate an save ""/foo/bar"" folder
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(fooDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &fooDir).Return(nil).Once()

		// Check if the folder /foo and /foo/bar exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, nil).Once()

		// CreateDir("/foo/bar") internal
		// Check if the folder already exists
		storageMock.On("GetByID", mock.Anything, fooDir.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, nil).Once()
		// Generate an save ""/foo/bar"" folder
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(barDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &barDir).Return(nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &barDir, res)
	})

	t.Run("MkdirAll with a folder already existing", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		now := time.Now().UTC()
		fooDir := INode{
			id:             uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62"),
			name:           "foo",
			parent:         ptr.To(ExampleAliceRoot.ID()),
			isDir:          true,
			createdAt:      now,
			lastModifiedAt: now,
		}
		barDir := INode{
			id:             uuid.UUID("1afc4ef3-d0e8-4efe-8e37-4d23acc5df9c"),
			name:           "bar",
			parent:         ptr.To(uuid.UUID("976246a7-ed3e-4556-af48-1fed703e7a62")),
			isDir:          true,
			createdAt:      now,
			lastModifiedAt: now,
		}

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		// Check if the folder /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&fooDir, nil).Once()

		// Check if the folder /foo and /foo/bar exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", ExampleAliceRoot.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, nil).Once()

		// CreateDir("/foo/bar") internal
		// Check if the folder already exists
		storageMock.On("GetByID", mock.Anything, fooDir.ID()).Return(&fooDir, nil).Once()
		storageMock.On("GetByNameAndParent", mock.Anything, "bar", fooDir.ID()).Return(nil, nil).Once()
		// Generate an save ""/foo/bar"" folder
		tools.ClockMock.On("Now").Return(now).Once()
		tools.UUIDMock.On("New").Return(barDir.ID()).Once()
		storageMock.On("Save", mock.Anything, &barDir).Return(nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.NoError(t, err)
		assert.EqualValues(t, &barDir, res)
	})

	t.Run("MkdirAll with a an existing file", func(t *testing.T) {
		tools := tools.NewMock(t)
		storageMock := NewMockStorage(t)
		service := NewService(tools, storageMock)

		storageMock.On("GetByID", mock.Anything, uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceRoot, nil).Once()

		// Check if the folder /foo exists
		storageMock.On("GetByNameAndParent", mock.Anything, "foo", uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")).Return(&ExampleAliceFile, nil).Once()

		res, err := service.MkdirAll(ctx, &PathCmd{
			Root:     ExampleAliceRoot.ID(),
			FullName: "/foo/bar",
		})

		assert.Nil(t, res)
		assert.EqualError(t, err, "mkdir /foo/bar: not a directory")
	})
}
