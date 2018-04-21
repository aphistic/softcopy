package ftp

import (
	"fmt"
	"net"
)

type ftpConn struct {
	conn net.Conn

	commands chan *Command
}

func newFtpConn(conn net.Conn) *ftpConn {
	fc := &ftpConn{
		conn:     conn,
		commands: make(chan *Command),
	}

	go fc.worker()

	return fc
}

func (c *ftpConn) Close() error {
	return c.conn.Close()
}

func (c *ftpConn) Command() chan *Command {
	return c.commands
}

func (c *ftpConn) Respond(r *Response) error {
	data, err := r.MarshalText()
	if err != nil {
		return err
	}
	data = append(data, []byte{'\r', '\n'}...)

	n, err := c.conn.Write(data)
	if err != nil {
		return err
	} else if n != len(data) {
		fmt.Printf("short write on ftp response\n")
		return fmt.Errorf("short write")
	}

	return nil
}

func (c *ftpConn) worker() {
	buf := make([]byte, 4096)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			fmt.Printf("error reading ftp conn: %s\n", err)
			continue
		}
		fmt.Printf("read from ftp conn: %s\n", buf[0:n])
		cmd, err := parseCommand(buf[0:n])
		if err != nil {
			fmt.Printf("error reading command: %s\n", err)
			continue
		}
		fmt.Printf("Read command: %#v\n", cmd)
	}

}
