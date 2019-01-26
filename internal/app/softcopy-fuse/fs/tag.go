package fs

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

var _ fusefs.NodeMkdirer = &fsByTagDir{}

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
	_, err := btd.fs.client.FindTagByName(ctx, &scproto.FindTagByNameRequest{
		Name: name,
	})
	if status.Code(err) == codes.NotFound {
		return nil, fuse.ENOENT
	} else if err != nil {
		btd.fs.logger.Error("error finding tag by name: %s", err)
		return nil, err
	}

	return &TagDir{
		tag: name,
		fs:  btd.fs,
	}, nil
}

func (btd *fsByTagDir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fusefs.Node, error) {
	_, err := btd.fs.client.CreateTag(ctx, &scproto.CreateTagRequest{
		Name: req.Name,
	})
	if status.Code(err) == codes.AlreadyExists {
		return nil, fuse.EEXIST
	} else if err != nil {
		return nil, err
	}

	return &TagDir{
		tag: req.Name,
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

func (td *TagDir) Rename(
	ctx context.Context,
	req *fuse.RenameRequest,
	newDir fusefs.Node,
) error {
	// TODO try moving from outside of mount point into a tag directory

	td.fs.logger.Debug("rename:\n%#v\n%#v", req, newDir)
	return fmt.Errorf("not implemented")
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
