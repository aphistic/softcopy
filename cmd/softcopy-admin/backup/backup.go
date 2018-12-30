package backup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kingpin"
	"google.golang.org/grpc"

	"github.com/aphistic/softcopy/cmd/softcopy-admin/config"
	"github.com/aphistic/softcopy/cmd/softcopy-admin/runner"
	"github.com/aphistic/softcopy/internal/consts"
	scproto "github.com/aphistic/softcopy/proto"
	"github.com/aphistic/softcopy/storage/backup"
)

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

func (r *Runner) CommandName() string {
	return "backup"
}

func (r *Runner) Setup(app *kingpin.Application) runner.Config {
	cfg := NewConfig()

	cmd := app.Command(
		CommandName,
		fmt.Sprintf("Back up %s database and documents", consts.ProcessName),
	)
	cmd.Flag("out", "Output directory for backup").
		Default("./softcopy-backup/").StringVar(&cfg.Out)

	return cfg
}

func (r *Runner) Run(cfg runner.Config, runCfg runner.Config) int {
	genCfg := cfg.(*config.Config)
	backupCfg := runCfg.(*Config)

	fmt.Printf("gen: %#v\n", genCfg)
	fmt.Printf("back: %#v\n", backupCfg)

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", genCfg.Host, genCfg.Port),
		grpc.WithInsecure(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error dialing server: %s\n", err)
		return 1
	}

	b, err := backup.NewBackup(backupCfg.Out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not create backup: %s\n", err)
		return 1
	}

	client := scproto.NewSoftcopyAdminClient(conn)

	ctx := context.Background()
	allFiles, err := client.AllFiles(ctx, &scproto.AllFileRequest{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting all files: %s\n", err)
		return 1
	}

	for {
		file, err := allFiles.Recv()
		if err == io.EOF {
			return 0
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting file: %s\n", err)
			return 1
		}

		fmt.Printf("Got file: %#v\n", file.File.Filename)
		err = b.WriteFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write backup file info: %s\n", err)
			continue
		}

		data, err := client.DownloadFile(ctx, &scproto.DownloadFileRequest{
			Id: file.File.Id,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not download file data: %s\n", err)
			continue
		}

		buf := bytes.NewBuffer(nil)
		for {
			fileData, err := data.Recv()
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting file part: %s\n", err)
				break
			}

			curRead := 0
			for {
				n, err := buf.Write(fileData.Data[curRead:])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing file part: %s\n", err)
					continue
				}
				curRead += n

				if curRead >= len(fileData.Data) {
					break
				}
			}
		}

		err = b.WriteData(file.File.Id, buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write backup file: %s\n", err)
			continue
		}
	}
}
