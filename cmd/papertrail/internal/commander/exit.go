package commander

import (
	"os"

	"github.com/c-bata/go-prompt"
)

type cmdExit struct{}

func newCmdExit() *cmdExit {
	return &cmdExit{}
}

func (c *cmdExit) SubCommands() map[string]ParserCmd {
	return map[string]ParserCmd{}
}

func (c *cmdExit) Description() string {
	return "Exit"
}

func (c *cmdExit) Suggestions(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func (c *cmdExit) Execute(s string) error {
	os.Exit(0)

	return nil
}
