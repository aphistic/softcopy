package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/alecthomas/kingpin"

	scfs "github.com/aphistic/softcopy/internal/app/softcopy-fuse/fs"
	"github.com/aphistic/softcopy/internal/pkg/logging"
)

var (
	mountPath    string
	serverHost   string
	serverPort   int
	fuseDebugLog bool
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
	app.Flag("fuse-debug", "Enable fuse debug logging").
		Default("false").
		BoolVar(&fuseDebugLog)

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

	logger.Info("mounting fuse filesystem at %s", mountPath)
	conn, err := fuse.Mount(mountPath)
	if err != nil {
		logger.Error("Could not mount: %s", err)
		return
	}
	defer func() {
		logger.Info("umounting fuse filesystem")
		err := fuse.Unmount(mountPath)
		if err != nil {
			logger.Error("could not unmount: %s\n", err)
		}
	}()
	defer conn.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan)

	go func() {
		logger.Info("starting softcopy connection to %s:%d", serverHost, serverPort)
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
	}()

mainLoop:
	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGURG:
			case os.Interrupt:
				logger.Info("got interrupt")
				break mainLoop
			default:
				logger.Info("got sig: %s", sig)
			}
		}
	}
}
