package apiserver

import (
	"github.com/golang/protobuf/ptypes"

	"github.com/aphistic/papertrail/proto"
	"github.com/aphistic/papertrail/storage/records"
)

func fileToGrpc(file *records.File) (*ptproto.File, error) {
	ts, err := ptypes.TimestampProto(file.DocumentDate)
	if err != nil {
		return nil, err
	}

	return &ptproto.File{
		Id:           file.ID.String(),
		Hash:         file.Hash,
		Filename:     file.Filename,
		DocumentDate: ts,
		Size:         file.Size,
	}, nil
}
