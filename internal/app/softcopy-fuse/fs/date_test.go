package fs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDateYearDirLookup(t *testing.T) {
	t.Run("name with zero prefix", func(t *testing.T) {
		fdyd := newFSDateYearDir(2020, nil)
		n, err := fdyd.Lookup(context.Background(), "09")
		require.NoError(t, err)
		require.NotNil(t, n)
	})
}
