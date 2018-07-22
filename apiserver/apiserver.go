package apiserver

import (
	"context"

	"github.com/efritz/nacelle"
	"github.com/efritz/nacelle/process"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aphistic/papertrail/api"
	"github.com/aphistic/papertrail/proto"
)

type apiProcess struct {
	Logger nacelle.Logger `service:"logger"`
	API    *api.Client    `service:"api"`
}

func NewProcess() nacelle.Process {
	return process.NewGRPCServer(
		&apiProcess{},
	)
}

func (ap *apiProcess) Init(config nacelle.Config, server *grpc.Server) error {
	ptproto.RegisterPapertrailServer(server, &apiServer{
		logger: ap.Logger,
		api:    ap.API,
	})

	return nil
}

type apiServer struct {
	logger nacelle.Logger
	api    *api.Client
}

func (as *apiServer) FindFilesWithTags(
	ctx context.Context,
	req *ptproto.FindFilesWithTagsRequest,
) (*ptproto.FindFilesWithTagsResponse, error) {
	files, err := as.api.FindFilesWithTags(req.TagNames)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &ptproto.FindFilesWithTagsResponse{
		Files: []*ptproto.File{},
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
	req *ptproto.FindFilesWithIdPrefixRequest,
) (*ptproto.FindFilesWithIdPrefixResponse, error) {
	as.logger.Debug("finding files with prefix '%s'", req.IdPrefix)
	files, err := as.api.FindFilesWithIdPrefix(req.IdPrefix)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &ptproto.FindFilesWithIdPrefixResponse{
		Files: []*ptproto.File{},
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
	req *ptproto.GetFileRequest,
) (*ptproto.GetFileResponse, error) {
	as.logger.Debug("getting file %s", req.Id)
	f, err := as.api.GetFile(req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resFile, err := fileToGrpc(f)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ptproto.GetFileResponse{
		File: &ptproto.TaggedFile{
			File: resFile,
		},
	}, nil
}
