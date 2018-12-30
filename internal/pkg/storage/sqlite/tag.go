package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"

	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

type sqliteTagIterator struct {
	rows *sql.Rows

	resChan chan *records.TagItem

	closeOnce sync.Once
	closeChan chan struct{}
}

func newSqliteTagIterator(rows *sql.Rows) *sqliteTagIterator {
	sti := &sqliteTagIterator{
		rows:      rows,
		resChan:   make(chan *records.TagItem),
		closeChan: make(chan struct{}),
	}

	go sti.worker()

	return sti
}

func (sti *sqliteTagIterator) worker() {
	defer func() {
		close(sti.resChan)
	}()

	for {
		ok := sti.rows.Next()
		if !ok {
			sti.Close()
			return
		}

		res := &records.TagItem{}

		tag := &records.Tag{}
		err := sti.rows.Scan(
			&tag.ID,
			&tag.Name,
			&tag.System,
		)
		if err != nil {
			res.Error = err
		} else {
			res.Tag = tag
		}

		select {
		case sti.resChan <- res:
		case <-sti.closeChan:
			sti.Close()
			return
		}
	}
}

func (sti *sqliteTagIterator) Tags() <-chan *records.TagItem {
	return sti.resChan
}

func (sti *sqliteTagIterator) Close() error {
	sti.closeOnce.Do(func() {
		close(sti.closeChan)
		sti.rows.Close()
	})

	return nil
}

func (c *Client) GetTags(names []string) ([]*records.Tag, error) {
	if len(names) == 0 {
		return []*records.Tag{}, nil
	}

	query := "SELECT id, name, system FROM tags WHERE name IN (?"
	query = query + strings.Repeat(",? ", len(names)-1)
	query = query + ");"

	args := make([]interface{}, 0)
	for _, name := range names {
		args = append(args, name)
	}

	rows, err := c.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*records.Tag, 0)
	for rows.Next() {
		foundTag := &records.Tag{}
		err = rows.Scan(&foundTag.ID, &foundTag.Name, &foundTag.System)
		if err != nil {
			return nil, err
		}

		res = append(res, foundTag)
	}

	if len(res) != len(names) {
		return nil, fmt.Errorf("could not find all tags specified")
	}

	return res, nil
}

func (c *Client) GetTagsForFile(id uuid.UUID) (records.TagIterator, error) {
	query := "SELECT id, name, system FROM tags t "
	query = query + "INNER JOIN file_tags ft ON ft.tag_id = t.id "
	query = query + "WHERE ft.file_id = ?;"

	rows, err := c.db.Query(query, id)
	if err != nil {
		return nil, err
	}

	return newSqliteTagIterator(rows), nil
}
