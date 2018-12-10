package ftp

import (
	"context"
	"fmt"
	"io"
)

type connectionHandler struct {
	ctx context.Context

	cfg   *config
	state *state

	svc  Service
	conn *ftpConn
}

func newConnectionHandler(ctx context.Context, cfg *config, svc Service, conn *ftpConn) *connectionHandler {
	return &connectionHandler{
		ctx: ctx,

		cfg:   cfg,
		state: &state{},

		svc:  svc,
		conn: conn,
	}
}

func (ch *connectionHandler) Run() {
	conn := ch.conn

	defer conn.Close()

	conn.Response() <- NewResponse(220, "Give it your all!")

	for {
		cmd, ok := <-conn.Command()
		if !ok {
			ch.cfg.logger.Printf("command chan closed\n")
			break
		}
		ch.cfg.logger.Printf("got command: %#v", cmd)
		switch c := cmd.(type) {
		case *BasicCommand:
			switch c.Command() {
			case "ABOR":
				if conn.dataConn != nil {
					conn.dataConn.Close()
					conn.dataConn = nil
				}

				conn.Response() <- NewResponse(226, "Aborted")
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
			case "PWD":
				conn.Response() <- NewResponse(257, "\"/\"")
			case "QUIT":
				conn.Close()
			case "SYST":
				conn.Response() <- NewResponse(215, "UNIX")
			}
		case *CwdCommand:
			fmt.Printf("CWD to %s\n", c.Path)
			conn.Response() <- NewResponse(200, fmt.Sprintf("directory changed to %s", c.Path))
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
			err := ch.svc.Authenticate(ch.state.Username, c.Password)
			if err != nil {
				fmt.Printf("auth error: %s\n", err)
				conn.Response() <- NewResponse(503, "User or pass rejected")
				break
			}

			ch.state.LoggedIn = true
			conn.Response() <- NewResponse(230, "User logged in")
		case *RetrCommand:
			conn.Response() <- NewResponse(550, fmt.Sprintf("%s not found", c.Path))
		case *SizeCommand:
			conn.Response() <- NewResponse(550, fmt.Sprintf("%s not found", c.Path))
		case *StorCommand:
			conn.Response() <- NewResponse(150, fmt.Sprintf("Receiving %s", c.Path))

			var file fileWriter = newMemoryFile()
			if ch.cfg.tempPath != "" {
				f, err := newDiskFile(ch.cfg.tempPath)
				if err != nil {
					fmt.Printf("error creating file: %s", err)
					conn.Response() <- NewResponse(451, fmt.Sprintf("Could not open file for storage"))
					break
				}
				file = f
			}

			buf := make([]byte, 8192)
			for {
				readN, err := conn.dataConn.Read(buf)
				if err == io.EOF {
					break
				} else if err != nil {
					fmt.Printf("error reading data: %s\n", err)
					file.Close()
					continue
				}

				totalWritten := 0
				for {
					writeN, err := file.Write(buf[totalWritten : readN-totalWritten])
					if err != nil {
						fmt.Printf("error writing file: %s", err)
						conn.Response() <- NewResponse(451, fmt.Sprintf("Could not write file"))

						file.Close()

						break
					}
					totalWritten += writeN
					if totalWritten >= readN {
						break
					}
				}
			}

			err := file.CompleteWriting()
			if err != nil {
				fmt.Printf("Error completing writing of file: %s\n", err)
				conn.Response() <- NewResponse(451, fmt.Sprintf("Could not receive file"))
				break
			}

			err = ch.svc.ReceivedFile(c.Path, file)
			if err != nil {
				conn.Response() <- newErrorResponse(
					err,
					451, "could not receive file",
				)
				break
			}

			file.Close()

			conn.Response() <- NewResponse(250, fmt.Sprintf("Received %s", c.Path))
		case *TypeCommand:
			conn.Response() <- NewResponse(200, fmt.Sprintf("TYPE set to %s", c.Type))
		case *UserCommand:
			ch.state.Username = c.User
			conn.Response() <- NewResponse(331, "User OK")
		}
	}
}
