package api

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"sync"

	"github.com/google/uuid"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/logging"
	"github.com/aphistic/softcopy/internal/pkg/storage"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
)

type openFile struct {
	manager *openFileManager

	mode     records.FileMode
	handleID uuid.UUID
	fileID   uuid.UUID

	hasher      hash.Hash
	writtenSize uint64
	storageFile storage.OpenFile
}

var _ records.OpenFile = &openFile{}

func newOpenFile(
	handleID uuid.UUID,
	fileID uuid.UUID,
	mode records.FileMode,
	storageFile storage.OpenFile,
	manager *openFileManager,
) *openFile {
	return &openFile{
		manager: manager,

		mode:     mode,
		handleID: handleID,
		fileID:   fileID,

		hasher:      sha256.New(),
		storageFile: storageFile,
	}
}

func (of *openFile) Mode() records.FileMode {
	return of.mode
}

func (of *openFile) HandleID() string {
	return of.handleID.String()
}

func (of *openFile) WrittenSize() uint64 {
	return of.writtenSize
}

func (of *openFile) WrittenHash() string {
	return fmt.Sprintf("%x", of.hasher.Sum(nil))
}

func (of *openFile) Seek(offset int64, whence int) (int64, error) {
	return of.storageFile.Seek(offset, whence)
}

func (of *openFile) Read(b []byte) (int, error) {
	if of.Mode() != records.FILE_MODE_READ {
		return 0, scerrors.ErrInvalidModeAction
	}

	of.manager.logger.Debug("performing read")

	return of.storageFile.Read(b)
}

func (of *openFile) Write(b []byte) (int, error) {
	if of.Mode() != records.FILE_MODE_WRITE {
		return 0, scerrors.ErrInvalidModeAction
	}

	n, err := of.storageFile.Write(b)

	if n > 0 {
		hashed := 0
		for {
			hashN, hashErr := of.hasher.Write(b[:n])
			if hashErr != nil {
				return hashN, hashErr
			}

			hashed += hashN
			of.writtenSize += uint64(hashN)
			if hashed >= n {
				break
			}
		}
	}

	if err != nil {
		return n, err
	}

	return n, err
}

func (of *openFile) Flush() error {
	return of.storageFile.Flush()
}

func (of *openFile) Close() error {
	// Closing is handled by the manager so the hash of the
	// file can be checked against the database to see if
	// a new file needs to be stored or if this is a copy of
	// an existing file.
	return of.manager.closeFile(of.handleID)
}

type managerOption func(*openFileManager)

func withLogger(logger logging.Logger) managerOption {
	return func(ofm *openFileManager) {
		ofm.logger = logger
	}
}

type openFileManager struct {
	logger logging.Logger

	fileStorage storage.File
	dataStorage storage.Data

	openFilesLock sync.RWMutex
	openHandleIDs map[uuid.UUID]*openFile
	openFileIDs   map[uuid.UUID]*openFile
}

func newOpenFileManager(
	fileStorage storage.File,
	dataStorage storage.Data,
	opts ...managerOption,
) *openFileManager {

	ofm := &openFileManager{
		logger: logging.NewNilLogger(),

		fileStorage: fileStorage,
		dataStorage: dataStorage,

		openHandleIDs: map[uuid.UUID]*openFile{},
		openFileIDs:   map[uuid.UUID]*openFile{},
	}

	for _, opt := range opts {
		opt(ofm)
	}

	return ofm
}

func (ofm *openFileManager) OpenFile(id uuid.UUID, mode records.FileMode) (*openFile, error) {
	switch mode {
	case records.FILE_MODE_READ:
		return ofm.openFileRead(id)
	case records.FILE_MODE_WRITE:
		return ofm.openFileWrite(id)
	default:
		return nil, fmt.Errorf("unknown file mode")
	}
}

func (ofm *openFileManager) openFileRead(id uuid.UUID) (*openFile, error) {
	dataFile, err := ofm.dataStorage.GetFile(id)
	if err != nil {
		return nil, err
	}

	md, err := ofm.dataStorage.FindMetadataByHash(dataFile.Hash)
	if err != nil {
		ofm.logger.Error("could not find md: %s", err)
		return nil, err
	}

	fsOpenFile, err := ofm.fileStorage.OpenFile(md.ID)
	if err != nil {
		ofm.logger.Error("could not open file storage: %s", err)
		return nil, err
	}

	handleID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	of := newOpenFile(
		handleID,
		id,
		records.FILE_MODE_READ,
		fsOpenFile,
		ofm,
	)

	ofm.openFilesLock.Lock()
	_, ok := ofm.openFileIDs[id]
	if ok {
		ofm.openFilesLock.Unlock()
		return nil, scerrors.ErrAlreadyOpen
	}

	ofm.openHandleIDs[handleID] = of
	ofm.openFileIDs[id] = of
	ofm.openFilesLock.Unlock()

	return of, nil
}

func (ofm *openFileManager) openFileWrite(id uuid.UUID) (*openFile, error) {
	fsOpenFile, err := ofm.fileStorage.OpenTempFile(id)
	if err != nil {
		ofm.logger.Error("could not open file storage: %s", err)
		return nil, err
	}

	handleID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	of := newOpenFile(handleID, id, records.FILE_MODE_WRITE, fsOpenFile, ofm)

	ofm.openFilesLock.Lock()
	_, ok := ofm.openFileIDs[id]
	if ok {
		ofm.openFilesLock.Unlock()
		return nil, scerrors.ErrAlreadyOpen
	}

	ofm.openHandleIDs[handleID] = of
	ofm.openFileIDs[id] = of
	ofm.openFilesLock.Unlock()

	return of, nil
}

func (ofm *openFileManager) closeFile(handleID uuid.UUID) error {
	ofm.logger.Debug("closing file handle %s", handleID)

	of, err := ofm.FileByHandle(handleID)
	if err != nil {
		return err
	}

	switch of.Mode() {
	case records.FILE_MODE_READ:
		// If we're open in read mode we just need to close the existing
		// file handle

		err = of.storageFile.Close()
		if err != nil {
			ofm.logger.Error("could not close read mode file: %s", err)
			return err
		}
	case records.FILE_MODE_WRITE:
		// If we're open in write mode we need to handle claiming or
		// dropping the temporary file.

		claimed := false
		hash := of.WrittenHash()
		_, err = ofm.dataStorage.FindMetadataByHash(hash)
		if err == scerrors.ErrNotFound {
			// If this hash hasn't been found before, add a new one to the
			// data store and claim the temporary file.
			id, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			err = of.storageFile.Claim(id)
			if err != nil {
				return err
			}
			claimed = true

			err = ofm.dataStorage.CreateMetadataWithID(hash, of.WrittenSize(), id)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		if !claimed {
			err = of.storageFile.Drop()
			if err != nil {
				return err
			}
		}

		err = ofm.dataStorage.UpdateFileHash(of.fileID, of.WrittenHash())
		if err != nil {
			return err
		}
	}

	ofm.openFilesLock.Lock()
	defer ofm.openFilesLock.Unlock()

	of, ok := ofm.openHandleIDs[handleID]
	if ok {
		delete(ofm.openFileIDs, of.fileID)
	}
	delete(ofm.openHandleIDs, of.handleID)

	return nil
}

func (ofm *openFileManager) FileByHandle(handleID uuid.UUID) (*openFile, error) {
	ofm.openFilesLock.RLock()
	defer ofm.openFilesLock.RUnlock()

	of, ok := ofm.openHandleIDs[handleID]
	if !ok {
		return nil, scerrors.ErrNotFound
	}

	return of, nil
}

func (c *Client) OpenFile(fileID uuid.UUID, mode records.FileMode) (records.OpenFile, error) {
	return c.openManager.OpenFile(fileID, mode)
}

func (c *Client) FileByHandle(handleID uuid.UUID) (records.OpenFile, error) {
	return c.openManager.FileByHandle(handleID)
}
