package logging

type NilLogger struct{}

var _ Logger = &NilLogger{}

func NewNilLogger() *NilLogger {
	return &NilLogger{}
}

func (nl *NilLogger) Debug(format string, args ...interface{}) {
	// Skip logging
}

func (nl *NilLogger) Info(format string, args ...interface{}) {
	// Skip logging
}

func (nl *NilLogger) Error(format string, args ...interface{}) {
	// Skip logging
}
