package vault

import (
	"io"
	"os"
	"sort"

	"github.com/aphistic/goblin"
)

func returnPart(n, curIdx int, files []goblin.File) (int, []os.FileInfo, error) {
	if curIdx >= len(files) {
		return curIdx, nil, io.EOF
	}

	var infos []os.FileInfo
	for idx, retDir := range files {
		if idx >= curIdx {
			fi, err := retDir.Stat()
			if err != nil {
				return curIdx + len(infos), infos, err
			}

			infos = append(infos, fi)
		}

		if n > 0 && len(infos) >= n {
			break
		}
	}

	curIdx += len(infos)

	sort.Slice(infos, func(i, j int) bool {
		l := infos[i]
		r := infos[j]
		return l.Name() < r.Name()
	})

	return curIdx, infos, nil
}
