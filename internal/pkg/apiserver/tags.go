package apiserver

import (
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
