package api

import (
	"io"
	"path"
	"time"

	"github.com/google/uuid"

	"github.com/aphistic/papertrail/storage/records"
)

func (c *Client) AddFile(name string, data io.Reader) error {
	fileID, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	fileDir := fileID.String()[0:4]
	filePath := path.Join(fileDir, fileID.String()+".dat")

	sha, err := c.fileStorage.WriteFile(filePath, data)
	if err != nil {
		return err
	}

	// See if the file already exists in the data
	_, err = c.dataStorage.GetFileByHash(sha)
	if err == nil {
		// It already exists so we can't add the file
		return ErrHashCollision
	}

	err = c.dataStorage.CreateFile(&records.File{
		ID:           fileID,
		Hash:         sha,
		Filename:     name,
		DocumentDate: time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}
