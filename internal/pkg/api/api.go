package api

import (
	"path"

	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/internal/pkg/storage"
	dataSqlite "github.com/aphistic/softcopy/internal/pkg/storage/data/sqlite"
	fileLocal "github.com/aphistic/softcopy/internal/pkg/storage/file/local"
)

type Initializer struct {
	Container nacelle.ServiceContainer `service:"container"`
	Logger    nacelle.Logger           `service:"logger"`
}

func NewInitializer() *Initializer {
	return &Initializer{}
}

func (i *Initializer) Init(config nacelle.Config) error {
	cfg := &Config{}
	if err := config.Load(cfg); err != nil {
		return err
	}

	fs, err := fileLocal.NewFileLocal(
		cfg.StorageRoot,
		fileLocal.WithLogger(i.Logger),
	)
	if err != nil {
		return err
	}

	ds, err := dataSqlite.NewClient(path.Join(cfg.StorageRoot, "softcopy.db"))
	if err != nil {
		return err
	}
	err = ds.Migrate()
	if err != nil {
		return err
	}

	err = i.Container.Set("api", &Client{
		cfg:    cfg,
		logger: i.Logger,

		openManager: newOpenFileManager(
			fs, ds,
			withLogger(i.Logger),
		),

		fileStorage: fs,
		dataStorage: ds,
	})
	if err != nil {
		return err
	}

	return nil
}

type Client struct {
	cfg    *Config
	logger nacelle.Logger

	openManager *openFileManager

	fileStorage storage.File
	dataStorage storage.Data
}
