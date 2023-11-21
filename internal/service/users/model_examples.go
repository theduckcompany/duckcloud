package users

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var now = time.Now().UTC()

var ExampleAlice = User{
	id:             uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:       "Alice",
	defaultSpaceID: uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	isAdmin:        true,
	status:         Active,
	password:       secret.NewText("alice-encrypted-password"),
	createdAt:      now,
}

var ExampleBob = User{
	id:             uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	username:       "Bob",
	defaultSpaceID: uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	isAdmin:        false,
	status:         Active,
	password:       secret.NewText("bob-encrypted-password"),
	createdAt:      now,
}

var ExampleInitializingAlice = User{
	id:             uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:       "Alice",
	defaultSpaceID: "",
	isAdmin:        true,
	status:         Initializing,
	password:       secret.NewText("alice-encrypted-password"),
	createdAt:      now,
}

var ExampleDeletingAlice = User{
	id:             uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:       "Alice",
	defaultSpaceID: uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	isAdmin:        true,
	status:         Deleting,
	password:       secret.NewText("alice-encrypted-password"),
	createdAt:      now,
}
