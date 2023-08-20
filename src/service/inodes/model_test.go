package inodes

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/ptr"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestInodeGetter(t *testing.T) {
	assert.Equal(t, ExampleRoot.ID(), uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"))
	assert.Equal(t, ExampleRoot.Name(), "")
	assert.Equal(t, ExampleRoot.UserID(), uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"))
	assert.Nil(t, ExampleRoot.Parent())
	assert.Equal(t, ExampleRoot.Mode(), 0o660|fs.ModeDir)
	assert.Equal(t, ExampleRoot.CreatedAt(), now)
	assert.Equal(t, ExampleRoot.LastModifiedAt(), now)
	assert.Equal(t, ExampleRoot.Size(), int64(0))

	assert.Equal(t, ExampleFile.Size(), int64(42))
	assert.Equal(t, ExampleFile.Parent(), ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")))
}
