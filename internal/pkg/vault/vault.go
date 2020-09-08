package vault

import (
	"fmt"
	"os"
	"strings"

	"github.com/aphistic/goblin"

	scproto "github.com/aphistic/softcopy/pkg/proto"
)

const (
	rootPath      = "."
	pathSeparator = "/"

	uploadPath = "upload"
	byTagPath  = "by-tag"
	byDatePath = "by-date"
)

func splitPath(path string) []string {
	return strings.Split(path, pathSeparator)
}

type Vault struct {
	client scproto.SoftcopyClient
}

func NewVault(client scproto.SoftcopyClient) *Vault {
	return &Vault{
		client: client,
	}
}

var _ goblin.Vault = &Vault{}

func (v *Vault) Open(name string) (goblin.File, error) {
	if name == rootPath {
		return newRootDir(v), nil
	}

	path := splitPath(name)
	if len(path) == 0 {
		return nil, fmt.Errorf("invalid path")
	}

	switch path[0] {
	case uploadPath:
		return newUploadDir(v), nil
	case byDatePath:
		return newByDateDir(v, path[1:])
	case byTagPath:
		return newByTagDir(v, path[1:])
	default:
		return nil, os.ErrNotExist
	}
}

func (v *Vault) ReadDir(name string) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v *Vault) ReadFile(name string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v *Vault) Stat(name string) (os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v *Vault) String() string {
	return "softcopy vault"
}
