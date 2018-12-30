package runner

import (
	"github.com/alecthomas/kingpin"
)

type Config interface{}

type Runner interface {
	CommandName() string
	Setup(*kingpin.Application) Config
	Run(Config, Config) int
}
