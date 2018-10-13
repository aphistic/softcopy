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

type CwdCommandSuite struct{}

func (s *CwdCommandSuite) TestUnmarshalText(t sweet.T) {
	var cmd CwdCommand
	err := cmd.UnmarshalText([]byte("mydirectory"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(CwdCommand{
		Path: "mydirectory",
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

type RetrCommandSuite struct{}

func (s *RetrCommandSuite) TestUnmarshalText(t sweet.T) {
	var cmd RetrCommand
	err := cmd.UnmarshalText([]byte("this_is_a_file_1234.txt"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(RetrCommand{
		Path: "this_is_a_file_1234.txt",
	}))
}

type SizeCommandSuite struct{}

func (s *SizeCommandSuite) TestUnmarshalText(t sweet.T) {
	var cmd SizeCommand
	err := cmd.UnmarshalText([]byte("this_is_a_file_1234.txt"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(SizeCommand{
		Path: "this_is_a_file_1234.txt",
	}))
}

type StorCommandSuite struct{}

func (s *StorCommandSuite) TestUnmarshalText(t sweet.T) {
	var cmd StorCommand
	err := cmd.UnmarshalText([]byte("this_is_a_file_1234.txt"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(StorCommand{
		Path: "this_is_a_file_1234.txt",
	}))
}

type TypeCommandSuite struct{}

func (s *TypeCommandSuite) TestUnmarshalText(t sweet.T) {
	var cmd TypeCommand
	err := cmd.UnmarshalText([]byte("A"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(TypeCommand{
		Type: TypeASCII,
	}))

	err = cmd.UnmarshalText([]byte("E"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(TypeCommand{
		Type: TypeEBCDIC,
	}))

	err = cmd.UnmarshalText([]byte("I"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(TypeCommand{
		Type: TypeImage,
	}))

	err = cmd.UnmarshalText([]byte("L"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(TypeCommand{
		Type: TypeLocal,
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
