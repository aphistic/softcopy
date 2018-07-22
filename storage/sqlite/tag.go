package sqlite

import (
	"fmt"
	"strings"

	"github.com/aphistic/papertrail/storage/records"
)

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
