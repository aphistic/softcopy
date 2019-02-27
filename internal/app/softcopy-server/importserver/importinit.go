package importserver

import (
	"fmt"

	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/importers"
	"github.com/aphistic/softcopy/internal/app/softcopy-server/importers/googledrive"
	"github.com/aphistic/softcopy/internal/app/softcopy-server/importers/sftp"
	"github.com/aphistic/softcopy/internal/pkg/config"
	scconfig "github.com/aphistic/softcopy/internal/pkg/config"
)

var importerCreators = map[string]importerCreator{
	"googledrive": func(
		name string,
		loader *config.OptionLoader,
	) (importers.Importer, error) {
		return googledrive.NewGoogleDriveImporter(name, loader)
	},
	"sftp": func(
		name string,
		loader *config.OptionLoader,
	) (importers.Importer, error) {
		return sftp.NewSFTPImporter(name, loader)
	},
}

type importerCreator func(string, *config.OptionLoader) (importers.Importer, error)

type ImportInit struct {
	ServiceContainer nacelle.ServiceContainer `service:"container"`
}

func NewInitializer() *ImportInit {
	return &ImportInit{}
}

func (ii *ImportInit) Init(config nacelle.Config) error {
	importCfg := &importersConfig{}
	err := config.Load(importCfg)
	if err != nil {
		return err
	}

	runners := importers.NewImportRunners()

	for _, importerCfg := range importCfg.Importers {
		importerCreate, ok := importerCreators[importerCfg.Type]
		if !ok {
			return fmt.Errorf("unknown importer type '%s'", importerCfg.Type)
		}

		importerName := "default"
		if importerCfg.Name != "" {
			importerName = importerCfg.Name
		}

		loader := scconfig.NewOptionLoader(importerCfg.Options)
		importer, err := importerCreate(importerName, loader)
		if err != nil {
			return err
		}

		err = ii.ServiceContainer.Inject(importer)
		if err != nil {
			return err
		}

		err = runners.AddRunner(importer)
		if err != nil {
			return fmt.Errorf("could not add new importer: %s", importer.Name())
		}
	}

	err = ii.ServiceContainer.Set("importers", runners)
	if err != nil {
		return err
	}

	return nil
}
