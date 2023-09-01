package users

import (
	"time"

	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var now = time.Now().UTC()

var ExampleInitializingAlice = User{
	id:        uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:  "Alice",
	fsRoot:    "",
	isAdmin:   true,
	status:    "initializing",
	password:  "alice-encrypted-password",
	createdAt: now,
}

var ExampleAlice = User{
	id:        uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3"),
	username:  "Alice",
	fsRoot:    uuid.UUID("f5c0d3d2-e1b9-492b-b5d4-bd64bde0128f"),
	isAdmin:   true,
	status:    "active",
	password:  "alice-encrypted-password",
	createdAt: now,
}

var ExampleBob = User{
	id:        uuid.UUID("0923c86c-24b6-4b9d-9050-e82b8408edf4"),
	username:  "Bob",
	isAdmin:   false,
	fsRoot:    uuid.UUID("49f06ad8-a7c2-4e21-b8c1-60d56dc83842"),
	status:    "active",
	password:  "bob-encrypted-password",
	createdAt: now,
}
