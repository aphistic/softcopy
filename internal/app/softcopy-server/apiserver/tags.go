package apiserver

import (
	"context"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	scproto "github.com/aphistic/softcopy/pkg/proto"
	"github.com/google/uuid"
)

func tagToGrpc(tag *records.Tag) (*scproto.Tag, error) {
	return &scproto.Tag{
		Id:     tag.ID.String(),
		Name:   tag.Name,
		System: tag.System,
	}, nil
}

func (as *apiServer) GetAllTags(ctx context.Context, req *scproto.GetAllTagsRequest) (*scproto.GetAllTagsResponse, error) {
	tagIter, err := as.api.AllTags()
	if err != nil {
		return nil, grpc.Errorf(codes.Internal, err.Error())
	}

	tags := []*scproto.Tag{}

tagLoop:
	for {
		select {
		case tagItem, ok := <-tagIter.Tags():
			if !ok {
				break tagLoop
			}

			if tagItem.Error != nil {
				return nil, grpc.Errorf(codes.Internal, tagItem.Error.Error())
			}

			tags = append(tags, &scproto.Tag{
				Id:     tagItem.Tag.ID.String(),
				Name:   tagItem.Tag.Name,
				System: tagItem.Tag.System,
			})
		case <-ctx.Done():
			return nil, grpc.Errorf(codes.Canceled, "context finished")
		}
	}

	return &scproto.GetAllTagsResponse{
		Tags: tags,
	}, nil
}

func (as *apiServer) FindTagByName(
	ctx context.Context,
	req *scproto.FindTagByNameRequest,
) (*scproto.FindTagByNameResponse, error) {
	tag, err := as.api.FindTagByName(req.GetName())
	if err == scerrors.ErrNotFound {
		return nil, status.Error(codes.NotFound, "tag not found")
	} else if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.FindTagByNameResponse{
		Tag: &scproto.Tag{
			Id:     tag.ID.String(),
			Name:   tag.Name,
			System: tag.System,
		},
	}, nil
}

func (as *apiServer) GetTagsForFile(
	ctx context.Context,
	req *scproto.GetTagsForFileRequest,
) (*scproto.GetTagsForFileResponse, error) {
	tags, err := as.api.GetTagsForFile(req.FileId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer tags.Close()

	var res []*scproto.Tag
	for {
		select {
		case tagItem := <-tags.Tags():
			if tagItem.Error != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}

			res = append(res, &scproto.Tag{
				Id:     tagItem.Tag.ID.String(),
				Name:   tagItem.Tag.Name,
				System: tagItem.Tag.System,
			})
		case <-ctx.Done():
			return nil, status.Error(codes.Internal, "request cancelled")
		}
	}

	return &scproto.GetTagsForFileResponse{
		Tags: res,
	}, nil
}

func (as *apiServer) UpdateFileTags(
	ctx context.Context,
	req *scproto.UpdateFileTagsRequest,
) (*scproto.UpdateFileTagsResponse, error) {
	id, err := uuid.Parse(req.GetFileId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// First make sure all the added tags are created before we try to
	// assign them.
	_, err = as.api.CreateTags(req.GetAddedTags())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = as.api.UpdateFileTags(
		id,
		req.GetAddedTags(),
		req.GetRemovedTags(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &scproto.UpdateFileTagsResponse{}, nil
}

func (as *apiServer) CreateTags(
	ctx context.Context,
	req *scproto.CreateTagsRequest,
) (*scproto.CreateTagsResponse, error) {
	tags, err := as.api.CreateTags(req.GetNames())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var res []*scproto.Tag
	for _, tag := range tags {
		res = append(res, &scproto.Tag{
			Id:     tag.ID.String(),
			Name:   tag.Name,
			System: tag.System,
		})
	}

	return &scproto.CreateTagsResponse{
		Tags: res,
	}, nil
}
