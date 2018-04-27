package ftp

import (
	"fmt"
	"io"
	"net"
)

type ftpConn struct {
	ctlConn  net.Conn
	dataConn dataConn

	commands  chan Command
	responses chan *Response
}

func newFtpConn(ctlConn net.Conn) *ftpConn {
	fc := &ftpConn{
		ctlConn:   ctlConn,
		commands:  make(chan Command),
		responses: make(chan *Response),
	}

	go fc.responseWorker()
	go fc.commandWorker()

	return fc
}

func (c *ftpConn) CloseData() error {
	dc := c.dataConn
	c.dataConn = nil

	return dc.Close()
}

func (c *ftpConn) Close() error {
	return c.ctlConn.Close()
}

func (c *ftpConn) Command() <-chan Command {
	return c.commands
}

func (c *ftpConn) Response() chan<- *Response {
	return c.responses
}

func (c *ftpConn) responseWorker() {
	for {
		select {
		case res := <-c.responses:
			fmt.Printf("got response\n")
			data, err := res.MarshalText()
			if err != nil {
				fmt.Printf("error marshaling response text: %s\n", err)
				break
			}
			data = append(data, []byte{'\r', '\n'}...)

			fmt.Printf("Writing\n")
			n, err := c.ctlConn.Write(data)
			if err != nil {
				fmt.Printf("error writing response: %s\n", err)
				break
			} else if n != len(data) {
				fmt.Printf("short write on ftp response\n")
				break
			}
		}
	}
}

func (c *ftpConn) commandWorker() {
	buf := make([]byte, 4096)
	for {
		n, err := c.ctlConn.Read(buf)
		if netErr, ok := err.(net.Error); ok {
			if !netErr.Temporary() {
				return
			}
		} else if err == io.EOF {
			fmt.Printf("eof on ftp conn\n")
			c.ctlConn.Close()
			return
		} else if err != nil {
			fmt.Printf("error reading ftp conn: %s\n", err)
			continue
		}
		fmt.Printf("read from ftp conn: %s\n", buf[0:n])
		cmd, err := parseCommand(buf[0:n])
		if err != nil {
			fmt.Printf("error reading command: %s\n", err)
			continue
		}
		c.commands <- cmd
	}

}

type dataConn interface {
	io.Writer
	io.Reader
	io.Closer
}

type activeDataConn struct {
	conn net.Conn
}

func dialActiveConn(addr string, port int) (*activeDataConn, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return nil, err
	}

	ac := &activeDataConn{
		conn: conn,
	}

	return ac, nil
}

func (ac *activeDataConn) Read(b []byte) (int, error) {
	return ac.conn.Read(b)
}

func (ac *activeDataConn) Write(b []byte) (int, error) {
	return ac.conn.Write(b)
}

func (ac *activeDataConn) Close() error {
	return ac.conn.Close()
}

type passiveDataConn struct {
	listener net.Listener
	tcpAddr  *net.TCPAddr
	conn     net.Conn
}

func bindPassiveConn() (*passiveDataConn, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("listener addr is not tcp")
	}

	pc := &passiveDataConn{
		listener: listener,
		tcpAddr:  tcpAddr,
	}

	go pc.listenWorker()

	return pc, nil
}

func (pc *passiveDataConn) Read(b []byte) (int, error) {
	if pc.conn == nil {
		return 0, fmt.Errorf("no connection")
	}

	return pc.conn.Read(b)
}

func (pc *passiveDataConn) Write(b []byte) (int, error) {
	if pc.conn == nil {
		return 0, fmt.Errorf("no connection")
	}

	return pc.conn.Write(b)
}

func (pc *passiveDataConn) Close() error {
	pc.conn.Close()
	pc.listener.Close()

	return nil
}

func (pc *passiveDataConn) Port() int {
	return pc.tcpAddr.Port
}

func (pc *passiveDataConn) listenWorker() {
	for {
		conn, err := pc.listener.Accept()
		if err != nil {
			fmt.Printf("unable to accept passive conn: %s\n", err)
			continue
		}

		pc.conn = conn

		break
	}
}
