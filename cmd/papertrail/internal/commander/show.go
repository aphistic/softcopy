package commander

import (
	"context"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"

	"github.com/aphistic/papertrail/proto"
	"github.com/dustin/go-humanize"
	"github.com/golang/protobuf/ptypes"
)

type cmdShow struct {
	w      Writer
	client ptproto.PapertrailClient
}

func newCmdShow(w Writer, client ptproto.PapertrailClient) *cmdShow {
	return &cmdShow{
		w:      w,
		client: client,
	}
}

func (c *cmdShow) SubCommands() map[string]ParserCmd {
	return map[string]ParserCmd{}
}

func (c *cmdShow) Description() string {
	return "Show details of a document"
}

func (c *cmdShow) Suggestions(d prompt.Document) []prompt.Suggest {
	idPrefix := strings.TrimSpace(d.GetWordBeforeCursor())

	if len(idPrefix) < 1 {
		// Don't show suggestions if less than one character is entered
		return []prompt.Suggest{}
	}

	res, err := c.client.FindFilesWithIdPrefix(context.Background(), &ptproto.FindFilesWithIdPrefixRequest{
		IdPrefix: idPrefix,
	})
	if err != nil {
		c.w.Printf("err: %s\n", err)
		return []prompt.Suggest{}
	}

	suggestions := []prompt.Suggest{}
	for _, f := range res.Files {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        f.Id,
			Description: f.Filename,
		})
	}

	return suggestions
}

func (c *cmdShow) Execute(s string) error {
	id := strings.TrimSpace(s)
	file, err := c.client.GetFile(context.Background(), &ptproto.GetFileRequest{
		Id: id,
	})
	if err != nil {
		return err
	}

	docDate, err := ptypes.Timestamp(file.File.File.DocumentDate)
	if err != nil {
		return err
	}

	c.w.Printf("\n")
	c.w.Printf("ID: %s\n", file.File.File.Id)
	c.w.Printf("Name: %s\n", file.File.File.Filename)
	c.w.Printf(
		"Date: %s (%s)\n",
		humanize.Time(docDate), docDate.Format(time.RFC850),
	)
	c.w.Printf("Size: %s\n", humanize.Bytes(uint64(file.File.File.Size)))

	return nil
}
