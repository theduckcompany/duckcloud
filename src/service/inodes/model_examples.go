package inodes

import (
	"io/fs"
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/ptr"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	now  time.Time = time.Now()
	now2 time.Time = time.Now().Add(time.Minute)
)

var ExampleAliceRoot INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "",
	userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	parent:         nil,
	mode:           0o660 | fs.ModeDir,
	size:           0,
	createdAt:      now,
	lastModifiedAt: now,
}

var ExampleAliceFile INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "foo",
	userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
	size:           42,
	mode:           0o660,
	createdAt:      now2,
	lastModifiedAt: now2,
}

var ExampleBobRoot INode = INode{
	id:             uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	name:           "",
	userID:         uuid.UUID("49f06ad8-a7c2-4e21-b8c1-60d56dc83842"),
	parent:         nil,
	mode:           0o660 | fs.ModeDir,
	size:           0,
	createdAt:      now,
	lastModifiedAt: now,
}
