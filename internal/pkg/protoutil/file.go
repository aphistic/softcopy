package protoutil

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"

	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

func FileModeToProto(mode records.FileMode) scproto.FileMode {
	switch mode {
	case records.FILE_MODE_READ:
		return scproto.FileMode_READ
	case records.FILE_MODE_WRITE:
		return scproto.FileMode_WRITE
	default:
		return scproto.FileMode_UNKNOWN
	}
}

func ProtoToFileMode(mode scproto.FileMode) records.FileMode {
	switch mode {
	case scproto.FileMode_READ:
		return records.FILE_MODE_READ
	case scproto.FileMode_WRITE:
		return records.FILE_MODE_WRITE
	default:
		return records.FILE_MODE_UNKNOWN
	}
}

func FileToProto(file *records.File) (*scproto.File, error) {
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

func ProtoToFile(file *scproto.File) (*records.File, error) {
	date, err := ptypes.Timestamp(file.GetDocumentDate())
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(file.GetId())
	if err != nil {
		return nil, err
	}

	return &records.File{
		ID:           id,
		Hash:         file.GetHash(),
		Filename:     file.GetFilename(),
		DocumentDate: date,
		Size:         file.GetSize(),
	}, nil
}
