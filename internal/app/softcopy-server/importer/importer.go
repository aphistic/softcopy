package importer

import (
	"context"
	"fmt"

	"github.com/aphistic/softcopy/internal/pkg/api"
	"github.com/aphistic/softcopy/internal/pkg/config"
	"github.com/aphistic/softcopy/internal/pkg/importers"
	"github.com/aphistic/softcopy/internal/pkg/importers/sftp"
	"github.com/efritz/nacelle"
)

var importerCreators = map[string]importerCreator{
	"sftp": func(loader *config.OptionLoader) (importers.Importer, error) {
		return sftp.NewSFTPImporter(loader)
	},
}

type importerCreator func(*config.OptionLoader) (importers.Importer, error)

type importerProcess struct {
	ServiceContainer nacelle.ServiceContainer `service:"container"`
	Logger           nacelle.Logger           `service:"logger"`
	API              *api.Client              `service:"api"`

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

	for _, importerCfg := range ip.importCfg.Importers {
		importerCreate, ok := importerCreators[importerCfg.Type]
		if !ok {
			return fmt.Errorf("unknown importer type '%s'", importerCfg.Type)
		}

		loader := config.NewOptionLoader(importerCfg.Options)
		importer, err := importerCreate(loader)
		if err != nil {
			return err
		}

		ip.Logger.Debug("container: %#v", ip.ServiceContainer)
		err = ip.ServiceContainer.Inject(importer)
		if err != nil {
			return err
		}

		ip.Logger.Info("Starting importer %s", importer.Name())

		go importer.Start(ctx)
	}

	<-ip.stopChan
	return nil
}

func (ip *importerProcess) Stop() error {
	close(ip.stopChan)
	return nil
}
