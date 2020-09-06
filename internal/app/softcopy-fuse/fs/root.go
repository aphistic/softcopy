package fs

import (
	"context"
	"os"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/google/uuid"
)

var byTagID = uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000001"))
var byDateID = uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000002"))
var uploadID = uuid.Must(uuid.Parse("00000000-0000-0000-0000-000000000003"))

type fsRootDir struct {
	fs *FileSystem
}

func newFSRootDir(fs *FileSystem) *fsRootDir {
	return &fsRootDir{
		fs: fs,
	}
}

func (rd *fsRootDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (rd *fsRootDir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	switch name {
	case "by-tag":
		return newFSByTagDir(rd.fs), nil
	case "by-date":
		return newFSByDateDir(rd.fs), nil
	case "upload":
		return newFSUploadDir(rd.fs), nil
	default:
		return nil, fuse.ENOENT
	}
}

func (rd *fsRootDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return []fuse.Dirent{
		{
			Inode: rd.fs.inodeForID(byTagID),
			Type:  fuse.DT_Dir,
			Name:  "by-tag",
		},
		{
			Inode: rd.fs.inodeForID(byDateID),
			Type:  fuse.DT_Dir,
			Name:  "by-date",
		},
		{
			Inode: rd.fs.inodeForID(uploadID),
			Type:  fuse.DT_Dir,
			Name:  "upload",
		},
	}, nil
}
