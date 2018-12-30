package backup

const CommandName = "backup"

type Config struct {
	Out string
}

func NewConfig() *Config {
	return &Config{}
}
