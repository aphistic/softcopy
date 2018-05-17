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
}
