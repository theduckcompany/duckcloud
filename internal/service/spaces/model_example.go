package spaces

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var now time.Time = time.Now().UTC()

var ExampleAlicePersonalSpace = Space{
	id:        uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	name:      "Alice's Space",
	owners:    Owners{"86bffce3-3f53-4631-baf8-8530773884f3"},
	createdAt: now,
	createdBy: users.ExampleAlice.ID(),
}

var ExampleBobPersonalSpace = Space{
	id:        uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	name:      "Bob's Space",
	owners:    Owners{"0923c86c-24b6-4b9d-9050-e82b8408edf4"},
	createdAt: now,
	createdBy: users.ExampleBob.ID(),
}

var ExampleAliceBobSharedSpace = Space{
	id:        uuid.UUID("c8943050-6bc5-4641-a4ba-672c1f03b4cd"),
	name:      "Alice and Bob Space",
	owners:    Owners{"86bffce3-3f53-4631-baf8-8530773884f3", "0923c86c-24b6-4b9d-9050-e82b8408edf4"},
	createdAt: now,
	createdBy: users.ExampleAlice.ID(),
}
