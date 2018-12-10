package main

import (
	"os"

	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/internal/pkg/api"
	"github.com/aphistic/softcopy/internal/pkg/apiserver"
	"github.com/aphistic/softcopy/internal/pkg/ftpserver"
)

func main() {
	res := nacelle.NewBootstrapper(
		"softcopy",
		func(runner nacelle.ProcessContainer, container nacelle.ServiceContainer) error {
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
