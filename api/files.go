package api

import (
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"

	"github.com/aphistic/papertrail/internal/consts"
	"github.com/aphistic/papertrail/storage/records"
)

func (c *Client) AddFile(name string, data io.Reader) error {
	fileID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	fileDir := fileID.String()[0:4]
	filePath := path.Join(fileDir, fileID.String()+".dat")

	sha, size, err := c.fileStorage.WriteFile(filePath, data)
	if err != nil {
		return err
	}

	// See if the file already exists in the data
	_, err = c.dataStorage.GetFileByHash(sha)
	if err == nil {
		// It already exists so we can't add the file
		return ErrHashCollision
	}

	err = c.dataStorage.CreateFileWithTags(
		&records.File{
			ID:           fileID,
			Hash:         sha,
			Filename:     name,
			DocumentDate: time.Now(),
			Size:         size,
		},
		[]string{consts.TagUnfiled},
	)
	if err != nil {
		return err
	}

	return nil
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

func (c *Client) FindFilesWithTags(tagNames []string) ([]*records.File, error) {
	return c.dataStorage.FindFilesWithTags(tagNames)
}

func (c *Client) FindFilesWithIdPrefix(idPrefix string) ([]*records.File, error) {
	return c.dataStorage.FindFilesWithIdPrefix(idPrefix)
}
