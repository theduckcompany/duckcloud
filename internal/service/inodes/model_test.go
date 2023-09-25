package inodes

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestInodeGetter(t *testing.T) {
	assert.Equal(t, ExampleAliceRoot.ID(), uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"))
	assert.Equal(t, ExampleAliceRoot.Name(), "")
	assert.Nil(t, ExampleAliceRoot.Parent())
	assert.Equal(t, ExampleAliceRoot.Mode(), 0o660|fs.ModeDir)
	assert.Equal(t, ExampleAliceRoot.CreatedAt(), now)
	assert.Equal(t, ExampleAliceRoot.LastModifiedAt(), now2)
	assert.Equal(t, ExampleAliceRoot.Size(), int64(0))

	assert.Equal(t, ExampleAliceFile.Size(), int64(42))
	assert.Equal(t, ExampleAliceFile.Parent(), ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")))
}
