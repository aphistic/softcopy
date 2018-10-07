package ftpserver

import (
	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/api"
	"github.com/aphistic/softcopy/ftpserver/ftp"
)

type ftpService struct {
	logger nacelle.Logger
	api    *api.Client
}

func newFTPService(api *api.Client, logger nacelle.Logger) *ftpService {
	return &ftpService{
		logger: logger,
		api:    api,
	}
}

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
	err := fs.api.AddFile(name, file)
	if err == api.ErrHashCollision {
		fs.logger.Error("FTP add failed: hash already exists")
		return ftp.NewFTPError(451, "file hash already exists")
	} else if err != nil {
		fs.logger.Error("Unknown error receiving file: %s", err)
		return err
	}

	fs.logger.Info("FTP added new file: %s", name)

	return nil
}
