package sqlite

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/aphistic/papertrail/storage"
	"github.com/aphistic/papertrail/storage/records"
)

func (c *Client) GetFile(id uuid.UUID) (*records.File, error) {
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) GetFileByHash(hash string) (*records.File, error) {
	rows, err := c.db.Query(`
		SELECT id, hash, filename, document_date FROM files
		WHERE hash = ?;
	`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, storage.ErrNotFound
	}

	res := &records.File{}
	err = rows.Scan(&res.ID, &res.Hash, &res.Filename, &res.DocumentDate)

	if err != nil {

		return nil, err
	}

	return res, nil
}

func (c *Client) CreateFile(file *records.File) error {
	_, err := c.db.Exec(`
		INSERT INTO files (id, hash, filename, document_date)
		VALUES (?, ?, ?, ?);
	`, file.ID.String(), file.Hash, file.Filename, file.DocumentDate)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateFile(file *records.File) error {
	return fmt.Errorf("not implemented")
}

func (c *Client) DeleteFile(id uuid.UUID) error {
	return fmt.Errorf("not implemented")
}
