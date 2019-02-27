package importers

import (
	"context"
	"sync"

	"github.com/gorilla/mux"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
)

type Importer interface {
	Name() string

	Start(context.Context) error
	Stop() error
}

type ImporterWebHandler interface {
	SetupWebHandlers(*mux.Router) error
}

type ImportRunners struct {
	importerLock sync.RWMutex
	importers    map[string]Importer
}

func NewImportRunners() *ImportRunners {
	return &ImportRunners{
		importers: map[string]Importer{},
	}
}

func (ir *ImportRunners) AddRunner(importer Importer) error {
	ir.importerLock.Lock()
	defer ir.importerLock.Unlock()

	_, ok := ir.importers[importer.Name()]
	if ok {
		return scerrors.ErrExists
	}

	ir.importers[importer.Name()] = importer

	return nil
}

func (ir *ImportRunners) Runners() []Importer {
	ir.importerLock.RLock()
	defer ir.importerLock.RUnlock()

	res := []Importer{}
	for _, importer := range ir.importers {
		res = append(res, importer)
	}

	return res
}
