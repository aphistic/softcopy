package commander

import (
	"github.com/c-bata/go-prompt"

	"github.com/aphistic/softcopy/internal/consts"
	"github.com/aphistic/softcopy/proto"
)

type Commander struct {
	client scproto.SoftcopyClient
	writer Writer
	parser *parser
}

func NewCommander(client scproto.SoftcopyClient) *Commander {
	w := &consoleWriter{}

	return &Commander{
		client: client,
		writer: w,
		parser: newParser(map[string]ParserCmd{
			"inbox": newCmdInbox(w, client),
			"show":  newCmdShow(w, client),
			"exit":  newCmdExit(),
		}),
	}
}

func (c *Commander) Startup() error {
	c.writer.Printf("softcopy v%s\n", consts.ProcessVersion)
	c.writer.Printf("Please use `exit` or `Ctrl-D` to exit.\n")
	c.writer.Printf("\n")
	return nil
}

func (c *Commander) Executor(s string) {
	err := c.parser.RunExecution(s)
	if err != nil {
		c.writer.Printf("error: %s\n", err)
	}
}

func (c *Commander) Completer(d prompt.Document) []prompt.Suggest {
	return c.parser.CompileSuggestions(d)
}

func (c *Commander) LivePrefix() (string, bool) {
	return "> ", true
}
