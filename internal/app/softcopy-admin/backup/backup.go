package backup

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kingpin"
	"google.golang.org/grpc"

	"github.com/aphistic/softcopy/internal/app/softcopy-admin/config"
	"github.com/aphistic/softcopy/internal/app/softcopy-admin/runner"
	"github.com/aphistic/softcopy/internal/pkg/consts"
	"github.com/aphistic/softcopy/internal/pkg/storage/backup"
	"github.com/aphistic/softcopy/pkg/proto"
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

	client := scproto.NewSoftcopyClient(conn)
	adminClient := scproto.NewSoftcopyAdminClient(conn)

	ctx := context.Background()
	allFiles, err := adminClient.AllFiles(ctx, &scproto.AllFileRequest{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting all files: %s\n", err)
		return 1
	}

	// TODO This impl is lazy on error handling, make it less so!
	hasher := sha256.New()
	fileBuf := bytes.NewBuffer(nil)
fileLoop:
	for {
		hasher.Reset()
		fileBuf.Reset()

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

		ofRes, err := client.OpenFile(ctx, &scproto.OpenFileRequest{
			Id:   file.GetFile().GetId(),
			Mode: scproto.FileMode_READ,
		})
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"Could not open file '%s' for download: %s\n",
				file.GetFile().GetId(), err,
			)
			continue
		}

		readOffset := uint64(0)
		for {
			readRes, err := client.ReadFile(ctx, &scproto.ReadFileRequest{
				HandleId: ofRes.GetHandleId(),
				Offset:   readOffset,
				Size:     4096,
			})
			if err != nil {
				client.CloseFile(ctx, &scproto.CloseFileRequest{
					HandleId: ofRes.GetHandleId(),
				})
				continue fileLoop
			}

			curRead := 0
			for {
				n, err := fileBuf.Write(readRes.GetData()[curRead:])
				if err == io.EOF {
					break
				} else if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing file part: %s\n", err)
					break
				}

				hashN, err := hasher.Write(readRes.GetData()[curRead:n])
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error writing hash part: %s\n", err)
					break
				}

				if n != hashN {
					fmt.Fprintf(os.Stderr, "Hash and file write were not equal\n")
					break
				}

				curRead += n

				if curRead >= len(readRes.GetData()) {
					break
				}
			}

			if uint64(fileBuf.Len()) >= file.GetFile().GetSize() {
				break
			}

			readOffset = uint64(fileBuf.Len())
		}

		err = b.WriteData(file.GetFile().GetId(), fileBuf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not write backup file: %s\n", err)
			continue
		}

		_, err = client.CloseFile(ctx, &scproto.CloseFileRequest{
			HandleId: ofRes.GetHandleId(),
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not close remote file: %s\n", err)
			continue
		}

		/*
			fmt.Printf(
				"local: %s\nclose: %s\nremote: %s\n",
				fmt.Sprintf("%x", hasher.Sum(nil)),
				closeRes.GetHash(),
				file.GetFile().GetHash(),
			)
		*/
		if fmt.Sprintf("%x", hasher.Sum(nil)) != file.GetFile().GetHash() {
			fmt.Fprintf(os.Stderr, "Remote and local file hashes did not match\n")
			continue
		}
	}
}
