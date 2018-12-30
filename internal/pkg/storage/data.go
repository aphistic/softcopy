package storage

import (
	"github.com/google/uuid"

	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

type Data interface {
	FindFilesWithTags(tagNames []string) ([]*records.File, error)
	FindFilesWithIdPrefix(idPrefix string) ([]*records.File, error)
	AllFiles() (records.FileIterator, error)
	GetFile(id uuid.UUID) (*records.File, error)
	GetFileByHash(hash string) (*records.File, error)
	CreateFile(*records.File) error
	CreateFileWithTags(*records.File, []string) error
	UpdateFile(*records.File) error
	DeleteFile(id uuid.UUID) error

	GetTags(names []string) ([]*records.Tag, error)
	GetTagsForFile(id uuid.UUID) (records.TagIterator, error)
}
