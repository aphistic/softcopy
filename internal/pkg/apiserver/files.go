package apiserver

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/protoutil"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

func (as *apiServer) GetFileYears(
	ctx context.Context,
	req *scproto.GetFileYearsRequest,
) (*scproto.GetFileYearsResponse, error) {
	years, err := as.api.GetFileYears()
	if err != nil {
		return nil, err
	}

	resYears := []int32{}
	for _, year := range years {
		resYears = append(resYears, int32(year))
	}

	return &scproto.GetFileYearsResponse{
		Years: resYears,
	}, nil
}

func (as *apiServer) GetFileMonths(
	ctx context.Context,
	req *scproto.GetFileMonthsRequest,
) (*scproto.GetFileMonthsResponse, error) {
	months, err := as.api.GetFileMonths(int(req.GetYear()))
	if err != nil {
		return nil, err
	}

	resMonths := []int32{}
	for _, month := range months {
		resMonths = append(resMonths, int32(month))
	}

	return &scproto.GetFileMonthsResponse{
		Months: resMonths,
	}, nil
}

func (as *apiServer) GetFileDays(
	ctx context.Context,
	req *scproto.GetFileDaysRequest,
) (*scproto.GetFileDaysResponse, error) {
	days, err := as.api.GetFileDays(int(req.GetYear()), int(req.GetMonth()))
	if err != nil {
		return nil, err
	}

	resDays := []int32{}
	for _, day := range days {
		resDays = append(resDays, int32(day))
	}

	return &scproto.GetFileDaysResponse{
		Days: resDays,
	}, nil
}

func (as *apiServer) GetFileWithDate(
	ctx context.Context,
	req *scproto.GetFileWithDateRequest,
) (*scproto.GetFileWithDateResponse, error) {
	date, err := ptypes.Timestamp(req.DocumentDate)
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	file, err := as.api.GetFileWithDate(req.GetFilename(), date)
	if err == scerrors.ErrNotFound {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf(
			"file %s on %s not found",
			req.GetFilename(), date,
		))
	} else if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	f, err := protoutil.FileToProto(file)
	if err != nil {
		return nil, err
	}

	return &scproto.GetFileWithDateResponse{
		File: f,
	}, nil
}

func (as *apiServer) FindFilesWithDate(
	ctx context.Context,
	req *scproto.FindFilesWithDateRequest,
) (*scproto.FindFilesWithDateResponse, error) {
	date, err := ptypes.Timestamp(req.GetDocumentDate())
	if err != nil {
		return nil, err
	}

	files, err := as.api.FindFilesWithDate(date)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &scproto.FindFilesWithDateResponse{
		Files: []*scproto.File{},
	}

	for _, file := range files {
		f, err := protoutil.FileToProto(file)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		res.Files = append(res.Files, f)
	}

	return res, nil
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
		f, err := protoutil.FileToProto(file)
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
		f, err := protoutil.FileToProto(file)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		res.Files = append(res.Files, f)
	}

	as.logger.Debug("found %d files with prefix '%s'", len(res.Files), req.IdPrefix)

	return res, nil
}

func (as *apiServer) CreateFile(
	ctx context.Context,
	req *scproto.CreateFileRequest,
) (*scproto.CreateFileResponse, error) {
	date, err := ptypes.Timestamp(req.GetDocumentDate())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	as.logger.Debug(
		"creating file %s - %s",
		date.Format("2006-01-02"),
		req.GetFilename(),
	)

	id, err := as.api.CreateFile(req.GetFilename(), date)
	if err == scerrors.ErrExists {
		return nil, status.Error(codes.AlreadyExists, "filename already exists on this date")
	} else if err != nil {
		as.logger.Error("error on create: %s", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.CreateFileResponse{
		Id: id.String(),
	}, nil
}

func (as *apiServer) GetFile(
	ctx context.Context,
	req *scproto.GetFileRequest,
) (*scproto.GetFileResponse, error) {
	f, err := as.api.GetFile(req.GetId())
	if err != nil {
		as.logger.Error("Could not get file %s: %s", req.GetId(), err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	resFile, err := protoutil.FileToProto(f)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.GetFileResponse{
		File: &scproto.TaggedFile{
			File: resFile,
		},
	}, nil
}

func (as *apiServer) RemoveFile(
	ctx context.Context,
	req *scproto.RemoveFileRequest,
) (*scproto.RemoveFileResponse, error) {
	err := as.api.RemoveFile(req.GetId())
	if err == scerrors.ErrNotFound {
		return nil, status.Error(codes.NotFound, "file not found")
	} else if err != nil {
		as.logger.Error("Could not remove file: %s", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.RemoveFileResponse{}, nil
}

func (as *apiServer) UpdateFileDate(
	ctx context.Context,
	req *scproto.UpdateFileDateRequest,
) (*scproto.UpdateFileDateResponse, error) {
	id, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	newDate, err := ptypes.Timestamp(req.GetNewDocumentDate())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = as.api.UpdateFileDate(id, req.GetNewFilename(), newDate)
	if err == scerrors.ErrExists {
		return nil, status.Error(codes.AlreadyExists, "destination exists")
	} else if err == scerrors.ErrNotFound {
		return nil, status.Error(codes.NotFound, "file not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.UpdateFileDateResponse{}, nil
}
