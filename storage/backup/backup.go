package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"

	scproto "github.com/aphistic/softcopy/proto"

	v1 "github.com/aphistic/softcopy/storage/backup/v1"
)

type manifest struct {
	Version float64 `json:"version"`
}

type Backup struct {
	driver  Driver
	rootDir string
}

type Driver interface {
	WriteFile(*scproto.TaggedFile) error
	WriteData(string, io.Reader) error
	WriteTag(*scproto.Tag) error
}

func NewBackup(root string) (*Backup, error) {
	fi, err := os.Stat(root)
	if os.IsNotExist(err) {
		return CreateBackup(root)
	} else if err != nil {
		fmt.Printf("err: %s\n", err)
		return nil, err
	}

	if fi.IsDir() {
		return LoadBackup(root)
	}

	return nil, fmt.Errorf("not a directory")
}

func CreateBackup(root string) (*Backup, error) {
	driver, err := v1.NewDriver(root)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(root, 0755)
	if err != nil {
		return nil, err
	}

	manifestPath := path.Join(root, "softcopy.json")
	err = writeManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	return &Backup{
		driver:  driver,
		rootDir: root,
	}, nil
}

func LoadBackup(root string) (*Backup, error) {
	return nil, fmt.Errorf("not implemented")
}

func writeManifest(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	mf := &manifest{
		Version: 1,
	}

	mfData, err := json.Marshal(mf)
	if err != nil {
		return err
	}
	mfData = append(mfData, '\n')

	curData := 0
	for {
		n, err := f.Write(mfData[curData:])
		if err != nil {
			return err
		}
		curData += n
		if n >= len(mfData) {
			break
		}
	}

	return nil
}

func (b *Backup) WriteFile(file *scproto.TaggedFile) error {
	return b.driver.WriteFile(file)
}

func (b *Backup) WriteData(id string, data io.Reader) error {
	return b.driver.WriteData(id, data)
}
