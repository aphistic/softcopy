package vault

import (
	"io"
	"testing"

	protomock "github.com/aphistic/softcopy/pkg/proto/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootReadDir(t *testing.T) {
	t.Run("size 0", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		rd := newRootDir(v)

		infos, err := rd.ReadDir(0)
		require.NoError(t, err)
		assert.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "by-date", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "by-tag", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "upload", infos[2].Name())

		infos, err = rd.ReadDir(0)
		assert.Equal(t, io.EOF, err)
		assert.Len(t, infos, 0)
	})

	t.Run("size -10", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		rd := newRootDir(v)

		infos, err := rd.ReadDir(-10)
		require.NoError(t, err)
		assert.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "by-date", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "by-tag", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "upload", infos[2].Name())

		infos, err = rd.ReadDir(-10)
		assert.Equal(t, io.EOF, err)
		assert.Len(t, infos, 0)
	})

	t.Run("size 1", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		rd := newRootDir(v)

		infos, err := rd.ReadDir(1)
		assert.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "by-date", infos[0].Name())

		infos, err = rd.ReadDir(1)
		assert.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "by-tag", infos[0].Name())

		infos, err = rd.ReadDir(1)
		assert.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "upload", infos[0].Name())

		infos, err = rd.ReadDir(1)
		assert.Equal(t, io.EOF, err)
		assert.Len(t, infos, 0)
	})

	t.Run("size 2", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		rd := newRootDir(v)

		infos, err := rd.ReadDir(2)
		assert.NoError(t, err)
		require.Len(t, infos, 2)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "by-date", infos[0].Name())
		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "by-tag", infos[1].Name())

		infos, err = rd.ReadDir(2)
		assert.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "upload", infos[0].Name())

		infos, err = rd.ReadDir(2)
		assert.Equal(t, io.EOF, err)
		assert.Len(t, infos, 0)
	})

	t.Run("exact size", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		rd := newRootDir(v)

		infos, err := rd.ReadDir(3)
		assert.NoError(t, err)
		require.Len(t, infos, 3)

		assert.Equal(t, true, infos[0].IsDir())
		assert.Equal(t, "by-date", infos[0].Name())

		assert.Equal(t, true, infos[1].IsDir())
		assert.Equal(t, "by-tag", infos[1].Name())

		assert.Equal(t, true, infos[2].IsDir())
		assert.Equal(t, "upload", infos[2].Name())

		infos, err = rd.ReadDir(3)
		assert.Equal(t, io.EOF, err)
		assert.Len(t, infos, 0)
	})
}
