package ftpserver

import (
	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/api"
	"github.com/aphistic/softcopy/ftpserver/ftp"
)

type Process struct {
	Logger nacelle.Logger `service:"logger"`
	API    *api.Client    `service:"api"`

	stopChan chan struct{}
}

func NewProcess() *Process {
	return &Process{
		stopChan: make(chan struct{}),
	}
}

func (p *Process) Init(config nacelle.Config) error {
	return nil
}

func (p *Process) Start() error {
	fs, err := ftp.NewFTPServer(
		func() ftp.Service {
			return newFTPService(p.API, p.Logger)
		},
		ftp.FTPRandomTempPath("softcopy-"),
		ftp.FTPLogger(&logger{
			logger: p.Logger,
		}),
	)
	if err != nil {
		return err
	}

	err = fs.Listen(8021)
	if err != nil {
		return err
	}
	defer fs.Close()

mainLoop:
	for {
		select {
		case <-p.stopChan:
			p.Logger.Info("Stopping FTP Server")
			break mainLoop
		}
	}
	return nil
}

func (p *Process) Stop() error {
	select {
	case <-p.stopChan:
	default:
		close(p.stopChan)
	}
	return nil
}
