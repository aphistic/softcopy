package vault

import (
	"os"
	"time"
)

type dirFileInfo struct {
	name    string
	modTime time.Time
	mode    os.FileMode
	size    int64
	sys     interface{}
}

var _ os.FileInfo = &dirFileInfo{}

func (dfi *dirFileInfo) Name() string {
	return dfi.name
}

func (dfi *dirFileInfo) IsDir() bool {
	return true
}

func (dfi *dirFileInfo) ModTime() time.Time {
	return dfi.modTime
}

func (dfi *dirFileInfo) Mode() os.FileMode {
	return dfi.mode
}

func (dfi *dirFileInfo) Size() int64 {
	return dfi.size
}

func (dfi *dirFileInfo) Sys() interface{} {
	return dfi.sys
}
