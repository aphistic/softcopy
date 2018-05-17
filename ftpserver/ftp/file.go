package ftp

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
)

type File interface {
	io.ReadCloser
}

type fileWriter interface {
	File

	Write(b []byte) (n int, err error)
	CompleteWriting() error
}

type diskFile struct {
	dirPath string
	file    *os.File
}

func newDiskFile(dirPath string) (*diskFile, error) {
	f, err := ioutil.TempFile(dirPath, "")
	if err != nil {
		return nil, err
	}

	return &diskFile{
		dirPath: dirPath,
		file:    f,
	}, nil
}

func (df *diskFile) Write(b []byte) (n int, err error) {
	return df.file.Write(b)
}

func (df *diskFile) CompleteWriting() error {
	// Writing is finished so close the file and re-open it
	// for reading
	err := df.file.Close()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(df.file.Name(), os.O_RDONLY, 0755)
	if err != nil {
		return err
	}

	df.file = f

	return nil
}

func (df *diskFile) Read(b []byte) (n int, err error) {
	return df.file.Read(b)
}

func (df *diskFile) Close() error {
	// When a disk file is closed make sure it's removed from the filesystem
	err := df.file.Close()
	if err != nil {
		return err
	}

	/*	err = os.Remove(df.file.Name())
		if err != nil {
			return err
		}
	*/
	return nil
}

type memoryFile struct {
	buf *bytes.Buffer
}

func newMemoryFile() *memoryFile {
	return &memoryFile{
		buf: bytes.NewBuffer(nil),
	}
}

func (mf *memoryFile) Write(b []byte) (n int, err error) {
	return mf.buf.Write(b)
}

func (mf *memoryFile) CompleteWriting() error {
	return nil
}

func (mf *memoryFile) Read(b []byte) (n int, err error) {
	return mf.buf.Read(b)
}

func (mf *memoryFile) Close() error {
	return nil
}
