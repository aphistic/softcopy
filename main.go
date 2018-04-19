package main

import (
	"os"

	"github.com/efritz/nacelle"
)

func main() {
	res := nacelle.NewBootstrapper(
		"papertrail",
		map[interface{}]interface{}{},
		func(runner *nacelle.ProcessRunner, container *nacelle.ServiceContainer) error {
			return nil
		},
	).Boot()

	if res != 0 {
		os.Exit(res)
	}
}
