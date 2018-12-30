package api

import (
	"github.com/aphistic/softcopy/storage/records"
	"github.com/google/uuid"
)

func (c *Client) GetTags(names []string) ([]*records.Tag, error) {
	tags, err := c.dataStorage.GetTags(names)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (c *Client) GetTagsForFile(id string) (records.TagIterator, error) {
	fileID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return c.dataStorage.GetTagsForFile(fileID)
}
