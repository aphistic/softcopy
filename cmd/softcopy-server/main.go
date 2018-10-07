package main

import (
	"os"

	"github.com/efritz/nacelle"
	"github.com/efritz/nacelle/process"

	"github.com/aphistic/softcopy/api"
	"github.com/aphistic/softcopy/apiserver"
	"github.com/aphistic/softcopy/ftpserver"
)

func main() {
	res := nacelle.NewBootstrapper(
		"softcopy",
		map[interface{}]interface{}{
			process.GRPCConfigToken: &process.GRPCConfig{},
			api.ConfigToken:         &api.Config{},
		},
		func(runner *nacelle.ProcessRunner, container *nacelle.ServiceContainer) error {
			runner.RegisterInitializer(
				api.NewInitializer(),
				nacelle.WithInitializerName("api"),
			)
			runner.RegisterProcess(
				ftpserver.NewProcess(),
				nacelle.WithProcessName("ftpserver"),
			)
			runner.RegisterProcess(
				apiserver.NewProcess(),
				nacelle.WithProcessName("apiserver"),
			)

			return nil
		},
	).Boot()

	if res != 0 {
		os.Exit(res)
	}
}
