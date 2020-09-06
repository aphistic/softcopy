package fs

import (
	"context"
	"os"
	"testing"

	"bazil.org/fuse"
	"github.com/aphistic/softcopy/internal/pkg/storage/records"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileAttr(t *testing.T) {
	t.Run("read only file mode", func(t *testing.T) {
		f := newFSFile(
			&records.File{Size: 1234},
			records.FILE_MODE_READ,
			nil,
		)

		attr := &fuse.Attr{}

		err := f.Attr(context.Background(), attr)
		require.NoError(t, err)

		assert.Equal(t, os.FileMode(0444), attr.Mode)
		assert.Equal(t, uint64(1234), attr.Size)
	})
}
