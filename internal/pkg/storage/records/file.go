package records

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type FileMode int

const (
	FILE_MODE_UNKNOWN FileMode = 0
	FILE_MODE_READ    FileMode = 1
	FILE_MODE_WRITE   FileMode = 2
)

type OpenFile interface {
	io.ReadWriteSeeker
	io.Closer

	Mode() FileMode

	HandleID() string
	Flush() error

	WrittenHash() string
	WrittenSize() uint64
}

type FileIterator interface {
	Files() <-chan *FileItem
	Close() error
}

type FileItem struct {
	File  *File
	Error error
}

type File struct {
	ID           uuid.UUID
	Hash         string
	Filename     string
	DocumentDate time.Time
	Size         uint64
}

type FileMetadata struct {
	ID       uuid.UUID
	Hash     string
	FileSize uint64
}
