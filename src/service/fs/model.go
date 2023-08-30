package fs

import (
	"io"
	"io/fs"
	"net/http"
)

type FileOrDirectory interface {
	http.File
	io.Writer
	fs.ReadDirFile
}
