package importers

import (
	"context"
)

type Importer interface {
	Name() string

	Start(context.Context) error
	Stop() error
}
