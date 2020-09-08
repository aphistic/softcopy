package vault

import (
	"fmt"
	"os"

	"github.com/aphistic/goblin"
)

type byTagDir struct {
	v    *Vault
	path []string
}

var _ goblin.ReadDirFile = &byTagDir{}

func newByTagDir(v *Vault, path []string) (*byTagDir, error) {
	return &byTagDir{
		v:    v,
		path: path,
	}, nil
}

func (btd *byTagDir) Stat() (os.FileInfo, error) {
	return &dirFileInfo{
		name: byTagPath,
		sys:  btd,
	}, nil
}

func (btd *byTagDir) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("cannot read: file is a directory")
}

func (btd *byTagDir) Close() error {
	return nil
}

func (btd *byTagDir) ReadDir(n int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}
