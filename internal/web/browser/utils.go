package browser

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/files"
)

func serveContent(w http.ResponseWriter, r *http.Request, inode *dfs.INode, file io.ReadSeeker, fileMeta *files.FileMeta) {
	if fileMeta != nil {
		w.Header().Set("ETag", fmt.Sprintf("W/%q", fileMeta.Checksum()))
		w.Header().Set("Content-Type", fileMeta.MimeType())
	}

	w.Header().Set("Expires", time.Now().Add(365*24*time.Hour).UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "max-age=31536000")

	http.ServeContent(w, r, inode.Name(), inode.LastModifiedAt(), file)
}
