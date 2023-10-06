package uploads

import (
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var now = time.Now().UTC()

var ExampleAliceUpload = Upload{
	id:         uuid.UUID("3155de86-01f0-42b0-b839-89e58159f4ba"),
	folderID:   folders.ExampleAlicePersonalFolder.ID(),
	dir:        inodes.ExampleAliceRoot.ID(),
	fileName:   "foo.txt",
	fileID:     *inodes.ExampleAliceFile.FileID(),
	uploadedAt: now,
}
