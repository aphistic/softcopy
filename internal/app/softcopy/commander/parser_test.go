package commander

import (
	"github.com/aphistic/sweet"
	"github.com/c-bata/go-prompt"
	. "github.com/onsi/gomega"
)

type testCmd struct {
	commands    map[string]ParserCmd
	description string
}

func (tc *testCmd) SubCommands() map[string]ParserCmd {
	return tc.commands
}

func (tc *testCmd) Description() string {
	return tc.description
}

func (tc *testCmd) Suggestions(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func (tc *testCmd) Execute(s string) error {
	return nil
}

type ParserSuite struct{}

func (s *ParserSuite) TestFindCommand(t sweet.T) {
	p := newParser(map[string]ParserCmd{
		"show": &testCmd{
			commands: map[string]ParserCmd{},
		},
	})

	_, extra, err := p.findCommand("show 1234")
	Expect(err).To(BeNil())
	Expect(extra).To(Equal("1234"))
}
