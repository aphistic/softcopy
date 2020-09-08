package fs

import (
	"context"
	"testing"

	"bazil.org/fuse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDateMonthDirLookup(t *testing.T) {
	t.Run("name with zero prefix", func(t *testing.T) {
		fdmd := newFSDateMonthDir(2020, 9, nil)
		n, err := fdmd.Lookup(context.Background(), "09")
		require.NoError(t, err)
		require.NotNil(t, n)
	})
	t.Run("name that is not a number", func(t *testing.T) {
		fdmd := newFSDateMonthDir(2020, 9, nil)
		n, err := fdmd.Lookup(context.Background(), ".git")
		require.EqualValues(t, fuse.ENOENT, err)
		assert.Nil(t, n)
	})
}
