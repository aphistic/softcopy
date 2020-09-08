package vault

import (
	"io"
	"testing"

	"github.com/aphistic/goblin"
	protomock "github.com/aphistic/softcopy/pkg/proto/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestByDateDir(v *Vault, path []string) *byDateDir {
	bdd, err := newByDateDir(v, path)
	if err != nil {
		panic(err)
	}
	return bdd
}

func TestReturnPart(t *testing.T) {
	t.Run("size -1, curIdx 0", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		files := []goblin.File{
			newTestByDateDir(v, []string{"2019"}),
			newTestByDateDir(v, []string{"2020"}),
			newTestByDateDir(v, []string{"2021"}),
		}

		newIdx, infos, err := returnPart(-1, 0, files)
		require.NoError(t, err)
		require.Len(t, infos, 3)
		assert.Equal(t, 3, newIdx)
		assert.Equal(t, "2019", infos[0].Name())
		assert.Equal(t, "2020", infos[1].Name())
		assert.Equal(t, "2021", infos[2].Name())
	})
	t.Run("size 0, curIdx 0", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		files := []goblin.File{
			newTestByDateDir(v, []string{"2019"}),
			newTestByDateDir(v, []string{"2020"}),
			newTestByDateDir(v, []string{"2021"}),
		}

		newIdx, infos, err := returnPart(0, 0, files)
		require.NoError(t, err)
		require.Len(t, infos, 3)
		assert.Equal(t, 3, newIdx)
		assert.Equal(t, "2019", infos[0].Name())
		assert.Equal(t, "2020", infos[1].Name())
		assert.Equal(t, "2021", infos[2].Name())
	})
	t.Run("size 1, curIdx 0", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		files := []goblin.File{
			newTestByDateDir(v, []string{"2019"}),
			newTestByDateDir(v, []string{"2020"}),
			newTestByDateDir(v, []string{"2021"}),
		}

		newIdx, infos, err := returnPart(1, 0, files)
		require.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, 1, newIdx)
		assert.Equal(t, "2019", infos[0].Name())
	})
	t.Run("size 1, curIdx 1", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		files := []goblin.File{
			newTestByDateDir(v, []string{"2019"}),
			newTestByDateDir(v, []string{"2020"}),
			newTestByDateDir(v, []string{"2021"}),
		}

		newIdx, infos, err := returnPart(1, 1, files)
		require.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, 2, newIdx)
		assert.Equal(t, "2020", infos[0].Name())
	})
	t.Run("size 1, curIdx 2", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		files := []goblin.File{
			newTestByDateDir(v, []string{"2019"}),
			newTestByDateDir(v, []string{"2020"}),
			newTestByDateDir(v, []string{"2021"}),
		}

		newIdx, infos, err := returnPart(1, 2, files)
		require.NoError(t, err)
		require.Len(t, infos, 1)
		assert.Equal(t, 3, newIdx)
		assert.Equal(t, "2021", infos[0].Name())
	})
	t.Run("size 1, curIdx 3", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		files := []goblin.File{
			newTestByDateDir(v, []string{"2019"}),
			newTestByDateDir(v, []string{"2020"}),
			newTestByDateDir(v, []string{"2021"}),
		}

		newIdx, infos, err := returnPart(1, 3, files)
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, infos)
		assert.Equal(t, 3, newIdx)
	})
}
