package ftp

type Logger interface {
	Printf(format string, args ...interface{})
}

type nilLogger struct{}

func newNilLogger() *nilLogger {
	return &nilLogger{}
}

func (nl *nilLogger) Printf(format string, args ...interface{}) {

}
