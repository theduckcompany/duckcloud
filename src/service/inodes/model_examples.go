package inodes

import (
	"io/fs"
	"time"

	"github.com/Peltoche/neurone/src/tools/uuid"
)

var (
	now  time.Time = time.Now()
	now2 time.Time = time.Now().Add(time.Minute)
)

var ExampleRoot INode = INode{
	id:             uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	name:           "",
	userID:         uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	parent:         NoParent,
	mode:           0o660 | fs.ModeDir,
	createdAt:      now,
	lastModifiedAt: now2,
}
