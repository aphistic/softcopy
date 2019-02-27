package template

import (
	"html/template"
	"sync"

	"github.com/aphistic/goblin"

	"github.com/aphistic/softcopy/internal/pkg/errors"
)

//go:generate goblin --name template --include ../../../../../web/template/*.tpl

type Template struct {
	vault goblin.Vault

	loadedLock sync.RWMutex
	loaded     map[string]*template.Template
}

func LoadTemplates() (*Template, error) {
	vault, err := loadVaultTemplate()
	if err != nil {
		return nil, err
	}

	return &Template{
		vault: vault,
		loaded: map[string]*template.Template{},
	}, nil
}

func (t *Template) Template(file string) (*template.Template, error) {
	t.loadedLock.RLock()
	tpl, ok := t.loaded[file]
	if ok {
		t.loadedLock.RUnlock()
		return tpl, nil
	}
	t.loadedLock.RUnlock()

	t.loadedLock.Lock()
	data, ok := t.vault.File(file)
	if !ok {
		t.loadedLock.Unlock()

		return nil, errors.ErrNotFound
	}
	tpl, err := template.New(file).Parse(string(data))
	if err != nil {
		return nil, err
	}

	t.loaded[file] = tpl
	t.loadedLock.Unlock()

	return tpl, nil
}
