package importserver

import (
	"context"

	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/importers"
	"github.com/aphistic/softcopy/internal/pkg/api"
)

type importerProcess struct {
	ServiceContainer nacelle.ServiceContainer `service:"container"`
	Logger           nacelle.Logger           `service:"logger"`
	API              *api.Client              `service:"api"`
	Importers        *importers.ImportRunners `service:"importers"`

	importCfg *importersConfig

	stopChan chan struct{}
}

func NewProcess() nacelle.Process {
	return &importerProcess{
		stopChan: make(chan struct{}),
	}
}

func (ip *importerProcess) Init(config nacelle.Config) error {
	importCfg := &importersConfig{}
	err := config.Load(importCfg)
	if err != nil {
		return err
	}
	ip.importCfg = importCfg

	return nil
}

func (ip *importerProcess) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, importer := range ip.Importers.Runners() {
		go importer.Start(ctx)
	}

	<-ip.stopChan
	return nil
}

func (ip *importerProcess) Stop() error {
	close(ip.stopChan)
	return nil
}
