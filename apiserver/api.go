package apiserver

import (
	"context"

	"github.com/efritz/nacelle"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aphistic/softcopy/api"
	scproto "github.com/aphistic/softcopy/proto"
)

type apiServer struct {
	logger nacelle.Logger
	api    *api.Client
}

func (as *apiServer) FindFilesWithTags(
	ctx context.Context,
	req *scproto.FindFilesWithTagsRequest,
) (*scproto.FindFilesWithTagsResponse, error) {
	files, err := as.api.FindFilesWithTags(req.TagNames)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &scproto.FindFilesWithTagsResponse{
		Files: []*scproto.File{},
	}

	for _, file := range files {
		f, err := fileToGrpc(file)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		res.Files = append(res.Files, f)
	}

	return res, nil
}

func (as *apiServer) FindFilesWithIdPrefix(
	ctx context.Context,
	req *scproto.FindFilesWithIdPrefixRequest,
) (*scproto.FindFilesWithIdPrefixResponse, error) {
	as.logger.Debug("finding files with prefix '%s'", req.IdPrefix)
	files, err := as.api.FindFilesWithIdPrefix(req.IdPrefix)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &scproto.FindFilesWithIdPrefixResponse{
		Files: []*scproto.File{},
	}

	for _, file := range files {
		f, err := fileToGrpc(file)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		res.Files = append(res.Files, f)
	}

	as.logger.Debug("found %d files with prefix '%s'", len(res.Files), req.IdPrefix)

	return res, nil
}

func (as *apiServer) GetFile(
	ctx context.Context,
	req *scproto.GetFileRequest,
) (*scproto.GetFileResponse, error) {
	as.logger.Debug("getting file %s", req.Id)
	f, err := as.api.GetFile(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resFile, err := fileToGrpc(f)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.GetFileResponse{
		File: &scproto.TaggedFile{
			File: resFile,
		},
	}, nil
}
