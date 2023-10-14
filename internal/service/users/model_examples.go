package users

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var now = time.Now().UTC()

var ExampleAlice = User{
	id:              uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:        "Alice",
	defaultFolderID: uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	isAdmin:         true,
	status:          Active,
	password:        "alice-encrypted-password",
	createdAt:       now,
}

var ExampleBob = User{
	id:              uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	username:        "Bob",
	defaultFolderID: uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	isAdmin:         false,
	status:          Active,
	password:        "bob-encrypted-password",
	createdAt:       now,
}

var ExampleInitializingAlice = User{
	id:              uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:        "Alice",
	defaultFolderID: "",
	isAdmin:         true,
	status:          Initializing,
	password:        "alice-encrypted-password",
	createdAt:       now,
}

var ExampleDeletingAlice = User{
	id:              uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:        "Alice",
	defaultFolderID: uuid.UUID("e97b60f7-add2-43e1-a9bd-e2dac9ce69ec"),
	isAdmin:         true,
	status:          Deleting,
	password:        "alice-encrypted-password",
	createdAt:       now,
}
