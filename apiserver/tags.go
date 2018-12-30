package apiserver

import (
	scproto "github.com/aphistic/softcopy/proto"
	"github.com/aphistic/softcopy/storage/records"
)

func tagToGrpc(tag *records.Tag) (*scproto.Tag, error) {
	return &scproto.Tag{
		Id:     int64(tag.ID),
		Name:   tag.Name,
		System: tag.System,
	}, nil
}
