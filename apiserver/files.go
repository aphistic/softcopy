package apiserver

import (
	"github.com/golang/protobuf/ptypes"

	"github.com/aphistic/softcopy/proto"
	"github.com/aphistic/softcopy/storage/records"
)

func fileToGrpc(file *records.File) (*scproto.File, error) {
	ts, err := ptypes.TimestampProto(file.DocumentDate)
	if err != nil {
		return nil, err
	}

	return &scproto.File{
		Id:           file.ID.String(),
		Hash:         file.Hash,
		Filename:     file.Filename,
		DocumentDate: ts,
		Size:         file.Size,
	}, nil
}
