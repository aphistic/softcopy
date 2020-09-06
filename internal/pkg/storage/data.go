package storage

import (
	"time"

	"github.com/google/uuid"

	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

type Data interface {
	CreateFile(string, time.Time) (uuid.UUID, error)
	CreateFileWithID(string, time.Time, uuid.UUID) error
	CreateFileWithTags(string, time.Time, []string) (uuid.UUID, error)
	CreateFileWithIDAndTags(string, time.Time, uuid.UUID, []string) error

	RemoveFile(uuid.UUID) error

	FindFilesWithDate(time.Time) ([]*records.File, error)
	FindFilesWithTags(tagNames []string) ([]*records.File, error)
	FindFilesWithIdPrefix(idPrefix string) ([]*records.File, error)

	GetFileWithDate(string, time.Time) (*records.File, error)

	GetFileYears() ([]int, error)
	GetFileMonths(int) ([]int, error)
	GetFileDays(int, int) ([]int, error)

	AllFiles() (records.FileIterator, error)
	GetFile(id uuid.UUID) (*records.File, error)
	GetFileByHash(hash string) (*records.File, error)
	UpdateFile(*records.File) error

	UpdateFileHash(uuid.UUID, string) error
	UpdateFileDate(uuid.UUID, string, time.Time) error

	AllTags() (records.TagIterator, error)
	GetTags([]string) ([]*records.Tag, error)
	GetTagsForFile(uuid.UUID) (records.TagIterator, error)
	FindTagByName(string) (*records.Tag, error)
	CreateTags([]string) ([]uuid.UUID, error)
	UpdateFileTags(id uuid.UUID, addedTags []string, removedTags []string) error

	FindMetadataByHash(hash string) (*records.FileMetadata, error)
	CreateMetadataWithID(string, uint64, uuid.UUID) error
}
