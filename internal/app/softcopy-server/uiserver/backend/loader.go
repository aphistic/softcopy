package backend

import (
	"fmt"
	"html/template"
	"os"
	"sync"
	"time"

	goblin "github.com/aphistic/goblin"
	"github.com/aphistic/softcopy/internal/pkg/consts"
	"github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/logging"
)

type loadedTemplate struct {
	LoadedTime time.Time
	Template   *template.Template
}

type LoaderOption func(*Loader)

func LoaderLogger(logger logging.Logger) LoaderOption {
	return func(l *Loader) {
		l.logger = logger
	}
}

type Loader struct {
	logger logging.Logger

	loadedLock sync.RWMutex
	loaded     map[string]loadedTemplate
	vault      goblin.Vault
}

func NewLoader(opts ...LoaderOption) (*Loader, error) {
	fsVault := goblin.NewFilesystemVault(os.Getenv(consts.EnvWebPath))

	memVault, err := loadVaultBackend()
	if err != nil {
		return nil, err
	}

	vs := goblin.NewVaultSelector(
		goblin.SelectEnvNotEmpty(consts.EnvWebPath, fsVault),
		goblin.SelectDefault(memVault),
	)

	l := &Loader{
		logger: logging.NewNilLogger(),
		loaded: map[string]loadedTemplate{},
		vault:  vs,
	}

	for _, opt := range opts {
		opt(l)
	}

	l.logger.Debug("backend using vault: %s", l.vault)

	return l, nil
}

func (l *Loader) ReadFile(path string) ([]byte, error) {
	return l.vault.ReadFile(path)
}

func (l *Loader) Template(path string) (*template.Template, error) {
	l.loadedLock.RLock()
	loadedTpl, ok := l.loaded[path]
	if ok {
		// Check if the template on the filesystem has been changed since the last time
		// we compiled the template. If not, return it.
		fi, err := l.vault.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("could not stat file: %w", err)
		}
		if fi.ModTime().Before(loadedTpl.LoadedTime) {
			l.loadedLock.RUnlock()
			return loadedTpl.Template, nil
		}
	}
	l.loadedLock.RUnlock()

	// If we haven't loaded the template yet or it's changed, build
	// it, cache it and return it.
	l.loadedLock.Lock()
	data, err := l.vault.ReadFile(path)
	if err != nil {
		l.loadedLock.Unlock()

		return nil, errors.ErrNotFound
	}

	tpl, err := template.New(path).Parse(string(data))
	if err != nil {
		return nil, err
	}
	l.loaded[path] = loadedTemplate{
		LoadedTime: time.Now(),
		Template:   tpl,
	}
	l.loadedLock.Unlock()

	return tpl, nil
}
