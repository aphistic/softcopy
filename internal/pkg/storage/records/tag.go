package records

import "github.com/google/uuid"

type TagIterator interface {
	Tags() <-chan *TagItem
	Close() error
}

type TagItem struct {
	Tag   *Tag
	Error error
}

type Tag struct {
	ID       uuid.UUID
	Name     string
	Category *TagCategory
	System   bool
}

type TagCategory struct {
	ID   uuid.UUID
	Name string
}
