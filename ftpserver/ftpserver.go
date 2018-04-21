package ftpserver

import (
	"github.com/efritz/nacelle"

	"github.com/aphistic/papertrail/ftpserver/ftp"
)

type Process struct {
	Logger nacelle.Logger `service:"logger"`

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
	fs := ftp.NewFTPServer()
	err := fs.Listen(8021)
	if err != nil {
		return err
	}

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
