package files

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var now = time.Now().UTC()

var ExampleFile1 = FileMeta{
	id:         uuid.UUID("abf05a02-8af9-4184-a46d-847f7d951c6b"),
	size:       42,
	mimetype:   "text/plain; charset=utf-8",
	checksum:   "3eWunOpspQ2soXv6HoPRiQ0HFoXeSMShH6SlEgIg1mM=",
	uploadedAt: now,
}

var ExampleFile2 = FileMeta{
	id:         uuid.UUID("66278d2b-7a4f-4764-ac8a-fc08f224eb66"),
	size:       22,
	mimetype:   "text/plain; charset=utf-8",
	checksum:   "9NiDSp5zOcgEDl+j00MP/WkGVPOlPRLejGD8Ga6PJ7M=",
	uploadedAt: now,
}
