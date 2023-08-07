package dav

import (
	"github.com/Peltoche/neurone/src/service/inodes"
	"golang.org/x/net/webdav"
)

type Service = webdav.FileSystem

func Init(inodes inodes.Service) Service {
	return NewFSService(inodes)
}
