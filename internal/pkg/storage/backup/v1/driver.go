package v1

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/golang/protobuf/ptypes"

	scproto "github.com/aphistic/softcopy/pkg/proto"
)

func writeAll(w io.Writer, data []byte) error {
	curData := 0
	for {
		n, err := w.Write(data[curData:])
		if err != nil {
			return err
		}
		curData += n
		if n >= len(data) {
			break
		}
	}

	return nil
}

type Driver struct {
	rootPath string
}

func NewDriver(root string) (*Driver, error) {
	return &Driver{
		rootPath: root,
	}, nil
}

func (d *Driver) WriteFile(file *scproto.TaggedFile) error {
	fileRoot := path.Join(
		d.rootPath,
		"files",
	)

	_, err := os.Stat(fileRoot)
	if os.IsNotExist(err) {
		err = os.MkdirAll(fileRoot, 0755)
		if err != nil {
			return err
		}
	}

	filePath := path.Join(
		fileRoot,
		fmt.Sprintf("%s.json", file.File.Id),
	)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	tags := []string{}
	for _, tag := range file.Tags {
		tags = append(tags, tag.GetName())
	}

	docDate, err := ptypes.Timestamp(file.File.GetDocumentDate())
	if err != nil {
		return err
	}

	fileData := &fileItem{
		ID: file.File.GetId(),
		Hash: &fileHash{
			Type:  "sha256",
			Value: file.File.GetHash(),
		},
		Filename:     file.File.GetFilename(),
		DocumentDate: docDate.Format(time.RFC3339Nano),
		Size:         float64(file.File.GetSize()),

		Tags: tags,
	}

	rawFile, err := json.Marshal(fileData)
	if err != nil {
		return err
	}

	err = writeAll(f, rawFile)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) WriteData(id string, data io.Reader) error {
	dataRoot := path.Join(
		d.rootPath,
		"data",
	)

	_, err := os.Stat(dataRoot)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dataRoot, 0755)
		if err != nil {
			return err
		}
	}

	dataPath := path.Join(
		dataRoot,
		fmt.Sprintf("%s.dat", id),
	)

	f, err := os.OpenFile(dataPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	for {
		n, err := data.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		curWrite := 0
		for {
			wn, err := f.Write(buf[curWrite:n])
			if err != nil {
				return err
			}

			curWrite += wn
			if curWrite >= n {
				break
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

func (d *Driver) WriteTag(tag *scproto.Tag) error {
	return fmt.Errorf("not implemented")
}
