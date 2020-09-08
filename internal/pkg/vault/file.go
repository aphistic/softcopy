package vault

import (
	"fmt"
	"os"

	"github.com/aphistic/goblin"
)

type vaultFile struct{}

var _ goblin.File = &vaultFile{}

func newVaultFile() *vaultFile {
	return &vaultFile{}
}

func (vf *vaultFile) Stat() (os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (vf *vaultFile) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (vf *vaultFile) Close() error {
	return fmt.Errorf("not implemented")
}
