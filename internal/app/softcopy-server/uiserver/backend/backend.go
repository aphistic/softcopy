package backend

import (
	"html/template"
	"sync"

	"github.com/aphistic/goblin"

	"github.com/aphistic/softcopy/internal/pkg/errors"
)

//go:generate goblin --name backend --include-root ../../../../../web --include **/*.tpl

type Template struct {
	vault goblin.Vault

	loadedLock sync.RWMutex
	loaded     map[string]*template.Template
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
	data, err := t.vault.ReadFile(file)
	if err != nil {
		t.loadedLock.Unlock()

		return nil, errors.ErrNotFound
	}
	tpl, err = template.New(file).Parse(string(data))
	if err != nil {
		return nil, err
	}

	t.loaded[file] = tpl
	t.loadedLock.Unlock()

	return tpl, nil
}
