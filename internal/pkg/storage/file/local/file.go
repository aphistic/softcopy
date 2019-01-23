package local

import (
	"os"
	"path"

	"github.com/google/uuid"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

type openLocalFile struct {
	fileLocal *FileLocal

	mode records.FileMode
	f    *os.File
}

func newOpenLocalFile(f *os.File, mode records.FileMode, fl *FileLocal) *openLocalFile {
	return &openLocalFile{
		fileLocal: fl,

		mode: mode,
		f:    f,
	}
}

func (olf *openLocalFile) Seek(offset int64, whence int) (int64, error) {
	olf.fileLocal.logger.Debug("seeking to %d whence %d", offset, whence)
	return olf.f.Seek(offset, whence)
}

func (olf *openLocalFile) Read(b []byte) (int, error) {
	if olf.mode != records.FILE_MODE_READ {
		return 0, scerrors.ErrInvalidModeAction
	}

	return olf.f.Read(b)
}

func (olf *openLocalFile) Write(b []byte) (int, error) {
	if olf.mode != records.FILE_MODE_WRITE {
		return 0, scerrors.ErrInvalidModeAction
	}

	return olf.f.Write(b)
}

func (olf *openLocalFile) Flush() error {
	return nil
}

func (olf *openLocalFile) Close() error {
	return olf.f.Close()
}

func (olf *openLocalFile) Claim(id uuid.UUID) error {
	if olf.mode != records.FILE_MODE_WRITE {
		return scerrors.ErrInvalidModeAction
	}

	err := olf.Close()
	if err != nil {
		return err
	}

	claimPath := path.Join(
		olf.fileLocal.basePath,
		id.String()[0:1],
		id.String()[1:2],
		id.String()+".dat",
	)

	claimDir := path.Dir(claimPath)

	err = ensureDirectory(claimDir)
	if err != nil {
		return err
	}

	err = os.Rename(olf.f.Name(), claimPath)
	if err != nil {
		return err
	}

	return nil
}

func (olf *openLocalFile) Drop() error {
	if olf.mode != records.FILE_MODE_WRITE {
		return scerrors.ErrInvalidModeAction
	}

	err := olf.Close()
	if err != nil {
		return err
	}

	err = os.Remove(olf.f.Name())
	if err != nil {
		return err
	}

	return nil
}
