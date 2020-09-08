package vault

import (
	"fmt"
	"io"
	"os"

	"github.com/aphistic/goblin"
)

type uploadDir struct {
	v *Vault
}

var _ goblin.ReadDirFile = &uploadDir{}

func newUploadDir(v *Vault) *uploadDir {
	return &uploadDir{
		v: v,
	}
}

func (ud *uploadDir) Stat() (os.FileInfo, error) {
	return &dirFileInfo{
		name: uploadPath,
		sys:  ud,
	}, nil
}

func (ud *uploadDir) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("cannot read: file is a directory")
}

func (ud *uploadDir) Close() error {
	return nil
}

func (ud *uploadDir) ReadDir(n int) ([]os.FileInfo, error) {
	var err error
	if n > 0 {
		err = io.EOF
	}

	return []os.FileInfo{}, err
}
