package commander

import (
	"strings"

	"github.com/c-bata/go-prompt"
)

type ParserCmd interface {
	Description() string
	Execute(s string) error
	Suggestions(d prompt.Document) []prompt.Suggest
	SubCommands() map[string]ParserCmd
}

type parser struct {
	commands map[string]ParserCmd
}

func newParser(commands map[string]ParserCmd) *parser {
	return &parser{
		commands: commands,
	}
}

func (p *parser) RunExecution(s string) error {
	cmd, extra, err := p.findCommand(s)
	if err != nil {
		return err
	}

	return cmd.Execute(extra)
}

func (p *parser) CompileSuggestions(d prompt.Document) []prompt.Suggest {
	cmd, _, err := p.findCommand(d.TextBeforeCursor())
	if err == ErrNotFound {
		return []prompt.Suggest{}
	}

	return prompt.FilterHasPrefix(cmd.Suggestions(d), d.GetWordBeforeCursor(), true)
}

func (p *parser) Description() string {
	return "Parser - seeing this is a bug!"
}

func (p *parser) SubCommands() map[string]ParserCmd {
	return p.commands
}

func (p *parser) Suggestions(d prompt.Document) []prompt.Suggest {
	if strings.ToLower(d.TextBeforeCursor()) == "" {
		return []prompt.Suggest{}
	}

	res := []prompt.Suggest{}
	for k, v := range p.SubCommands() {
		res = append(res, prompt.Suggest{
			Text:        k,
			Description: v.Description(),
		})
	}

	return res
}

func (p *parser) Execute(s string) error {
	return nil
}

func (p *parser) findCommand(s string) (ParserCmd, string, error) {
	parts := strings.Split(s, " ")

	var curCmd ParserCmd = p
	for idx, part := range parts {
		sub, ok := curCmd.SubCommands()[part]
		if !ok {
			extras := []string{}
			for extraIdx := idx; idx < len(parts); idx++ {
				extras = append(extras, parts[extraIdx])
			}

			return curCmd, strings.Join(extras, " "), nil
		}
		curCmd = sub
	}

	return curCmd, "", nil
}
