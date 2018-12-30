package ftpserver

import (
	"github.com/efritz/nacelle"
)

type logger struct {
	logger nacelle.Logger
}

func (l *logger) Printf(format string, args ...interface{}) {
	l.logger.Debug(format, args...)
}
