package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/c-bata/go-prompt"
	"google.golang.org/grpc"

	"github.com/aphistic/softcopy/cmd/softcopy/internal/commander"
	"github.com/aphistic/softcopy/internal/consts"
	"github.com/aphistic/softcopy/proto"
)

var (
	host string
	port int
)

func main() {
	app := kingpin.New(consts.ProcessName, "")
	app.Flag("host", fmt.Sprintf("Hostname of %s server", consts.ProcessName)).Short('h').
		Default("localhost").StringVar(&host)
	app.Flag("port", fmt.Sprintf("Port of %s server", consts.ProcessName)).Short('p').
		Default("6000").IntVar(&port)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing arguments: %s\n", err)
		os.Exit(1)
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to server: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := scproto.NewSoftcopyClient(conn)

	client.FindFilesWithTags(context.Background(), &scproto.FindFilesWithTagsRequest{})

	cmdr := commander.NewCommander(client)
	cmdr.Startup()
	p := prompt.New(
		cmdr.Executor,
		cmdr.Completer,
		prompt.OptionTitle("softcopy"),
		prompt.OptionPrefixTextColor(prompt.Cyan),
		prompt.OptionLivePrefix(cmdr.LivePrefix),
	)

	p.Run()
}
