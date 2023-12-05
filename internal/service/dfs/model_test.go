package dfs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestInodeGetter(t *testing.T) {
	assert.Equal(t, ExampleAliceRoot.ID(), uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"))
	assert.Equal(t, ExampleAliceRoot.Name(), "")
	assert.Nil(t, ExampleAliceRoot.Parent())
	assert.Nil(t, ExampleAliceRoot.FileID())
	assert.Equal(t, ExampleAliceRoot.CreatedAt(), now)
	assert.Equal(t, ExampleAliceRoot.CreatedBy(), users.ExampleAlice.ID())
	assert.Equal(t, ExampleAliceRoot.LastModifiedAt(), now2)
	assert.True(t, ExampleAliceRoot.IsDir())

	assert.Equal(t, ExampleAliceFile.Size(), uint64(42))
	assert.False(t, ExampleAliceFile.IsDir())
	assert.Equal(t, ExampleAliceFile.FileID(), ptr.To(uuid.UUID("abf05a02-8af9-4184-a46d-847f7d951c6b")))
	assert.Equal(t, ExampleAliceFile.Parent(), ExampleAliceFile.parent)
	assert.Equal(t, ExampleAliceFile.SpaceID(), ExampleAliceFile.spaceID)
}

func Test_Inodes_Commands(t *testing.T) {
	t.Run("CreateFileCmd", func(t *testing.T) {
		cmd := CreateFileCmd{
			Space:      nil, // invalid
			Parent:     &ExampleAliceRoot,
			Name:       "Foobar",
			File:       &files.ExampleFile1,
			UploadedAt: now,
			UploadedBy: &users.ExampleAlice,
		}

		err := cmd.Validate()
		assert.EqualError(t, err, "Space: cannot be blank.")
	})

	t.Run("CreateRootDirCmd", func(t *testing.T) {
		cmd := CreateRootDirCmd{
			CreatedBy: nil,
			Space:     &spaces.ExampleAlicePersonalSpace,
		}

		err := cmd.Validate()
		assert.EqualError(t, err, "CreatedBy: cannot be blank.")
	})
}
