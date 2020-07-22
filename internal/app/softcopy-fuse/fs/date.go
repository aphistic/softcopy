package fs

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"bazil.org/fuse"
	fusefs "bazil.org/fuse/fs"
	"github.com/gogo/protobuf/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/aphistic/softcopy/internal/pkg/protoutil"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

type fsByDateDir struct {
	fs *FileSystem
}

func newFSByDateDir(fs *FileSystem) *fsByDateDir {
	return &fsByDateDir{
		fs: fs,
	}
}

func (bdd *fsByDateDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (bdd *fsByDateDir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	year, err := strconv.ParseInt(name, 0, 0)
	if err != nil {
		return nil, err
	}

	return newFSDateYearDir(int(year), bdd.fs), nil
}

func (bdd *fsByDateDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	res, err := bdd.fs.client.GetFileYears(ctx, &scproto.GetFileYearsRequest{})
	if err != nil {
		bdd.fs.logger.Error("error in by-date readdirall: %s", err)
		return nil, err
	}

	entries := []fuse.Dirent{}
	for _, year := range res.GetYears() {
		entries = append(entries, fuse.Dirent{
			Inode: 1,
			Type:  fuse.DT_Dir,
			Name:  fmt.Sprintf("%d", year),
		})
	}

	return entries, nil
}

type fsDateYearDir struct {
	fs   *FileSystem
	year int
}

func newFSDateYearDir(year int, fs *FileSystem) *fsDateYearDir {
	return &fsDateYearDir{
		fs:   fs,
		year: year,
	}
}

func (dyd *fsDateYearDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (dyd *fsDateYearDir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	month, err := strconv.ParseInt(name, 0, 0)
	if err != nil {
		return nil, err
	}

	return newFSDateMonthDir(dyd.year, int(month), dyd.fs), nil
}

func (dyd *fsDateYearDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	res, err := dyd.fs.client.GetFileMonths(ctx, &scproto.GetFileMonthsRequest{
		Year: int32(dyd.year),
	})
	if err != nil {
		dyd.fs.logger.Error("error in year readdirall: %s", err)
		return nil, err
	}

	entries := []fuse.Dirent{}
	for _, month := range res.GetMonths() {
		entries = append(entries, fuse.Dirent{
			Inode: 1,
			Type:  fuse.DT_Dir,
			Name:  fmt.Sprintf("%02d", month),
		})
	}

	return entries, nil
}

type fsDateMonthDir struct {
	fs    *FileSystem
	year  int
	month int
}

func newFSDateMonthDir(year int, month int, fs *FileSystem) *fsDateMonthDir {
	return &fsDateMonthDir{
		fs:    fs,
		year:  year,
		month: month,
	}
}

func (dmd *fsDateMonthDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (dmd *fsDateMonthDir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	day, err := strconv.ParseInt(name, 0, 0)
	if err != nil {
		return nil, err
	}

	return newFSDateDayDir(dmd.year, dmd.month, int(day), dmd.fs), nil
}

func (dmd *fsDateMonthDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	res, err := dmd.fs.client.GetFileDays(ctx, &scproto.GetFileDaysRequest{
		Year:  int32(dmd.year),
		Month: int32(dmd.month),
	})
	if err != nil {
		dmd.fs.logger.Error("error in month readdirall: %s", err)
		return nil, err
	}

	entries := []fuse.Dirent{}
	for _, day := range res.GetDays() {
		entries = append(entries, fuse.Dirent{
			Inode: 1,
			Type:  fuse.DT_Dir,
			Name:  fmt.Sprintf("%02d", day),
		})
	}

	return entries, nil
}

type fsDateDayDir struct {
	fs    *FileSystem
	year  int
	month int
	day   int
}

func newFSDateDayDir(year int, month int, day int, fs *FileSystem) *fsDateDayDir {
	return &fsDateDayDir{
		fs:    fs,
		year:  year,
		month: month,
		day:   day,
	}
}

func (ddd *fsDateDayDir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Mode = os.ModeDir | 0555
	return nil
}

func (ddd *fsDateDayDir) Lookup(ctx context.Context, name string) (fusefs.Node, error) {
	ts, err := types.TimestampProto(time.Date(
		ddd.year, time.Month(ddd.month), ddd.day,
		0, 0, 0, 0, time.Local,
	))
	if err != nil {
		return nil, err
	}

	res, err := ddd.fs.client.GetFileWithDate(ctx, &scproto.GetFileWithDateRequest{
		DocumentDate: ts,
		Filename:     name,
	})
	if status.Code(err) == codes.NotFound {
		return nil, fuse.ENOENT
	} else if err != nil {
		return nil, err
	}

	file, err := protoutil.ProtoToFile(res.GetFile())
	if err != nil {
		return nil, err
	}

	return newFSFile(file, records.FILE_MODE_READ, ddd.fs), nil
}

func (td *fsDateDayDir) Rename(
	ctx context.Context,
	req *fuse.RenameRequest,
	newDir fusefs.Node,
) error {
	docDate, err := types.TimestampProto(
		time.Date(td.year, time.Month(td.month), td.day, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		return err
	}

	res, err := td.fs.client.GetFileWithDate(ctx, &scproto.GetFileWithDateRequest{
		Filename:     req.OldName,
		DocumentDate: docDate,
	})
	if status.Code(err) == codes.NotFound {
		return fuse.ENOENT
	} else if err != nil {
		return err
	}

	newDateDir, ok := newDir.(*fsDateDayDir)
	if !ok {
		return fuse.ENOTSUP
	}

	newDocDate, err := types.TimestampProto(
		time.Date(
			newDateDir.year, time.Month(newDateDir.month), newDateDir.day,
			0, 0, 0, 0, time.UTC,
		),
	)

	_, err = td.fs.client.UpdateFileDate(ctx, &scproto.UpdateFileDateRequest{
		FileId:          res.GetFile().GetId(),
		NewFilename:     req.NewName,
		NewDocumentDate: newDocDate,
	})
	if err != nil {
		return err
	}

	return nil
}

func (ddd *fsDateDayDir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	ts, err := types.TimestampProto(time.Date(
		ddd.year, time.Month(ddd.month), ddd.day,
		0, 0, 0, 0, time.Local,
	))
	if err != nil {
		return nil, err
	}

	res, err := ddd.fs.client.FindFilesWithDate(ctx, &scproto.FindFilesWithDateRequest{
		DocumentDate: ts,
	})
	if err != nil {
		return nil, err
	}

	entries := []fuse.Dirent{}
	for _, file := range res.GetFiles() {
		entries = append(entries, fuse.Dirent{
			Inode: 1,
			Type:  fuse.DT_File,
			Name:  file.GetFilename(),
		})
	}

	return entries, nil
}

func (ddd *fsDateDayDir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	ts, err := types.TimestampProto(time.Date(
		ddd.year, time.Month(ddd.month), ddd.day,
		0, 0, 0, 0, time.Local,
	))
	if err != nil {
		return err
	}

	fileRes, err := ddd.fs.client.GetFileWithDate(ctx, &scproto.GetFileWithDateRequest{
		Filename:     req.Name,
		DocumentDate: ts,
	})
	if status.Code(err) == codes.NotFound {
		return fuse.ENOENT
	} else if err != nil {
		return err
	}

	_, err = ddd.fs.client.RemoveFile(ctx, &scproto.RemoveFileRequest{
		Id: fileRes.GetFile().GetId(),
	})
	if status.Code(err) == codes.NotFound {
		return fuse.ENOENT
	} else if err != nil {
		return err
	}

	return nil
}
