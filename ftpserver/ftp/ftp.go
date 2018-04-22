package ftp

// See https://cr.yp.to/ftp.html for some useful FTP protocol info
// FTP RFC: https://tools.ietf.org/html/rfc959
// FTP Extensions: https://tools.ietf.org/html/rfc2428

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

	conn.Response() <- NewResponse(220, "Give it your all!")

	for {
		cmd, ok := <-conn.Command()
		if !ok {
			fmt.Printf("command chan closed\n")
			break
		}
		fmt.Printf("got command: %v\n", cmd)
		switch c := cmd.(type) {
		case *BasicCommand:
			switch c.Command() {
			case "EPSV":
				pc, err := bindPassiveConn()
				if err != nil {
					fmt.Printf("error binding passive listener: %s\n", err)
					break
				}
				conn.dataConn = pc

				conn.Response() <- NewResponse(229, fmt.Sprintf("Entering EPSV mode (|||%d|)", pc.Port()))
			case "LIST":
				conn.Response() <- NewResponse(150, "Here comes the directory listing")
				simList := []byte("list of files\r\n")
				_, err := conn.dataConn.Write(simList)
				if err != nil {
					fmt.Printf("error writing files: %s\n", err)
					break
				}
				conn.CloseData()
				conn.Response() <- NewResponse(226, "Directory send OK")
			case "QUIT":
				conn.Close()
			case "SYST":
				conn.Response() <- NewResponse(215, "UNIX")
			}
		case *EprtCommand:
			addr := fmt.Sprintf("%s", c.Address)
			if c.Version == 6 {
				addr = fmt.Sprintf("[%s]", addr)
			}

			dConn, err := dialActiveConn(addr, c.Port)
			if err != nil {
				fmt.Printf("Error dialing data conn: %s\n", err)
			}
			conn.dataConn = dConn

			conn.Response() <- NewResponse(200, "EPRT command successful")
		case *PassCommand:
			conn.Response() <- NewResponse(230, "User logged in")
		case *UserCommand:
			conn.Response() <- NewResponse(331, "User ok")
		}
	}
}
