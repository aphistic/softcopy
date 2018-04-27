package ftp

import (
	"testing"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

func TestMain(m *testing.M) {
	RegisterFailHandler(sweet.GomegaFail)

	sweet.Run(m, func(s *sweet.S) {
		s.AddSuite(&CommandSuite{})
		s.AddSuite(&ResponseSuite{})

		s.AddSuite(&CwdCommandSuite{})
		s.AddSuite(&EprtCommandSuite{})
		s.AddSuite(&PassCommandSuite{})
		s.AddSuite(&RetrCommandSuite{})
		s.AddSuite(&SizeCommandSuite{})
		s.AddSuite(&StorCommandSuite{})
		s.AddSuite(&TypeCommandSuite{})
		s.AddSuite(&UserCommandSuite{})
	})
}
