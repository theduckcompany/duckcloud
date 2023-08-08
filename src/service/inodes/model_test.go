package inodes

import (
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInodeGetter(t *testing.T) {
	now := time.Now()
	now2 := time.Now()

	inode := INode{
		id:             uuid.UUID("some-id"),
		name:           "test",
		userID:         uuid.UUID("some-user-id"),
		parent:         NoParent,
		nodeType:       Directory,
		createdAt:      now,
		lastModifiedAt: now2,
	}

	assert.Equal(t, inode.ID(), uuid.UUID("some-id"))
	assert.Equal(t, inode.Name(), "test")
	assert.Equal(t, inode.UserID(), uuid.UUID("some-user-id"))
	assert.Equal(t, inode.Parent(), NoParent)
	assert.Equal(t, inode.Type(), Directory)
	assert.Equal(t, inode.CreatedAt(), now)
	assert.Equal(t, inode.LastModifiedAt(), now2)
	assert.Equal(t, inode.ModTime(), now2)
}
