package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type File interface {
	ReadFile(filePath string) (io.ReadCloser, error)
	WriteFile(filePath string, r io.Reader) (string, int64, error)
}

type FileLocal struct {
	basePath string
}

func NewFileLocal(basePath string) (*FileLocal, error) {
	// Make sure the base path exists. If it doesn't try to create it
	fi, err := os.Stat(basePath)
	if os.IsNotExist(err) {
		// If the path doesn't exist try to create it
		err = os.MkdirAll(basePath, 0755)
		if err != nil {
			return nil, err
		}
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", basePath)
	} else if err != nil {
		return nil, err
	}

	f, err := ioutil.TempFile(basePath, "softcopy-")
	if err != nil {
		return nil, fmt.Errorf("could not write to %s", basePath)
	}
	defer os.Remove(f.Name())
	defer f.Close()

	n, err := f.Write([]byte("test"))
	if err != nil || n < 1 {
		return nil, fmt.Errorf("could not write to %s", basePath)
	}

	return &FileLocal{
		basePath: basePath,
	}, nil
}

func (fl *FileLocal) ReadFile(filePath string) (io.ReadCloser, error) {
	readPath := path.Join(fl.basePath, filePath)

	f, err := os.Open(readPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (fl *FileLocal) WriteFile(filePath string, r io.Reader) (string, int64, error) {
	writePath := path.Join(fl.basePath, filePath)

	// Make sure the dir we're writing to exists
	writeDir := path.Dir(writePath)
	dirStat, err := os.Stat(writeDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(writeDir, 0755)
		if err != nil {
			return "", 0, err
		}
	} else if err != nil {
		return "", 0, err
	} else if !dirStat.IsDir() {
		return "", 0, fmt.Errorf("destination is not a directory")
	}

	// Open the file for writing
	f, err := os.OpenFile(writePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", 0, err
	}

	// Get the sha256 of the file as we're writing it
	h := sha256.New()

	fileSize := int64(0)
	buf := make([]byte, 8192)
	for {
		readN, err := r.Read(buf)
		if err == io.EOF {
			f.Close()

			return fmt.Sprintf("%x", h.Sum(nil)), fileSize, nil
		} else if err != nil {
			f.Close()
			os.Remove(writePath)

			return "", 0, err
		}

		totalWritten := 0
		for {
			writeN, err := h.Write(buf[totalWritten : readN-totalWritten])
			if err != nil {
				f.Close()
				os.Remove(writePath)

				return "", 0, err
			}

			totalWritten += writeN

			if totalWritten >= readN {
				break
			}
		}

		totalWritten = 0
		for {
			writeN, err := f.Write(buf[totalWritten : readN-totalWritten])
			if err != nil {
				f.Close()
				os.Remove(writePath)

				return "", 0, err
			}

			totalWritten += writeN

			if totalWritten >= readN {
				break
			}
		}
	}
}
