package ftp

import (
	"net"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type CommandSuite struct{}

func (s *CommandSuite) TestParseCommand(t sweet.T) {
	cmd, err := parseCommand([]byte("USER myusername\r\n"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(&UserCommand{
		User: "myusername",
	}))
}

func (s *CommandSuite) TestParseCommandShort(t sweet.T) {
	cmd, err := parseCommand([]byte("SYST\r\n"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(&BasicCommand{
		cmd:   "SYST",
		other: nil,
	}))
}

func (s *CommandSuite) TestParseCommandEprt(t sweet.T) {
	cmd, err := parseCommand([]byte("EPRT |2|::1|56587|\r\n"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(&EprtCommand{
		Version: 6,
		Address: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		Port:    56587,
	}))
}

type PassCommandSuite struct{}

func (s *PassCommandSuite) TestUnmarshalText(t sweet.T) {
	var cmd PassCommand
	err := cmd.UnmarshalText([]byte("mypassword"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(PassCommand{
		Password: "mypassword",
	}))
}

type EprtCommandSuite struct{}

func (s *EprtCommandSuite) TestUnmarshalTextIPv6(t sweet.T) {
	var cmd EprtCommand
	err := cmd.UnmarshalText([]byte("|2|::1|56587|"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(EprtCommand{
		Version: 6,
		Address: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		Port:    56587,
	}))
}

type UserCommandSuite struct{}

func (s *UserCommandSuite) TestUnmarshalText(t sweet.T) {
	var cmd UserCommand
	err := cmd.UnmarshalText([]byte("myusername"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(UserCommand{
		User: "myusername",
	}))
}
