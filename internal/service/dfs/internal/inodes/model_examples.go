package inodes

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
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
	size:           0,
	createdAt:      now,
	createdBy:      users.ExampleAlice.ID(),
	lastModifiedAt: now2,
	fileID:         nil,
}

var ExampleAliceDir INode = INode{
	id:             uuid.UUID("5592dac5-55f4-4206-87ca-8f31fc05e506"),
	name:           "dir-a",
	parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
	spaceID:        spaces.ExampleAlicePersonalSpace.ID(),
	size:           42,
	createdAt:      now,
	createdBy:      users.ExampleAlice.ID(),
	lastModifiedAt: now2,
	fileID:         nil,
}

var ExampleAliceFile2 INode = INode{
	id:             uuid.UUID("733da385-5cb8-4851-af12-927c252b6c1f "),
	name:           "file.txt",
	parent:         ptr.To(uuid.UUID("5592dac5-55f4-4206-87ca-8f31fc05e506")),
	spaceID:        spaces.ExampleAlicePersonalSpace.ID(),
	size:           42,
	createdAt:      now,
	createdBy:      users.ExampleAlice.ID(),
	lastModifiedAt: now2,
	fileID:         ptr.To(uuid.UUID("2007688f-022f-4d4a-a704-e86d66070227")),
}

var ExampleAliceFile INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "foo",
	parent:         ptr.To(uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f")),
	spaceID:        spaces.ExampleAlicePersonalSpace.ID(),
	size:           42,
	createdAt:      now,
	createdBy:      users.ExampleAlice.ID(),
	lastModifiedAt: now2,
	fileID:         ptr.To(files.ExampleFile1.ID()),
}

var ExampleBobRoot INode = INode{
	id:             uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	name:           "",
	parent:         nil,
	spaceID:        spaces.ExampleBobPersonalSpace.ID(),
	size:           0,
	createdAt:      now,
	createdBy:      users.ExampleBob.ID(),
	lastModifiedAt: now2,
	fileID:         nil,
}
