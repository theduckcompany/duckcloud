package files

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var now = time.Now().UTC()

var ExampleSealedKey secret.SealedKey

//nolint:gochecknoinits // Don't have much choice here
func init() {
	masterKey, _ := secret.NewKey()
	key, _ := secret.NewKey()

	res, _ := secret.SealKey(masterKey, key)

	ExampleSealedKey = *res
}

var ExampleFile1 = FileMeta{
	id:         uuid.UUID("abf05a02-8af9-4184-a46d-847f7d951c6b"),
	size:       42,
	mimetype:   "text/plain; charset=utf-8",
	checksum:   "wGKmdG7y2opGyALNvIp9pmFCJXgoaQ2-3EMdM03ADKQ=",
	key:        &ExampleSealedKey,
	uploadedAt: now,
}

var ExampleFile2 = FileMeta{
	id:         uuid.UUID("66278d2b-7a4f-4764-ac8a-fc08f224eb66"),
	size:       22,
	mimetype:   "text/plain; charset=utf-8",
	key:        &ExampleSealedKey,
	checksum:   "SDoHdxhNbtfFu9ZN9PGKKc6wW1Dk1P3YJbU3LK-gehY=",
	uploadedAt: now,
}
