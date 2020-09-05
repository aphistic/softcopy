package frontend

import (
	"os"
	"path"

	"github.com/aphistic/goblin"
	"github.com/aphistic/softcopy/internal/pkg/consts"
	"github.com/aphistic/softcopy/internal/pkg/logging"
)

type LoaderOption func(*Loader)

func LoaderLogger(logger logging.Logger) LoaderOption {
	return func(l *Loader) {
		l.logger = logger
	}
}

type Loader struct {
	logger logging.Logger

	frontendVault goblin.Vault
	staticVault   goblin.Vault
}

func NewLoader(opts ...LoaderOption) (*Loader, error) {
	frontendFSVault := goblin.NewFilesystemVault(os.Getenv(consts.EnvWebUIPath))
	frontendMemVault, err := loadVaultFrontend()
	if err != nil {
		return nil, err
	}
	frontendVs := goblin.NewVaultSelector(
		goblin.SelectEnvNotEmpty(consts.EnvWebUIPath, frontendFSVault),
		goblin.SelectDefault(frontendMemVault),
	)

	staticFSVault := goblin.NewFilesystemVault(
		path.Join(os.Getenv(consts.EnvWebUIPath), "static"),
	)
	staticMemVault, err := loadVaultStatic()
	if err != nil {
		return nil, err
	}
	staticVs := goblin.NewVaultSelector(
		goblin.SelectEnvNotEmpty(consts.EnvWebUIPath, staticFSVault),
		goblin.SelectDefault(staticMemVault),
	)

	l := &Loader{
		logger:        logging.NewNilLogger(),
		frontendVault: frontendVs,
		staticVault:   staticVs,
	}

	for _, opt := range opts {
		opt(l)
	}

	l.logger.Debug("frontend using vault: %s", l.frontendVault)
	l.logger.Debug("frontend static using vault: %s", l.staticVault)

	return l, nil
}

func (l *Loader) ReadFile(path string) ([]byte, error) {
	return l.frontendVault.ReadFile(path)
}

func (l *Loader) ReadStaticFile(path string) ([]byte, error) {
	return l.staticVault.ReadFile(path)
}
