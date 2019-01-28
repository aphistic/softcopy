package main

import (
	"os"

	"github.com/efritz/zubrin"

	"github.com/efritz/nacelle"

	"github.com/aphistic/softcopy/internal/app/softcopy-server/importer"
	"github.com/aphistic/softcopy/internal/pkg/api"
	"github.com/aphistic/softcopy/internal/pkg/apiserver"
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

			runner.RegisterProcess(
				apiserver.NewProcess(),
				nacelle.WithProcessName("apiserver"),
			)
			runner.RegisterProcess(
				importer.NewProcess(),
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
