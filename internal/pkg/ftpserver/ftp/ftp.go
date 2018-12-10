package ftp

// See https://cr.yp.to/ftp.html for some useful FTP protocol info
// FTP RFC: https://tools.ietf.org/html/rfc959
// FTP Extensions:
// https://tools.ietf.org/html/rfc2428
// https://tools.ietf.org/html/rfc3659

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

type DataType int

func (dt DataType) String() string {
	switch dt {
	case TypeASCII:
		return "ASCII"
	case TypeEBCDIC:
		return "EBCDIC"
	case TypeImage:
		return "Image"
	case TypeLocal:
		return "Local"
	default:
		return "Unknown"
	}
}

const (
	TypeASCII = iota
	TypeEBCDIC
	TypeImage
	TypeLocal
)

type FTPOption func(c *config)

// FTPTempPath sets a local path to use for temporary file uploads
func FTPTempPath(tempPath string) FTPOption {
	return func(c *config) {
		c.tempPath = tempPath
	}
}

// FTPRandomTempPath picks a random path
func FTPRandomTempPath(prefix string) FTPOption {
	return func(c *config) {
		c.randomTempPath = true
		c.randomTempPathPrefix = prefix
	}
}

func FTPLogger(logger Logger) FTPOption {
	return func(c *config) {
		c.logger = logger
		if c.logger == nil {
			c.logger = newNilLogger()
		}
	}
}

type FTPServer struct {
	serviceFactory func() Service

	listener net.Listener

	cfg *config
}

func NewFTPServer(serviceFactory func() Service, opts ...FTPOption) (*FTPServer, error) {
	if serviceFactory == nil {
		return nil, fmt.Errorf("service factory cannot be nil")
	}

	fs := &FTPServer{
		serviceFactory: serviceFactory,
		cfg: &config{
			logger: newNilLogger(),
		},
	}

	for _, f := range opts {
		f(fs.cfg)
	}

	return fs, nil
}

func (s *FTPServer) Close() error {
	if s.cfg.randomTempPath {
		err := os.RemoveAll(s.cfg.tempPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *FTPServer) Listen(port int) error {
	if s.cfg.randomTempPath {
		path, err := ioutil.TempDir("", s.cfg.randomTempPathPrefix)
		if err != nil {
			return err
		}
		s.cfg.tempPath = path
	}
	if s.cfg.tempPath != "" {
		// If we have a temp path set, make sure:
		// 1. It exists or that we can create it if it doesn't
		// 2. It is writable
		fi, err := os.Stat(s.cfg.tempPath)
		if os.IsNotExist(err) {
			err := os.MkdirAll(s.cfg.tempPath, 0755)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if !fi.IsDir() {
			return fmt.Errorf("%s is not a directory", s.cfg.tempPath)
		}

		// The directory exists now, try writing a file to it to see
		// if it's writable
		tempFile, err := ioutil.TempFile(s.cfg.tempPath, "")
		if err != nil {
			fmt.Printf("path err: %s\n", err)
			return err
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		n, err := tempFile.WriteString("testing!")
		if err != nil {
			return err
		}
		if n < 1 {
			return fmt.Errorf("failed to write to temp directory")
		}
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	s.listener = listener

	go s.worker()

	return nil
}

func (s *FTPServer) worker() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("error accepting conn: %s\n", err)
			continue
		}
		s.cfg.logger.Printf("accepted connection from: %s", conn.RemoteAddr())

		ch := newConnectionHandler(
			context.Background(),
			s.cfg,
			s.serviceFactory(),
			newFtpConn(conn, s.cfg.logger),
		)

		go ch.Run()
	}
}
