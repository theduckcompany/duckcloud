package fs

import (
	"io"
	"net/http"
)

type File interface {
	http.File
	io.Writer
}
