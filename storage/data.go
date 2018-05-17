package storage

import (
	"github.com/google/uuid"

	"github.com/aphistic/papertrail/storage/records"
)

type Data interface {
	GetFile(id uuid.UUID) (*records.File, error)
	GetFileByHash(hash string) (*records.File, error)
	CreateFile(*records.File) error
	UpdateFile(*records.File) error
	DeleteFile(id uuid.UUID) error
}
