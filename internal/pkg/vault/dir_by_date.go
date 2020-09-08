package vault

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/aphistic/goblin"
	scproto "github.com/aphistic/softcopy/pkg/proto"
)

type byDateDir struct {
	v    *Vault
	path []string

	dirIdx int

	dateYear  int
	dateMonth int
	dateDay   int
}

var _ goblin.ReadDirFile = &byDateDir{}

func newByDateDir(v *Vault, path []string) (*byDateDir, error) {
	if len(path) > 3 {
		return nil, ErrInvalidPath
	}

	var dateYear, dateMonth, dateDay int
	if len(path) >= 1 {
		year, err := strconv.ParseInt(path[0], 10, 0)
		if err != nil {
			return nil, ErrInvalidPath
		}
		dateYear = int(year)
	}
	if len(path) >= 2 {
		month, err := strconv.ParseInt(path[1], 10, 0)
		if err != nil {
			return nil, ErrInvalidPath
		}
		dateMonth = int(month)
	}
	if len(path) >= 3 {
		day, err := strconv.ParseInt(path[2], 10, 0)
		if err != nil {
			return nil, err
		}
		dateDay = int(day)
	}

	return &byDateDir{
		v:    v,
		path: path,

		dateYear:  dateYear,
		dateMonth: dateMonth,
		dateDay:   dateDay,
	}, nil
}

func (bdd *byDateDir) Stat() (os.FileInfo, error) {
	name := byDatePath
	switch {
	case bdd.dateDay != 0:
		name = fmt.Sprintf("%02d", bdd.dateDay)
	case bdd.dateMonth != 0:
		name = fmt.Sprintf("%02d", bdd.dateMonth)
	case bdd.dateYear != 0:
		name = fmt.Sprintf("%04d", bdd.dateYear)
	}

	return &dirFileInfo{
		name: name,
		sys:  bdd,
	}, nil
}

func (bdd *byDateDir) Read(buf []byte) (int, error) {
	return 0, fmt.Errorf("cannot read: file is a directory")
}

func (bdd *byDateDir) Close() error {
	bdd.dirIdx = 0
	return nil
}

func (bdd *byDateDir) ReadDir(n int) ([]os.FileInfo, error) {
	ctx := context.Background()
	switch {
	case bdd.dateYear != 0 && bdd.dateMonth != 0 && bdd.dateDay != 0:
		return bdd.readDirDay(ctx, n)
	case bdd.dateYear != 0 && bdd.dateMonth != 0:
		return bdd.readDirMonth(ctx, n)
	case bdd.dateYear != 0:
		return bdd.readDirYear(ctx, n)
	default:
		return bdd.readDirRoot(ctx, n)
	}
}

func (bdd *byDateDir) readDirRoot(ctx context.Context, n int) ([]os.FileInfo, error) {
	res, err := bdd.v.client.GetFileYears(ctx, &scproto.GetFileYearsRequest{})
	if err != nil {
		return nil, err
	}

	yearDirs := make([]goblin.File, 0, len(res.GetYears()))
	for _, year := range res.GetYears() {
		yearDir, err := newByDateDir(bdd.v, []string{strconv.Itoa(int(year))})
		if err != nil {
			return nil, err
		}
		yearDirs = append(yearDirs, yearDir)
	}

	newIdx, infos, err := returnPart(n, bdd.dirIdx, yearDirs)
	bdd.dirIdx = newIdx

	return infos, err
}

func (bdd *byDateDir) readDirYear(ctx context.Context, n int) ([]os.FileInfo, error) {
	res, err := bdd.v.client.GetFileMonths(ctx, &scproto.GetFileMonthsRequest{
		Year: int32(bdd.dateYear),
	})
	if err != nil {
		return nil, err
	}

	monthDirs := make([]goblin.File, 0, len(res.GetMonths()))
	for _, month := range res.GetMonths() {
		monthDir, err := newByDateDir(bdd.v, []string{
			fmt.Sprintf("%04d", bdd.dateYear),
			fmt.Sprintf("%02d", month),
		})
		if err != nil {
			return nil, err
		}
		monthDirs = append(monthDirs, monthDir)
	}

	newIdx, infos, err := returnPart(n, bdd.dirIdx, monthDirs)
	bdd.dirIdx = newIdx

	return infos, err
}

func (bdd *byDateDir) readDirMonth(ctx context.Context, n int) ([]os.FileInfo, error) {
	res, err := bdd.v.client.GetFileDays(ctx, &scproto.GetFileDaysRequest{
		Year:  int32(bdd.dateYear),
		Month: int32(bdd.dateMonth),
	})
	if err != nil {
		return nil, err
	}

	dayDirs := make([]goblin.File, 0, len(res.GetDays()))
	for _, day := range res.GetDays() {
		dayDir, err := newByDateDir(bdd.v, []string{
			fmt.Sprintf("%04d", bdd.dateYear),
			fmt.Sprintf("%02d", bdd.dateMonth),
			fmt.Sprintf("%02d", day),
		})
		if err != nil {
			return nil, err
		}
		dayDirs = append(dayDirs, dayDir)
	}

	newIdx, infos, err := returnPart(n, bdd.dirIdx, dayDirs)
	bdd.dirIdx = newIdx

	return infos, err
}

func (bdd *byDateDir) readDirDay(ctx context.Context, n int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}
