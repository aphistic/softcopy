package ftpserver

import (
	"fmt"

	"github.com/aphistic/papertrail/ftpserver/ftp"
	"io/ioutil"
)

type ftpService struct{}

func (fs *ftpService) Init() error {
	return nil
}
func (fs *ftpService) Close() error {
	return nil
}

func (fs *ftpService) Authenticate(username, password string) error {
	return nil
}

func (fs *ftpService) ReceivedFile(name string, file ftp.File) error {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("read err: %s\n", err)
		return err
	}

	fmt.Printf("Contents:\n%s\n", data)

	return nil
}
