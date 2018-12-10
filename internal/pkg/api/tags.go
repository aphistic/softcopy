package api

import (
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

func (c *Client) GetTags(names []string) ([]*records.Tag, error) {
	tags, err := c.dataStorage.GetTags(names)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
