package logging

import (
	"github.com/aphistic/gomol"
	console "github.com/aphistic/gomol-console"
)

type GomolLogger struct {
	base *gomol.Base
}

func NewGomolLogger() (*GomolLogger, error) {
	base := gomol.NewBase()
	base.SetLogLevel(gomol.LevelDebug)

	gcConfig := console.NewConsoleLoggerConfig()
	gc, err := console.NewConsoleLogger(gcConfig)
	if err != nil {
		return nil, err
	}

	err = base.AddLogger(gc)
	if err != nil {
		return nil, err
	}

	err = base.InitLoggers()
	if err != nil {
		return nil, err
	}

	return &GomolLogger{
		base: base,
	}, nil
}

func (gl *GomolLogger) Debug(format string, args ...interface{}) {
	gl.base.Debugf(format, args...)
}

func (gl *GomolLogger) Info(format string, args ...interface{}) {
	gl.base.Infof(format, args...)
}

func (gl *GomolLogger) Error(format string, args ...interface{}) {
	gl.base.Errorf(format, args...)
}

func (gl *GomolLogger) Shutdown() error {
	return gl.base.ShutdownLoggers()
}
