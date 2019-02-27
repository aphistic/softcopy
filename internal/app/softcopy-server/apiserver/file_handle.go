package apiserver

import (
	"context"
	"io"

	"github.com/aphistic/softcopy/internal/pkg/protoutil"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	scproto "github.com/aphistic/softcopy/pkg/proto"
	"github.com/google/uuid"
)

func (as *apiServer) OpenFile(
	ctx context.Context,
	req *scproto.OpenFileRequest,
) (*scproto.OpenFileResponse, error) {
	as.logger.Debug("opening file %s with mode %s", req.GetId(), req.GetMode())

	handleID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	of, err := as.api.OpenFile(
		handleID,
		protoutil.ProtoToFileMode(req.GetMode()),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.OpenFileResponse{
		HandleId: of.HandleID(),
	}, nil
}

func (as *apiServer) ReadFile(
	ctx context.Context,
	req *scproto.ReadFileRequest,
) (*scproto.ReadFileResponse, error) {
	as.logger.Debug(
		"read handle %s at %d of size %d",
		req.GetHandleId(),
		req.GetOffset(),
		req.GetSize(),
	)

	handleID, err := uuid.Parse(req.GetHandleId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	of, err := as.api.FileByHandle(handleID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = of.Seek(int64(req.GetOffset()), io.SeekStart)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resBuf := make([]byte, req.GetSize())
	curOffset := uint64(0)

	buf := make([]byte, req.GetSize(), req.GetSize())
	for {
		n, err := of.Read(buf)
		if err != io.EOF && err != nil {
			as.logger.Error("error reading file: %s", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		resBuf = append(resBuf[curOffset:], buf[:n]...)
		curOffset += uint64(n)

		if err == io.EOF {
			break
		} else if curOffset >= req.Size {
			break
		}
	}

	return &scproto.ReadFileResponse{
		Data: buf[:curOffset],
	}, nil
}

func (as *apiServer) WriteFile(
	ctx context.Context,
	req *scproto.WriteFileRequest,
) (*scproto.WriteFileResponse, error) {
	as.logger.Debug("write file")

	handleID, err := uuid.Parse(req.GetHandleId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	of, err := as.api.FileByHandle(handleID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	curData := 0
	for {
		n, err := of.Write(req.GetData()[curData:])
		if err != nil {
			as.logger.Error("error writing: %s", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

		curData += n
		if curData >= len(req.GetData()) {
			break
		}
	}

	return &scproto.WriteFileResponse{
		AmountWritten: uint64(curData),
	}, nil
}

func (as *apiServer) FlushFile(
	ctx context.Context,
	req *scproto.FlushFileRequest,
) (*scproto.FlushFileResponse, error) {
	as.logger.Debug("flush file")

	handleID, err := uuid.Parse(req.GetHandleId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	of, err := as.api.FileByHandle(handleID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = of.Flush()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.FlushFileResponse{}, nil
}

func (as *apiServer) CloseFile(
	ctx context.Context,
	req *scproto.CloseFileRequest,
) (*scproto.CloseFileResponse, error) {
	as.logger.Debug("close file")

	handleID, err := uuid.Parse(req.GetHandleId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	of, err := as.api.FileByHandle(handleID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	as.logger.Debug("file hash: %s", of.WrittenHash())

	err = of.Close()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.CloseFileResponse{
		Hash: of.WrittenHash(),
	}, nil
}
