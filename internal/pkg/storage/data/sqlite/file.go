package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aphistic/softcopy/internal/pkg/consts"

	"github.com/google/uuid"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

func rowsToFile(rows *sql.Rows) (*records.File, error) {
	file := &records.File{}
	err := rows.Scan(
		&file.ID,
		&file.Filename,
		&file.DocumentDate,
		&file.Hash,
		&file.Size,
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}

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
		file, err := rowsToFile(sfi.rows)
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

func (c *Client) GetFileYears() ([]int, error) {
	rows, err := c.db.Query(`
		SELECT DISTINCT strftime('%Y', document_date) AS document_year FROM files
		ORDER BY document_year;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	years := []int{}
	for rows.Next() {
		var year int

		err = rows.Scan(&year)
		if err != nil {
			return nil, err
		}

		years = append(years, year)
	}

	return years, nil
}

func (c *Client) GetFileMonths(year int) ([]int, error) {
	rows, err := c.db.Query(`
		SELECT DISTINCT
			strftime('%m', document_date) AS document_month
		FROM files
		WHERE strftime('%Y', document_date) = ?
		ORDER BY document_month;
	`,
		fmt.Sprintf("%0000d", year),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	months := []int{}
	for rows.Next() {
		var month int

		err = rows.Scan(&month)
		if err != nil {
			return nil, err
		}

		months = append(months, month)
	}

	return months, nil
}

func (c *Client) GetFileDays(year int, month int) ([]int, error) {
	rows, err := c.db.Query(`
		SELECT DISTINCT
			strftime('%d', document_date) AS document_day
		FROM files
		WHERE
			strftime('%Y', document_date) = ? AND
			strftime('%m', document_date) = ?
		ORDER BY document_day;
	`,
		fmt.Sprintf("%d", year),
		fmt.Sprintf("%02d", month),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	days := []int{}
	for rows.Next() {
		var day int

		err = rows.Scan(&day)
		if err != nil {
			return nil, err
		}

		days = append(days, day)
	}

	return days, nil
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
		SELECT
			f.id,
			f.filename,
			f.document_date,
			f.hash,
			ifnull(fm.file_size, 0) AS file_size
		FROM files f
		LEFT JOIN file_metadata fm ON f.hash = fm.hash
		WHERE f.id = ? ORDER BY f.filename;
	`, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, scerrors.ErrNotFound
	}

	res := &records.File{}
	err = rows.Scan(&res.ID, &res.Filename, &res.DocumentDate, &res.Hash, &res.Size)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) GetFileByHash(hash string) (*records.File, error) {
	rows, err := c.db.Query(`
		SELECT id, filename, document_date, hash FROM files
		WHERE hash = ? ORDER BY filename;
	`, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, scerrors.ErrNotFound
	}

	res := &records.File{}
	err = rows.Scan(&res.ID, &res.Hash, &res.Filename, &res.DocumentDate, &res.Size)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) CreateFile(filename string, documentDate time.Time) (uuid.UUID, error) {
	return c.CreateFileWithTags(
		filename,
		documentDate,
		[]string{
			consts.TagUnfiled,
		},
	)
}

func (c *Client) CreateFileWithID(
	filename string,
	documentDate time.Time,
	fileID uuid.UUID,
) error {
	return c.CreateFileWithIDAndTags(
		filename,
		documentDate,
		fileID,
		[]string{
			consts.TagUnfiled,
		},
	)
}

func (c *Client) CreateFileWithTags(
	filename string,
	documentDate time.Time,
	tagNames []string,
) (uuid.UUID, error) {
	fileID, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	err = c.CreateFileWithIDAndTags(
		filename, documentDate,
		fileID, tagNames,
	)

	if err != nil {
		return uuid.Nil, err
	}

	return fileID, nil
}

func (c *Client) CreateFileWithIDAndTags(
	filename string,
	documentDate time.Time,
	fileID uuid.UUID,
	tagNames []string,
) error {
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
		INSERT INTO files (id, filename, document_date, hash)
		VALUES (?, ?, ?, ?);
	`,
		fileID.String(),
		filename,
		documentDate,
		"",
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, tag := range tags {
		_, err = tx.Exec(`
			INSERT INTO file_tags (file_id, tag_id)
			VALUES (?, ?);
		`,
			fileID.String(),
			tag.ID,
		)
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

func (c *Client) UpdateFileHash(id uuid.UUID, hash string) error {
	res, err := c.db.Exec(
		"UPDATE files SET hash = ? WHERE id = ?",
		hash, id,
	)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected < 1 {
		return scerrors.ErrNotFound
	}

	return nil
}

func (c *Client) UpdateFile(file *records.File) error {
	return fmt.Errorf("not implemented")
}

func (c *Client) RemoveFile(id uuid.UUID) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM file_tags WHERE file_id = ?;", id.String())
	if err != nil {
		return err
	}

	res, err := tx.Exec("DELETE FROM files WHERE id = ?;", id.String())
	if err != nil {
		return err
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra < 1 {
		return scerrors.ErrNotFound
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetFileWithDate(filename string, date time.Time) (*records.File, error) {
	if filename == "" {
		return nil, fmt.Errorf("empty file name")
	}

	query := `
		SELECT
			f.id,
			f.filename,
			f.document_date,
			f.hash,
			ifnull(fm.file_size, 0) AS file_size
		FROM files f
		LEFT JOIN file_metadata fm ON f.hash = fm.hash
		WHERE f.filename = ? AND date(f.document_date) = ?;
	`

	rows, err := c.db.Query(query, filename, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, scerrors.ErrNotFound
	}

	file, err := rowsToFile(rows)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *Client) FindFilesWithDate(documentDate time.Time) ([]*records.File, error) {
	query := `
		SELECT
			f.id,
			f.filename,
			f.document_date,
			f.hash,
			ifnull(fm.file_size, 0) AS file_size
		FROM files f
		LEFT JOIN file_metadata fm ON f.hash = fm.hash
		WHERE date(f.document_date) = ?
	`

	rows, err := c.db.Query(query, documentDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	files := []*records.File{}
	for rows.Next() {
		file, err := rowsToFile(rows)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	return files, nil
}

func (c *Client) FindFilesWithTags(tagNames []string) ([]*records.File, error) {
	if len(tagNames) < 1 {
		return []*records.File{}, nil
	}

	query := "SELECT f.id, f.filename, f.document_date, f.hash, fm.file_size FROM files f "
	query = query + "LEFT JOIN file_metadata fm ON f.hash = fm.hash "
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

		foundFile, err := rowsToFile(rows)
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
		foundFile, err := rowsToFile(rows)
		if err != nil {
			return nil, err
		}

		res = append(res, foundFile)
	}

	return res, nil
}
