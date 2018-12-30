package records

import (
	"time"

	"github.com/google/uuid"
)

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
	Size         int64
}
