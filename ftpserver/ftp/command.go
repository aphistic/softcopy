package ftp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Command interface {
	Command() string
}

type BasicCommand struct {
	cmd   string
	other []byte
}

func (c *BasicCommand) Command() string {
	return c.cmd
}

func parseCommand(data []byte) (Command, error) {
	cmd := &BasicCommand{}

	inCommand := true
	otherStart := 0
	for idx := 0; idx < len(data); idx++ {
		if inCommand {
			if data[idx] == ' ' {
				cmd.cmd = strings.ToUpper(string(data[0:idx]))
				inCommand = false
				otherStart = idx + 1
				continue
			} else if data[idx] == '\r' && len(data) > idx+1 && data[idx+1] == '\n' {
				cmd.cmd = strings.ToUpper(string(data[0:idx]))
				return chooseCommand(cmd)
			}
		} else {
			if data[idx] == '\r' && len(data) > idx+1 && data[idx+1] == '\n' {
				cmd.other = data[otherStart:idx]
				return chooseCommand(cmd)
			}
		}
	}

	return nil, fmt.Errorf("reached end")
}

func chooseCommand(cmd *BasicCommand) (Command, error) {
	switch cmd.Command() {
	case "EPRT":
		c := &EprtCommand{}
		err := c.UnmarshalText(cmd.other)
		if err != nil {
			return nil, err
		}
		return c, nil
	case "PASS":
		c := &PassCommand{}
		err := c.UnmarshalText(cmd.other)
		if err != nil {
			return nil, err
		}
		return c, nil
	case "USER":
		c := &UserCommand{}
		err := c.UnmarshalText(cmd.other)
		if err != nil {
			return nil, err
		}
		return c, nil
	}

	return cmd, nil
}

type EprtCommand struct {
	Version int
	Address net.IP
	Port    int
}

func (c *EprtCommand) Command() string {
	return "EPRT"
}

func (c *EprtCommand) UnmarshalText(data []byte) error {
	part := -1

	afStr := []byte{}
	ipStr := []byte{}
	portStr := []byte{}

parseLoop:
	for idx := 0; idx < len(data); idx++ {
		if data[idx] == '|' {
			part++
			continue
		}

		switch part {
		case 0:
			afStr = append(afStr, data[idx])
		case 1:
			ipStr = append(ipStr, data[idx])
		case 2:
			portStr = append(portStr, data[idx])
		default:
			break parseLoop
		}
	}

	ip := net.ParseIP(string(ipStr))
	if ip == nil {
		return fmt.Errorf("could not parse ip in eprt")
	}

	port, err := strconv.ParseInt(string(portStr), 10, 0)
	if err != nil {
		return err
	}

	switch string(afStr) {
	case "1":
		c.Version = 4
	case "2":
		c.Version = 6
	default:
		return fmt.Errorf("unknown protocol version")
	}

	c.Address = ip
	c.Port = int(port)

	return nil
}

type PassCommand struct {
	Password string
}

func (c *PassCommand) UnmarshalText(data []byte) error {
	c.Password = string(data)
	return nil
}
func (c *PassCommand) Command() string {
	return "PASS"
}

type UserCommand struct {
	User string
}

func (c *UserCommand) UnmarshalText(data []byte) error {
	c.User = string(data)
	return nil
}
func (c *UserCommand) Command() string {
	return "USER"
}
