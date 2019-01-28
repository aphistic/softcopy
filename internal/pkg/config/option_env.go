package config

import (
	"os"
)

type EnvLoader interface {
	LookupEnv(string) (string, bool)
}

type realEnvLoader struct{}

func (rel *realEnvLoader) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}
