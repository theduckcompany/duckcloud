package inodes

import (
	"io/fs"
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	now        = time.Now()
	now2       = time.Now().Add(time.Minute)
	someFileID = uuid.UUID("3f3d0cab-341a-497c-ac65-cf6cb69e7142")
)

var ExampleRoot INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "",
	userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	fileID:         nil,
	parent:         NoParent,
	mode:           0o660 | fs.ModeDir,
	createdAt:      now,
	lastModifiedAt: now,
}

var ExampleFile INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "foo",
	userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	fileID:         &someFileID,
	parent:         uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	mode:           0o660,
	createdAt:      now2,
	lastModifiedAt: now2,
}
