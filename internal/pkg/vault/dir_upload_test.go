package vault

import (
	"io"
	"testing"

	protomock "github.com/aphistic/softcopy/pkg/proto/mock"
	"github.com/stretchr/testify/assert"
)

func TestUploadReadDir(t *testing.T) {
	t.Run("returns empty with size -1", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		ud := newUploadDir(v)

		infos, err := ud.ReadDir(-1)
		assert.NoError(t, err)
		assert.Len(t, infos, 0)
	})
	t.Run("returns empty with size 0", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		ud := newUploadDir(v)

		infos, err := ud.ReadDir(0)
		assert.NoError(t, err)
		assert.Len(t, infos, 0)
	})
	t.Run("returns empty with size 1", func(t *testing.T) {
		v := NewVault(protomock.NewMockSoftcopyClient())
		ud := newUploadDir(v)

		infos, err := ud.ReadDir(1)
		assert.Equal(t, io.EOF, err)
		assert.Len(t, infos, 0)
	})
}
