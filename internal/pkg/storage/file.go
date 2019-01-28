package storage

import (
	"io"

	"github.com/google/uuid"
)

type File interface {
	OpenFile(uuid.UUID) (OpenFile, error)
	OpenTempFile(uuid.UUID) (OpenFile, error)

	ReadFile(filePath string) (io.ReadCloser, error)
	ReadFileFromOffset(filePath string, offset uint64) (io.ReadCloser, error)
}

type OpenFile interface {
	io.ReadWriteSeeker

	Flush() error

	// Close closes a file opened for reading
	Close() error

	// Claim closes a file opened for writing and keeps it in
	// persistent storage with the given ID
	Claim(uuid.UUID) error
	// Drop closes a file opened for writing and removes it
	// from storage
	Drop() error
}
