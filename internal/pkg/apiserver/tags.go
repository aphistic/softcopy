package apiserver

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

func tagToGrpc(tag *records.Tag) (*scproto.Tag, error) {
	return &scproto.Tag{
		Id:     int64(tag.ID),
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
				Id:     int64(tagItem.Tag.ID),
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
