package fs

import (
	"context"
	"os"
	"time"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/gogo/protobuf/types"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aphistic/softcopy/internal/pkg/protoutil"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

type fsUploadDir struct {
	fs *FileSystem
}

func newFSUploadDir(fs *FileSystem) *fsUploadDir {
	return &fsUploadDir{
		fs: fs,
	}
}

func (ud *fsUploadDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (ud *fsUploadDir) Create(
	ctx context.Context,
	req *fuse.CreateRequest,
	res *fuse.CreateResponse,
) (fusefs.Node, fusefs.Handle, error) {
	ud.fs.logger.Debug("create req: %#v", req)

	docDate, err := types.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, nil, err
	}

	createRes, err := ud.fs.client.CreateFile(ctx, &scproto.CreateFileRequest{
		Filename:     req.Name,
		DocumentDate: docDate,
	})
	if status.Code(err) == codes.AlreadyExists {
		return nil, nil, fuse.EEXIST
	} else if err != nil {
		ud.fs.logger.Error("upload create error: %s", err)
		return nil, nil, err
	}

	fileRes, err := ud.fs.client.GetFile(ctx, &scproto.GetFileRequest{
		Id: createRes.GetId(),
	})
	if err != nil {
		ud.fs.logger.Error("upload get file err: %s", err)
		return nil, nil, err
	}

	fileRecord, err := protoutil.ProtoToFile(fileRes.GetFile().GetFile())
	if err != nil {
		return nil, nil, err
	}

	openRes, err := ud.fs.client.OpenFile(ctx, &scproto.OpenFileRequest{
		Id:   createRes.GetId(),
		Mode: scproto.FileMode_WRITE,
	})
	if err != nil {
		return nil, nil, err
	}

	handleID, err := uuid.Parse(openRes.GetHandleId())
	if err != nil {
		return nil, nil, err
	}

	ud.fs.logger.Debug("create: %#v", openRes)

	file := newFSFile(fileRecord, records.FILE_MODE_WRITE, ud.fs)
	fileHandle := newFSFileHandle(handleID, file.file.ID, ud.fs)

	return file, fileHandle, nil
}
