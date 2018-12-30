package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/aphistic/softcopy/storage"
	"github.com/aphistic/softcopy/storage/records"
)

type sqliteFileIterator struct {
	rows *sql.Rows

	resChan chan *records.FileItem

	closeOnce sync.Once
	closeChan chan struct{}
}

func newSqliteFileIterator(rows *sql.Rows) *sqliteFileIterator {
	sfi := &sqliteFileIterator{
		rows:      rows,
		resChan:   make(chan *records.FileItem),
		closeChan: make(chan struct{}),
	}

	go sfi.worker()

	return sfi
}

func (sfi *sqliteFileIterator) worker() {
	defer func() {
		close(sfi.resChan)
		sfi.rows.Close()
	}()

	for {
		ok := sfi.rows.Next()
		if !ok {
			sfi.Close()
			return
		}

		res := &records.FileItem{}
		file := &records.File{}
		err := sfi.rows.Scan(
			&file.ID,
			&file.Hash,
			&file.Filename,
			&file.DocumentDate,
			&file.Size,
		)
		if err != nil {
			res.Error = err
		} else {
			res.File = file
		}

		select {
		case sfi.resChan <- res:
		case <-sfi.closeChan:
			sfi.Close()
			return
		}
	}
}

func (sfi *sqliteFileIterator) Files() <-chan *records.FileItem {
	return sfi.resChan
}

func (sfi *sqliteFileIterator) Close() error {
	sfi.closeOnce.Do(func() {
		close(sfi.closeChan)
	})
	return nil
}

func (c *Client) AllFiles() (records.FileIterator, error) {
	rows, err := c.db.Query(`
		SELECT id, hash, filename, document_date, file_size FROM files
		ORDER BY filename;
	`)
	if err != nil {
		return nil, err
	}

	return newSqliteFileIterator(rows), nil
}

func (c *Client) GetFile(id uuid.UUID) (*records.File, error) {
	rows, err := c.db.Query(`
		SELECT id, hash, filename, document_date, file_size FROM files
		WHERE id = ? ORDER BY filename;
	`, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, storage.ErrNotFound
	}

	res := &records.File{}
	err = rows.Scan(&res.ID, &res.Hash, &res.Filename, &res.DocumentDate, &res.Size)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) GetFileByHash(hash string) (*records.File, error) {
	rows, err := c.db.Query(`
		SELECT id, hash, filename, document_date, file_size FROM files
		WHERE hash = ? ORDER BY filename;
	`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, storage.ErrNotFound
	}

	res := &records.File{}
	err = rows.Scan(&res.ID, &res.Hash, &res.Filename, &res.DocumentDate, &res.Size)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) CreateFile(file *records.File) error {
	_, err := c.db.Exec(`
		INSERT INTO files (id, hash, filename, document_date, file_size)
		VALUES (?, ?, ?, ?, ?);
	`, file.ID.String(), file.Hash, file.Filename, file.DocumentDate, file.Size)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateFileWithTags(file *records.File, tagNames []string) error {
	// Get the tags we're adding first
	tags, err := c.GetTags(tagNames)
	if err != nil {
		return err
	}

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO files (id, hash, filename, document_date, file_size)
		VALUES (?, ?, ?, ?, ?);
	`, file.ID.String(), file.Hash, file.Filename, file.DocumentDate, file.Size)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, tag := range tags {
		_, err = tx.Exec(`
			INSERT INTO file_tags (file_id, tag_id)
			VALUES (?, ?);
		`, file.ID.String(), tag.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
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

func (c *Client) FindFilesWithTags(tagNames []string) ([]*records.File, error) {
	if len(tagNames) < 1 {
		return []*records.File{}, nil
	}

	query := "SELECT f.id, f.hash, f.filename, f.document_date, f.file_size FROM files f "
	query = query + "INNER JOIN file_tags ft ON f.id = ft.file_id "
	query = query + "INNER JOIN tags t ON t.id = ft.tag_id "
	query = query + "WHERE t.name IN (?"
	query = query + strings.Repeat(", ?", len(tagNames)-1)
	query = query + ") ORDER BY f.filename;"

	args := []interface{}{}
	for _, name := range tagNames {
		args = append(args, name)
	}

	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*records.File{}
	for rows.Next() {
		foundFile := &records.File{}

		err = rows.Scan(
			&foundFile.ID,
			&foundFile.Hash,
			&foundFile.Filename,
			&foundFile.DocumentDate,
			&foundFile.Size,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, foundFile)
	}

	return res, nil
}

func (c *Client) FindFilesWithIdPrefix(idPrefix string) ([]*records.File, error) {
	query := "SELECT id, hash, filename, document_date, file_size FROM files f "
	query = query + "WHERE id LIKE ? ORDER BY id;"

	rows, err := c.db.Query(query, fmt.Sprintf("%s%%", idPrefix))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []*records.File{}
	for rows.Next() {
		foundFile := &records.File{}

		err = rows.Scan(
			&foundFile.ID,
			&foundFile.Hash,
			&foundFile.Filename,
			&foundFile.DocumentDate,
			&foundFile.Size,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, foundFile)
	}

	return res, nil
}
