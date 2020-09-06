package fs

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	fusefs "bazil.org/fuse/fs"

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
	_, err := btd.fs.client.CreateTags(ctx, &scproto.CreateTagsRequest{
		Names: []string{req.Name},
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
		return nil, fuse.ENOENT
	}

	grpcDate, err := types.TimestampProto(date)
	if err != nil {
		td.fs.logger.Error("could not parse timestamp: %s", err)
		return nil, err
	}

	res, err := td.fs.client.GetFileWithDate(ctx, &scproto.GetFileWithDateRequest{
		DocumentDate: grpcDate,
		Filename:     filename,
	})
	if err != nil && status.Code(err) == codes.NotFound {
		return nil, fuse.ENOENT
	} else if err != nil {
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

	addTag := ""
	if tagDir, ok := newDir.(*TagDir); ok {
		addTag = tagDir.tag
	}

	oldDate, oldName, err := splitFullFilename(req.OldName)
	if err != nil {
		return fmt.Errorf("could not get old filename parts: %w", err)
	}
	newDate, newName, err := splitFullFilename(req.NewName)
	if err != nil {
		return fmt.Errorf("could not get new filename parts: %w", err)
	}

	grpcDate, err := types.TimestampProto(oldDate)
	if err != nil {
		return err
	}

	getRes, err := td.fs.client.GetFileWithDate(ctx, &scproto.GetFileWithDateRequest{
		Filename:     oldName,
		DocumentDate: grpcDate,
	})
	if err != nil && status.Code(err) == codes.NotFound {
		return fuse.ENOENT
	} else if err != nil {
		return err
	}
	oldFile, err := protoutil.ProtoToFile(getRes.File)
	if err != nil {
		return err
	}

	if addTag != "" {
		err = td.addTagToFile(ctx, oldFile, addTag)
	}
	if err != nil {
		return err
	}

	if oldDate != newDate || oldName != newName {
		// TODO probably doesn't work
		err = td.renameFile(ctx, oldFile, newDate, newName)
	}
	if err != nil {
		return err
	}

	return nil
}

func (td *TagDir) addTagToFile(ctx context.Context, oldFile *records.File, tag string) error {
	_, err := td.fs.client.UpdateFileTags(ctx, &scproto.UpdateFileTagsRequest{
		FileId:    oldFile.ID.String(),
		AddedTags: []string{tag},
	})
	if err != nil {
		return err
	}

	return nil
}

func (td *TagDir) renameFile(
	ctx context.Context,
	oldFile *records.File,
	newDate time.Time,
	newName string,
) error {
	newProtoDate, err := types.TimestampProto(newDate)
	if err != nil {
		return err
	}

	_, err = td.fs.client.UpdateFileDate(ctx, &scproto.UpdateFileDateRequest{
		FileId:          oldFile.ID.String(),
		NewDocumentDate: newProtoDate,
		NewFilename:     newName,
	})
	if err != nil {
		return err
	}

	return nil
}

func (td *TagDir) Create(
	ctx context.Context,
	req *fuse.CreateRequest,
	res *fuse.CreateResponse,
) (fs.Node, fs.Handle, error) {
	td.fs.logger.Debug("tag dir create: %v", req)

	docDate, err := types.TimestampProto(time.Now().UTC())
	if err != nil {
		return nil, nil, err
	}

	if req.Flags&fuse.OpenCreate != fuse.OpenCreate {
		td.fs.logger.Error("not allowing tag create without create file flag")
		return nil, nil, fuse.EPERM
	}
	if req.Flags&fuse.OpenWriteOnly != fuse.OpenWriteOnly &&
		req.Flags&fuse.OpenReadWrite != fuse.OpenReadWrite {
		td.fs.logger.Error("not allowing tag create without write file flag")
		return nil, nil, fuse.EPERM
	}

	// Maybe if file already exists it's ok because it means it's probably the
	// same one?
	createRes, err := td.fs.client.CreateFile(ctx, &scproto.CreateFileRequest{
		Filename:     req.Name,
		DocumentDate: docDate,
	})
	if status.Code(err) == codes.AlreadyExists {
		return nil, nil, fuse.EEXIST
	} else if err != nil {
		td.fs.logger.Error("tag create error: %s", err)
		return nil, nil, err
	}

	fileRes, err := td.fs.client.GetFile(ctx, &scproto.GetFileRequest{
		Id: createRes.GetId(),
	})
	if err != nil {
		td.fs.logger.Error("tag create get file err: %s", err)
		return nil, nil, err
	}

	fileRecord, err := protoutil.ProtoToFile(fileRes.GetFile().GetFile())
	if err != nil {
		return nil, nil, err
	}

	openRes, err := td.fs.client.OpenFile(ctx, &scproto.OpenFileRequest{
		Id:   createRes.GetId(),
		Mode: scproto.FileMode_WRITE,
	})
	if err != nil {
		return nil, nil, err
	}

	handleID, err := uuid.Parse(openRes.GetHandleId())
	if err != nil {
		return nil, nil, err
	}

	td.fs.logger.Debug("tag create: %#v", openRes)

	file := newFSFile(fileRecord, records.FILE_MODE_WRITE, td.fs)
	fileHandle := newFSFileHandle(handleID, file.file.ID, td.fs)

	return file, fileHandle, nil
}

func (td *TagDir) Remove(
	ctx context.Context,
	req *fuse.RemoveRequest,
) error {
	rmDate, rmName, err := splitFullFilename(req.Name)
	if err != nil {
		return err
	}

	protoDate, err := types.TimestampProto(rmDate)
	if err != nil {
		return err
	}

	protoFile, err := td.fs.client.GetFileWithDate(ctx, &scproto.GetFileWithDateRequest{
		Filename:     rmName,
		DocumentDate: protoDate,
	})
	if err != nil && status.Code(err) == codes.NotFound {
		return fuse.ENOENT
	} else if err != nil {
		return err
	}

	_, err = td.fs.client.UpdateFileTags(ctx, &scproto.UpdateFileTagsRequest{
		FileId:      protoFile.File.Id,
		RemovedTags: []string{td.tag},
	})
	if err != nil {
		return err
	}

	return nil
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
		date, err := types.TimestampFromProto(file.GetDocumentDate())
		if err != nil {
			td.fs.logger.Error(
				"could not get date for %s: %s",
				file.GetFilename(), err,
			)
			return nil, err
		}

		entries = append(entries, fuse.Dirent{
			Inode: uint64(idx),
			Type:  fuse.DT_File,
			Name:  getFullFilename(date, file.GetFilename()),
		})
	}

	return entries, nil
}
