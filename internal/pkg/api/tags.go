package api

import (
	"github.com/google/uuid"

	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

func (c *Client) AllTags() (records.TagIterator, error) {
	return c.dataStorage.AllTags()
}

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

func (c *Client) FindTagByName(name string) (*records.Tag, error) {
	return c.dataStorage.FindTagByName(name)
}

func (c *Client) CreateTag(name string) (*records.Tag, error) {
	id, err := c.dataStorage.CreateTag(name)
	if err != nil {
		return nil, err
	}

	return &records.Tag{
		ID:   int(id),
		Name: name,
	}, nil
}
