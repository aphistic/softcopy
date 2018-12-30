package ftp

type Service interface {
	Init() error
	Close() error

	Authenticate(username, password string) error

	ReceivedFile(name string, file File) error
}
