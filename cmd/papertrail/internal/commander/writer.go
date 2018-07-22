package commander

import (
	"fmt"
	"io"
	"os"
)

type Writer interface {
	io.Writer

	Printf(format string, args ...interface{})
}

type consoleWriter struct{}

func (cw *consoleWriter) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

func (cw *consoleWriter) Printf(format string, args ...interface{}) {
	out := []byte(fmt.Sprintf(format, args...))

	cur := 0
	for {
		n, err := cw.Write(out[cur:])
		cur += n
		if err != nil {
			return
		}
		if cur >= len(out) {
			break
		}
	}
}
