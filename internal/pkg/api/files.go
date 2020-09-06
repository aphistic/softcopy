package api

import (
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

func (c *Client) AllFiles() (records.FileIterator, error) {
	files, err := c.dataStorage.AllFiles()
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (c *Client) GetFileYears() ([]int, error) {
	return c.dataStorage.GetFileYears()
}

func (c *Client) GetFileMonths(year int) ([]int, error) {
	return c.dataStorage.GetFileMonths(year)
}

func (c *Client) GetFileDays(year int, month int) ([]int, error) {
	return c.dataStorage.GetFileDays(year, month)
}

func (c *Client) CreateFile(filename string, documentDate time.Time) (uuid.UUID, error) {
	file, err := c.dataStorage.GetFileWithDate(filename, documentDate)
	if err != nil && err != scerrors.ErrNotFound {
		return uuid.Nil, err
	} else if file != nil {
		return uuid.Nil, scerrors.ErrExists
	}

	return c.dataStorage.CreateFile(filename, documentDate)
}

func (c *Client) GetFile(id string) (*records.File, error) {
	fileID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid file id")
	}

	file, err := c.dataStorage.GetFile(fileID)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *Client) ReadFile(id string) (io.ReadCloser, error) {
	return c.ReadFileFromOffset(id, 0)
}

func (c *Client) ReadFileFromOffset(id string, offset uint64) (io.ReadCloser, error) {
	fileID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid file id")
	}

	f, err := c.dataStorage.GetFile(fileID)
	if err != nil {
		return nil, err
	}

	md, err := c.dataStorage.FindMetadataByHash(f.Hash)
	if err != nil {
		return nil, err
	}

	filePath := path.Join(
		md.ID.String()[0:1],
		md.ID.String()[1:2],
		md.ID.String()+".dat",
	)

	return c.fileStorage.ReadFileFromOffset(filePath, offset)
}

func (c *Client) RemoveFile(id string) error {
	fileID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid file id")
	}

	return c.dataStorage.RemoveFile(fileID)
}

func (c *Client) UpdateFileTags(
	id uuid.UUID,
	addedTags []string,
	removedTags []string,
) error {
	// Make sure the current file exists
	_, err := c.dataStorage.GetFile(id)
	if err != nil {
		return err
	}

	err = c.dataStorage.UpdateFileTags(id, addedTags, removedTags)
	if err != nil {
		return err
	}

	return nil
}
func (c *Client) UpdateFileDate(
	id uuid.UUID,
	newFilename string,
	newDate time.Time,
) error {
	// Make sure the current file exists
	_, err := c.dataStorage.GetFile(id)
	if err != nil {
		return err
	}

	// Make sure the new date and filename are available
	_, err = c.dataStorage.GetFileWithDate(newFilename, newDate)
	if err != scerrors.ErrNotFound {
		return err
	}

	err = c.dataStorage.UpdateFileDate(id, newFilename, newDate)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetFileWithDate(filename string, date time.Time) (*records.File, error) {
	return c.dataStorage.GetFileWithDate(filename, date)
}

func (c *Client) FindFilesWithDate(documentDate time.Time) ([]*records.File, error) {
	return c.dataStorage.FindFilesWithDate(documentDate)
}

func (c *Client) FindFilesWithTags(tagNames []string) ([]*records.File, error) {
	return c.dataStorage.FindFilesWithTags(tagNames)
}

func (c *Client) FindFilesWithIdPrefix(idPrefix string) ([]*records.File, error) {
	return c.dataStorage.FindFilesWithIdPrefix(idPrefix)
}
