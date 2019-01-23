package sqlite

import (
	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	"github.com/google/uuid"
)

func (c *Client) CreateMetadataWithID(hash string, fileSize uint64, id uuid.UUID) error {
	_, err := c.db.Exec(`
		INSERT INTO file_metadata (id, hash, file_size)
		VALUES (?, ?, ?);
	`,
		id, hash, fileSize,
	)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) FindMetadataByHash(hash string) (*records.FileMetadata, error) {
	rows, err := c.db.Query(`
		SELECT id, hash, file_size
		FROM file_metadata fm
		WHERE hash = ?
	`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, scerrors.ErrNotFound
	}

	md := &records.FileMetadata{}
	err = rows.Scan(&md.ID, &md.Hash, &md.FileSize)
	if err != nil {
		return nil, err
	}

	return md, nil
}
