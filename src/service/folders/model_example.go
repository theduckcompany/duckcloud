package folders

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var now time.Time = time.Now().UTC()

var ExampleAlicePersonalFolder = Folder{
	id:        uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	name:      "Alice's Folder",
	owners:    Owners{"86bffce3-3f53-4631-baf8-8530773884f3"},
	rootFS:    uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	createdAt: now,
}

var ExampleBobPersonalFolder = Folder{
	id:        uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	name:      "Bob's Folder",
	owners:    Owners{"0923c86c-24b6-4b9d-9050-e82b8408edf4"},
	rootFS:    uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	createdAt: now,
}
