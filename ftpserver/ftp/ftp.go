package ftp

import (
	"fmt"
	"net"
)

type FTPServer struct {
	listener net.Listener
}

func NewFTPServer() *FTPServer {
	return &FTPServer{}
}

func (s *FTPServer) Listen(port int) error {
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
		fmt.Printf("accepted connection from: %s\n", conn.RemoteAddr())
		go s.connectionHandler(newFtpConn(conn))
	}
}

func (s *FTPServer) connectionHandler(conn *ftpConn) {
	defer conn.Close()

	err := conn.Respond(NewResponse(220, "Give it your all!"))
	if err != nil {
		fmt.Printf("error responding with 220: %s\n", err)
		return
	}

	for {
		cmd, ok := <-conn.Command()
		if !ok {
			fmt.Printf("command chan closed\n")
			break
		}
		fmt.Printf("got command: %v\n", cmd)
	}
}
