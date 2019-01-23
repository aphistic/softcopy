package local

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/uuid"

	"github.com/aphistic/softcopy/internal/pkg/logging"
	"github.com/aphistic/softcopy/internal/pkg/storage"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

func ensureDirectory(directory string) error {
	fi, err := os.Stat(directory)
	if os.IsNotExist(err) {
		err = os.MkdirAll(directory, 0755)
		if err != nil {
			return err
		}
	} else if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", directory)
	} else if err != nil {
		return err
	}

	return nil
}

type FileLocalOption func(*FileLocal)

func WithLogger(logger logging.Logger) FileLocalOption {
	return func(fl *FileLocal) {
		fl.logger = logger
	}
}

type FileLocal struct {
	logger logging.Logger

	basePath string
}

func NewFileLocal(basePath string, opts ...FileLocalOption) (*FileLocal, error) {
	err := ensureDirectory(basePath)
	if err != nil {
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

	fl := &FileLocal{
		basePath: basePath,
	}

	for _, opt := range opts {
		opt(fl)
	}

	return fl, nil
}

func (fl *FileLocal) OpenFile(id uuid.UUID) (storage.OpenFile, error) {
	filePath := path.Join(
		fl.basePath,
		id.String()[0:1],
		id.String()[1:2],
		id.String()+".dat",
	)

	fl.logger.Debug("opening %s for read at %s", id, filePath)

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	of := newOpenLocalFile(f, records.FILE_MODE_READ, fl)

	return of, nil
}

func (fl *FileLocal) OpenTempFile(handleID uuid.UUID) (storage.OpenFile, error) {
	filePath := path.Join(
		fl.basePath,
		"tmp",
		handleID.String()+".dat",
	)

	fl.logger.Debug("opening temp handle %s at %s", handleID, filePath)

	fileDir := path.Dir(filePath)
	err := ensureDirectory(fileDir)
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	of := newOpenLocalFile(f, records.FILE_MODE_WRITE, fl)

	return of, nil
}

func (fl *FileLocal) ReadFile(filePath string) (io.ReadCloser, error) {
	readPath := path.Join(fl.basePath, filePath)

	f, err := os.Open(readPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (fl *FileLocal) ReadFileFromOffset(
	filePath string,
	offset uint64,
) (io.ReadCloser, error) {
	readPath := path.Join(fl.basePath, filePath)

	f, err := os.OpenFile(readPath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (fl *FileLocal) WriteFile(filePath string, r io.Reader) (string, uint64, error) {
	writePath := path.Join(fl.basePath, filePath)

	// Make sure the dir we're writing to exists
	writeDir := path.Dir(writePath)

	err := ensureDirectory(writeDir)
	if err != nil {
		return "", 0, err
	}

	// Open the file for writing
	f, err := os.OpenFile(writePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", 0, err
	}

	// Get the sha256 of the file as we're writing it
	h := sha256.New()

	fileSize := uint64(0)
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
