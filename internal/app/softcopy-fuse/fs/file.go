package fs

import (
	"context"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aphistic/softcopy/internal/pkg/protoutil"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

type fsFile struct {
	fs   *FileSystem
	mode records.FileMode
	file *records.File
}

func newFSFile(file *records.File, mode records.FileMode, fs *FileSystem) *fsFile {
	return &fsFile{
		fs:   fs,
		mode: mode,
		file: file,
	}
}

func (f *fsFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = 0444
	attr.Size = f.file.Size
	return nil
}

func (f *fsFile) Open(
	ctx context.Context,
	req *fuse.OpenRequest,
	res *fuse.OpenResponse,
) (fusefs.Handle, error) {
	openRes, err := f.fs.client.OpenFile(ctx, &scproto.OpenFileRequest{
		Id:   f.file.ID.String(),
		Mode: protoutil.FileModeToProto(f.mode),
	})
	if status.Code(err) == codes.NotFound {
		return nil, fuse.ENOENT
	} else if err != nil {
		f.fs.logger.Error("could not open file: %s", err)
		return nil, err
	}

	handleID, err := uuid.Parse(openRes.GetHandleId())
	if err != nil {
		f.fs.logger.Error("invalid handle id when opening file: %s", err)
		return nil, err
	}

	return newFSFileHandle(handleID, f.file.ID, f.fs), nil
}

type fsFileHandle struct {
	fs *FileSystem

	handleID uuid.UUID
	fileID   uuid.UUID
}

func newFSFileHandle(handleID uuid.UUID, fileID uuid.UUID, fs *FileSystem) *fsFileHandle {
	return &fsFileHandle{
		fs: fs,

		handleID: handleID,
		fileID:   fileID,
	}
}

func (fh *fsFileHandle) Read(
	ctx context.Context,
	req *fuse.ReadRequest,
	res *fuse.ReadResponse,
) error {
	protoRes, err := fh.fs.client.ReadFile(ctx, &scproto.ReadFileRequest{
		HandleId: fh.handleID.String(),
		Offset:   uint64(req.Offset),
		Size:     uint64(req.Size),
	})
	if err != nil {
		fh.fs.logger.Error("read file error: %s", err)
		return err
	}

	res.Data = protoRes.GetData()

	return nil
}

func (fh *fsFileHandle) Write(
	ctx context.Context,
	req *fuse.WriteRequest,
	res *fuse.WriteResponse,
) error {
	writeRes, err := fh.fs.client.WriteFile(ctx, &scproto.WriteFileRequest{
		HandleId: fh.handleID.String(),
		Data:     req.Data,
	})
	if err != nil {
		fh.fs.logger.Error("write error: %s", err)
		return err
	}

	res.Size = int(writeRes.GetAmountWritten())

	return nil
}

func (fh *fsFileHandle) Flush(
	ctx context.Context,
	req *fuse.FlushRequest,
) error {
	_, err := fh.fs.client.FlushFile(ctx, &scproto.FlushFileRequest{
		HandleId: fh.handleID.String(),
	})
	if err != nil {
		fh.fs.logger.Error("flush error: %s", err)
		return err
	}

	_, err = fh.fs.client.CloseFile(ctx, &scproto.CloseFileRequest{
		HandleId: fh.handleID.String(),
	})
	if err != nil {
		fh.fs.logger.Error("Error closing file: %s", err)
		return err
	}

	return nil
}
