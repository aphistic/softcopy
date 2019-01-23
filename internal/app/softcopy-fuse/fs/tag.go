package fs

import (
	"context"
	"fmt"
	"os"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/golang/protobuf/ptypes"

	"github.com/aphistic/softcopy/internal/pkg/protoutil"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

type fsByTagDir struct {
	fs *FileSystem
}

func newFSByTagDir(fs *FileSystem) *fsByTagDir {
	return &fsByTagDir{
		fs: fs,
	}
}

func (btd *fsByTagDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (btd *fsByTagDir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	return &TagDir{
		tag: name,
		fs:  btd.fs,
	}, nil
}

func (btd *fsByTagDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	res, err := btd.fs.client.GetAllTags(ctx, &scproto.GetAllTagsRequest{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading all tags: %s\n", err)
		return nil, err
	}

	entries := []fuse.Dirent{}
	for idx, tag := range res.GetTags() {
		entries = append(entries, fuse.Dirent{
			Inode: uint64(idx),
			Type:  fuse.DT_Dir,
			Name:  tag.GetName(),
		})
	}

	return entries, nil
}

type TagDir struct {
	tag string
	fs  *FileSystem
}

func (td *TagDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0755
	return nil
}

func (td *TagDir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	date, filename, err := splitFullFilename(name)
	if err != nil {
		return nil, err
	}

	grpcDate, err := ptypes.TimestampProto(date)
	if err != nil {
		td.fs.logger.Error("could not parse timestamp: %s", err)
		return nil, err
	}

	res, err := td.fs.client.GetFileWithDate(ctx, &scproto.GetFileWithDateRequest{
		DocumentDate: grpcDate,
		Filename:     filename,
	})
	if err != nil {
		td.fs.logger.Error("could not find file with date: %s", err)
		return nil, err
	}

	file, err := protoutil.ProtoToFile(res.GetFile())
	if err != nil {
		return nil, err
	}

	return newFSFile(file, records.FILE_MODE_READ, td.fs), nil
}

func (td *TagDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	res, err := td.fs.client.FindFilesWithTags(ctx, &scproto.FindFilesWithTagsRequest{
		TagNames: []string{td.tag},
	})
	if err != nil {
		td.fs.logger.Error(
			"could not read all files for tag '%s': %s",
			td.tag, err,
		)
		return nil, err
	}

	entries := []fuse.Dirent{}
	for idx, file := range res.GetFiles() {
		date, err := ptypes.Timestamp(file.GetDocumentDate())
		if err != nil {
			td.fs.logger.Error(
				"could not get date for %s: %s",
				file.GetFilename(), err,
			)
			return nil, err
		}

		entries = append(entries, fuse.Dirent{
			Inode: uint64(idx),
			Type:  fuse.DT_Dir,
			Name:  getFullFilename(date, file.GetFilename()),
		})
	}

	return entries, nil
}
