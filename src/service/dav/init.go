package dav

import (
	"golang.org/x/net/webdav"
)

type Service = webdav.FileSystem

func Init() Service {
	return NewFSService()
}
