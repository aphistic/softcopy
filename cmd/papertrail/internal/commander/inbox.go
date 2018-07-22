package commander

import (
	"context"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/dustin/go-humanize"
	"github.com/golang/protobuf/ptypes"
	"github.com/olekukonko/tablewriter"

	"github.com/aphistic/papertrail/internal/consts"
	"github.com/aphistic/papertrail/proto"
)

type cmdInbox struct {
	w      Writer
	client ptproto.PapertrailClient
}

func newCmdInbox(w Writer, client ptproto.PapertrailClient) *cmdInbox {
	return &cmdInbox{
		w:      w,
		client: client,
	}
}

func (c *cmdInbox) SubCommands() map[string]ParserCmd {
	return map[string]ParserCmd{}
}

func (c *cmdInbox) Description() string {
	return "Show documents in the Inbox"
}

func (c *cmdInbox) Suggestions(d prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func (c *cmdInbox) Execute(s string) error {
	res, err := c.client.FindFilesWithTags(context.Background(), &ptproto.FindFilesWithTagsRequest{
		TagNames: []string{
			consts.TagUnfiled,
		},
	})
	if err != nil {
		return err
	}

	t := tablewriter.NewWriter(c.w)
	t.SetBorder(false)
	t.SetHeader([]string{
		"ID",
		"Filename",
		"Date",
		"Size",
	})
	for _, file := range res.Files {
		docDate, err := ptypes.Timestamp(file.DocumentDate)
		if err != nil {
			continue
		}
		docDate = docDate.Local()

		t.Append([]string{
			file.Id,
			file.Filename,
			docDate.Format(time.RFC1123),
			humanize.Bytes(uint64(file.Size)),
		})
	}
	t.Render()

	return nil
}
