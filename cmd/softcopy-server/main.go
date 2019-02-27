package main

import (
	"github.com/aphistic/softcopy/internal/app/softcopy-server/uiserver"
	"os"

	"github.com/efritz/nacelle"
	"github.com/efritz/zubrin"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/apiserver"
	"github.com/aphistic/softcopy/internal/app/softcopy-server/importserver"
	"github.com/aphistic/softcopy/internal/pkg/api"
)

func main() {
	sourcer := zubrin.NewMultiSourcer(
		zubrin.NewOptionalFileSourcer("/etc/softcopy/config.yml", zubrin.ParseYAML),
		zubrin.NewOptionalFileSourcer("/etc/softcopy/config.yaml", zubrin.ParseYAML),
		zubrin.NewOptionalFileSourcer("config.yaml", zubrin.ParseYAML),
		zubrin.NewOptionalFileSourcer("config.yml", zubrin.ParseYAML),
		zubrin.NewEnvSourcer(""),
	)

	res := nacelle.NewBootstrapper(
		"softcopy",
		func(runner nacelle.ProcessContainer, container nacelle.ServiceContainer) error {
			runner.RegisterInitializer(
				api.NewInitializer(),
				nacelle.WithInitializerName("api"),
			)
			runner.RegisterInitializer(
				importserver.NewInitializer(),
				nacelle.WithInitializerName("importers"),
			)

			runner.RegisterProcess(
				apiserver.NewProcess(),
				nacelle.WithProcessName("apiserver"),
			)
			runner.RegisterProcess(
				uiserver.NewProcess(),
				nacelle.WithProcessName("uiserver"),
			)
			runner.RegisterProcess(
				importserver.NewProcess(),
				nacelle.WithProcessName("importers"),
			)

			return nil
		},
		nacelle.WithConfigSourcer(sourcer),
	).Boot()

	if res != 0 {
		os.Exit(res)
	}
}
