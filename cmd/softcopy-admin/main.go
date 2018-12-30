package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"

	"github.com/aphistic/softcopy/cmd/softcopy-admin/backup"
	"github.com/aphistic/softcopy/cmd/softcopy-admin/config"
	"github.com/aphistic/softcopy/cmd/softcopy-admin/runner"
	"github.com/aphistic/softcopy/internal/consts"
)

func main() {
	cmdRunners := []runner.Runner{
		backup.NewRunner(),
	}

	cfg := config.NewConfig()
	app := kingpin.New(
		fmt.Sprintf("%s-admin", consts.ProcessName),
		fmt.Sprintf("Administrative functions for %s", consts.ProcessName),
	)
	app.Flag("host", fmt.Sprintf("Host for %s server", consts.ProcessName)).
		Default("localhost").StringVar(&cfg.Host)
	app.Flag("port", fmt.Sprintf("Port for %s server", consts.ProcessName)).
		Default("6000").IntVar(&cfg.Port)

	runners := map[string]runner.Runner{}
	configs := map[string]runner.Config{}
	for _, runner := range cmdRunners {
		runners[runner.CommandName()] = runner
		configs[runner.CommandName()] = runner.Setup(app)
	}

	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing command line args: %s", err)
		os.Exit(1)
	}

	runner, ok := runners[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "Couldn't find command '%s'", cmd)
		os.Exit(1)
	}
	config, ok := configs[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "Couldn't find command '%s'", cmd)
		os.Exit(1)
	}

	ret := runner.Run(cfg, config)
	if ret != 0 {
		os.Exit(ret)
	}
}
