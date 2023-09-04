package inodes

import (
	"io/fs"
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/ptr"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
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
	mode:           0o660 | fs.ModeDir,
	size:           0,
	createdAt:      now,
	lastModifiedAt: now2,
}

var ExampleAliceFile INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "foo",
	parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
	checksum:       "some-sha256-checksum",
	size:           42,
	mode:           0o660,
	createdAt:      now,
	lastModifiedAt: now2,
}

var ExampleBobRoot INode = INode{
	id:             uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	name:           "",
	parent:         nil,
	checksum:       "",
	mode:           0o660 | fs.ModeDir,
	size:           0,
	createdAt:      now,
	lastModifiedAt: now2,
}
