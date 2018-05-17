package main

import (
	"os"

	"github.com/efritz/nacelle"

	"github.com/aphistic/papertrail/api"
	"github.com/aphistic/papertrail/ftpserver"
)

func main() {
	res := nacelle.NewBootstrapper(
		"papertrail",
		map[interface{}]interface{}{
			api.ConfigToken: &api.Config{},
		},
		func(runner *nacelle.ProcessRunner, container *nacelle.ServiceContainer) error {
			runner.RegisterInitializer(
				api.NewInitializer(),
				nacelle.WithInitializerName("api"),
			)
			runner.RegisterProcess(
				ftpserver.NewProcess(),
				nacelle.WithProcessName("ftp"),
			)

			return nil
		},
	).Boot()

	if res != 0 {
		os.Exit(res)
	}
}
