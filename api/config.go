package api

type Config struct {
	StorageRoot string `env:"STORAGE_ROOT" default:"./data"`
}

type configToken struct{}

var ConfigToken = &configToken{}
