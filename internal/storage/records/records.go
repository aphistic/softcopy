package records

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           uuid.UUID
	Hash         string
	Filename     string
	DocumentDate time.Time
	Size         int64
}

type Tag struct {
	ID       int
	Name     string
	Category *TagCategory
	System   bool
}

type TagCategory struct {
	ID   int
	Name string
}
