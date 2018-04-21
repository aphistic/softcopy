package main

import (
	"os"

	"github.com/efritz/nacelle"

	"github.com/aphistic/papertrail/ftpserver"
)

func main() {
	res := nacelle.NewBootstrapper(
		"papertrail",
		map[interface{}]interface{}{},
		func(runner *nacelle.ProcessRunner, container *nacelle.ServiceContainer) error {
			runner.RegisterProcess(ftpserver.NewProcess())
			return nil
		},
	).Boot()

	if res != 0 {
		os.Exit(res)
	}
}
