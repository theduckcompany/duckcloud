package inodes

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	now  = time.Now().UTC()
	now2 = time.Now().Add(time.Minute).UTC()
)

var ExampleAliceRoot INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "",
	parent:         nil,
	checksum:       "",
	mimetype:       nil,
	size:           0,
	createdAt:      now,
	lastModifiedAt: now2,
	fileID:         nil,
}

var ExampleAliceDir INode = INode{
	id:             uuid.UUID("5592dac5-55f4-4206-87ca-8f31fc05e506"),
	name:           "dir-a",
	parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
	checksum:       "",
	mimetype:       nil,
	size:           42,
	createdAt:      now,
	lastModifiedAt: now2,
	fileID:         nil,
}

var ExampleAliceFile2 INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "file.txt",
	parent:         ptr.To(uuid.UUID("5592dac5-55f4-4206-87ca-8f31fc05e506")),
	mimetype:       ptr.To("application/pdf"),
	checksum:       "some-sha256-checksum",
	size:           42,
	createdAt:      now,
	lastModifiedAt: now2,
	fileID:         ptr.To(uuid.UUID("abf05a02-8af9-4184-a46d-847f7d951c6b")),
}

var ExampleAliceFile INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "foo",
	parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
	mimetype:       ptr.To("application/pdf"),
	checksum:       "some-sha256-checksum",
	size:           42,
	createdAt:      now,
	lastModifiedAt: now2,
	fileID:         ptr.To(uuid.UUID("abf05a02-8af9-4184-a46d-847f7d951c6b")),
}

var ExampleBobRoot INode = INode{
	id:             uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	name:           "",
	parent:         nil,
	mimetype:       nil,
	checksum:       "",
	size:           0,
	createdAt:      now,
	lastModifiedAt: now2,
	fileID:         nil,
}
