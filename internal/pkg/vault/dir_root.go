package vault

import (
	"fmt"
	"os"

	"github.com/aphistic/goblin"
)

type rootDir struct {
	v *Vault

	dirIdx int
}

func newRootDir(v *Vault) *rootDir {
	return &rootDir{
		v: v,
	}
}

var _ goblin.ReadDirFile = &rootDir{}

func (rd *rootDir) Stat() (os.FileInfo, error) {
	return &dirFileInfo{
		name: ".",
		sys:  rd,
	}, nil
}

func (rd *rootDir) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("cannot read: file is a directory")
}

func (rd *rootDir) Close() error {
	rd.dirIdx = 0
	return nil
}

func (rd *rootDir) ReadDir(n int) ([]os.FileInfo, error) {
	bdd, err := newByDateDir(rd.v, []string{})
	if err != nil {
		return nil, err
	}

	btd, err := newByTagDir(rd.v, []string{})
	if err != nil {
		return nil, err
	}

	retDirs := []goblin.File{
		bdd,
		btd,
		newUploadDir(rd.v),
	}

	newIdx, infos, err := returnPart(n, rd.dirIdx, retDirs)
	rd.dirIdx = newIdx

	return infos, err
}
