package logging

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}
