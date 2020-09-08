package vault

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	scproto "github.com/aphistic/softcopy/pkg/proto"
	protomock "github.com/aphistic/softcopy/pkg/proto/mock"
)

func TestByDateDirNew(t *testing.T) {
	t.Run("root path", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		bdd, err := newByDateDir(v, []string{})
		assert.NoError(t, err)
		assert.Equal(t, 0, bdd.dateYear)
		assert.Equal(t, 0, bdd.dateMonth)
		assert.Equal(t, 0, bdd.dateDay)
	})
	t.Run("valid year", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		bdd, err := newByDateDir(v, []string{"2020"})
		assert.NoError(t, err)
		assert.Equal(t, 2020, bdd.dateYear)
		assert.Equal(t, 0, bdd.dateMonth)
		assert.Equal(t, 0, bdd.dateDay)
	})
	t.Run("valid month", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		bdd, err := newByDateDir(v, []string{"2020", "09"})
		assert.NoError(t, err)
		assert.Equal(t, 2020, bdd.dateYear)
		assert.Equal(t, 9, bdd.dateMonth)
		assert.Equal(t, 0, bdd.dateDay)
	})
	t.Run("valid day", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		bdd, err := newByDateDir(v, []string{"2020", "09", "19"})
		assert.NoError(t, err)
		assert.Equal(t, 2020, bdd.dateYear)
		assert.Equal(t, 9, bdd.dateMonth)
		assert.Equal(t, 19, bdd.dateDay)
	})
	t.Run("too long", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		bdd, err := newByDateDir(v, []string{"2020", "09", "19", "1234"})
		assert.Equal(t, ErrInvalidPath, err)
		assert.Nil(t, bdd)
	})
}

func TestByDateDirRoot(t *testing.T) {
	t.Run("size -1", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileYearsFunc.PushHook(
			func(
				ctx context.Context,
				req *scproto.GetFileYearsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileYearsResponse, error) {
				assert.Equal(t, &scproto.GetFileYearsRequest{}, req)

				return &scproto.GetFileYearsResponse{
					Years: []int32{1990, 2020, 2131},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(-1)
		require.NoError(t, err)
		require.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "1990", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "2020", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "2131", infos[2].Name())

		infos, err = bdd.ReadDir(-1)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size 0", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileYearsFunc.PushHook(
			func(
				ctx context.Context,
				req *scproto.GetFileYearsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileYearsResponse, error) {
				assert.Equal(t, &scproto.GetFileYearsRequest{}, req)

				return &scproto.GetFileYearsResponse{
					Years: []int32{1990, 2020, 2131},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(0)
		require.NoError(t, err)
		require.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "1990", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "2020", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "2131", infos[2].Name())

		infos, err = bdd.ReadDir(0)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size 2", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileYearsFunc.SetDefaultHook(
			func(
				ctx context.Context,
				req *scproto.GetFileYearsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileYearsResponse, error) {
				assert.Equal(t, &scproto.GetFileYearsRequest{}, req)

				return &scproto.GetFileYearsResponse{
					Years: []int32{1990, 2020, 2131},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(2)
		require.NoError(t, err)
		require.Len(t, infos, 2)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "1990", infos[0].Name())
		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "2020", infos[1].Name())

		infos, err = bdd.ReadDir(2)
		require.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "2131", infos[0].Name())

		infos, err = bdd.ReadDir(2)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size exact", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileYearsFunc.SetDefaultHook(
			func(
				ctx context.Context,
				req *scproto.GetFileYearsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileYearsResponse, error) {
				assert.Equal(t, &scproto.GetFileYearsRequest{}, req)

				return &scproto.GetFileYearsResponse{
					Years: []int32{1990, 2020, 2131},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(3)
		require.NoError(t, err)
		require.Len(t, infos, 3)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "1990", infos[0].Name())
		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "2020", infos[1].Name())
		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "2131", infos[2].Name())

		infos, err = bdd.ReadDir(3)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})
}

func TestByDateDirYear(t *testing.T) {
	t.Run("size -1", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileMonthsFunc.PushHook(
			func(
				ctx context.Context,
				req *scproto.GetFileMonthsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileMonthsResponse, error) {
				assert.Equal(t, &scproto.GetFileMonthsRequest{
					Year: 2020,
				}, req)

				return &scproto.GetFileMonthsResponse{
					Months: []int32{1, 2, 3},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(-1)
		require.NoError(t, err)
		require.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "01", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "02", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "03", infos[2].Name())

		infos, err = bdd.ReadDir(-1)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size 0", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileMonthsFunc.PushHook(
			func(
				ctx context.Context,
				req *scproto.GetFileMonthsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileMonthsResponse, error) {
				assert.Equal(t, &scproto.GetFileMonthsRequest{
					Year: 2020,
				}, req)

				return &scproto.GetFileMonthsResponse{
					Months: []int32{1, 2, 3},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(0)
		require.NoError(t, err)
		require.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "01", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "02", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "03", infos[2].Name())

		infos, err = bdd.ReadDir(0)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size 2", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileMonthsFunc.SetDefaultHook(
			func(
				ctx context.Context,
				req *scproto.GetFileMonthsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileMonthsResponse, error) {
				assert.Equal(t, &scproto.GetFileMonthsRequest{
					Year: 2020,
				}, req)

				return &scproto.GetFileMonthsResponse{
					Months: []int32{1, 2, 3},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(2)
		require.NoError(t, err)
		require.Len(t, infos, 2)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "01", infos[0].Name())
		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "02", infos[1].Name())

		infos, err = bdd.ReadDir(2)
		require.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "03", infos[0].Name())

		infos, err = bdd.ReadDir(2)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size exact", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileMonthsFunc.PushHook(
			func(
				ctx context.Context,
				req *scproto.GetFileMonthsRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileMonthsResponse, error) {
				assert.Equal(t, &scproto.GetFileMonthsRequest{
					Year: 2020,
				}, req)

				return &scproto.GetFileMonthsResponse{
					Months: []int32{1, 2, 3},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(3)
		require.NoError(t, err)
		require.Len(t, infos, 3)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "01", infos[0].Name())
		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "02", infos[1].Name())
		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "03", infos[2].Name())

		infos, err = bdd.ReadDir(3)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})
}

func TestByDateDirMonth(t *testing.T) {
	t.Run("size -1", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileDaysFunc.SetDefaultHook(
			func(
				ctx context.Context,
				req *scproto.GetFileDaysRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileDaysResponse, error) {
				assert.Equal(t, &scproto.GetFileDaysRequest{
					Year:  2020,
					Month: 9,
				}, req)

				return &scproto.GetFileDaysResponse{
					Days: []int32{4, 5, 6},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020", "09"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(-1)
		require.NoError(t, err)
		require.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "04", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "05", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "06", infos[2].Name())

		infos, err = bdd.ReadDir(-1)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size 0", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileDaysFunc.SetDefaultHook(
			func(
				ctx context.Context,
				req *scproto.GetFileDaysRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileDaysResponse, error) {
				assert.Equal(t, &scproto.GetFileDaysRequest{
					Year:  2020,
					Month: 9,
				}, req)

				return &scproto.GetFileDaysResponse{
					Days: []int32{4, 5, 6},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020", "09"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(0)
		require.NoError(t, err)
		require.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "04", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "05", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "06", infos[2].Name())

		infos, err = bdd.ReadDir(0)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size 2", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileDaysFunc.SetDefaultHook(
			func(
				ctx context.Context,
				req *scproto.GetFileDaysRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileDaysResponse, error) {
				assert.Equal(t, &scproto.GetFileDaysRequest{
					Year:  2020,
					Month: 9,
				}, req)

				return &scproto.GetFileDaysResponse{
					Days: []int32{4, 5, 6},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020", "09"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(2)
		require.NoError(t, err)
		require.Len(t, infos, 2)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "04", infos[0].Name())
		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "05", infos[1].Name())

		infos, err = bdd.ReadDir(2)
		require.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "06", infos[0].Name())

		infos, err = bdd.ReadDir(2)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})

	t.Run("size exact", func(t *testing.T) {
		scc := protomock.NewMockSoftcopyClient()
		scc.GetFileDaysFunc.SetDefaultHook(
			func(
				ctx context.Context,
				req *scproto.GetFileDaysRequest,
				opts ...grpc.CallOption,
			) (*scproto.GetFileDaysResponse, error) {
				assert.Equal(t, &scproto.GetFileDaysRequest{
					Year:  2020,
					Month: 9,
				}, req)

				return &scproto.GetFileDaysResponse{
					Days: []int32{4, 5, 6},
				}, nil
			},
		)

		v := NewVault(scc)
		bdd, err := newByDateDir(v, []string{"2020", "09"})
		require.NoError(t, err)

		infos, err := bdd.ReadDir(3)
		require.NoError(t, err)
		require.Len(t, infos, 3)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "04", infos[0].Name())
		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "05", infos[1].Name())
		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "06", infos[2].Name())

		infos, err = bdd.ReadDir(3)
		require.Equal(t, io.EOF, err)
		require.Nil(t, infos)
	})
}
