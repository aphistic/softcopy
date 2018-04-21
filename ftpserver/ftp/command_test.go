package ftp

import (
	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type CommandSuite struct{}

func (s *CommandSuite) TestParseCommand(t sweet.T) {
	cmd, err := parseCommand([]byte("USER myusername\r\n"))
	Expect(err).To(BeNil())
	Expect(cmd).To(Equal(&Command{
		Command: "USER",
		Other:   "myusername",
	}))
}
