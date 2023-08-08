package inodes

import (
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
)

var now time.Time = time.Now()

var ExampleRoot INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "",
	userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	parent:         NoParent,
	nodeType:       Directory,
	createdAt:      now,
	lastModifiedAt: now,
}
