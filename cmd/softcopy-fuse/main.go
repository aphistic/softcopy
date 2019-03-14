package main

import (
	"fmt"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/alecthomas/kingpin"

	scfs "github.com/aphistic/softcopy/internal/app/softcopy-fuse/fs"
	"github.com/aphistic/softcopy/internal/pkg/logging"
)

var (
	mountPath string
	serverHost string
	serverPort int
)

func main() {
	app := kingpin.New("softcopy", "Softcopy")
	app.Flag("mount", "Mount path").
		Short('m').
		Default("scmount").
		StringVar(&mountPath)
	app.Flag("host", "Server host").
		Short('h').
		Default("localhost").
		StringVar(&serverHost)
	app.Flag("port", "Server port").
		Short('p').
		Default("6000").
		IntVar(&serverPort)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command line: %s\n", err)
		return
	}

	logger, err := logging.NewGomolLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize logging: %s\n", err)
		return
	}
	defer logger.Shutdown()

	conn, err := fuse.Mount(mountPath)
	if err != nil {
		logger.Error("Could not mount: %s", err)
		return
	}
	defer func() {
		err := fuse.Unmount(mountPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not unmount: %s\n", err)
		}
	}()
	defer conn.Close()

	connectedFS, err := scfs.NewFileSystem(
		serverHost, serverPort,
		scfs.WithLogger(logger),
	)
	if err != nil {
		logger.Error("Error connecting to server: %s\n", err)
		return
	}
	err = fs.Serve(conn, connectedFS)
	if err != nil {
		logger.Error("Error in serve: %s\n", err)
	}
}
