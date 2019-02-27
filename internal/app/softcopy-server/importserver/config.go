package importserver

import (
	"github.com/aphistic/softcopy/internal/pkg/config"
)

type importersConfig struct {
	Importers []*importerConfig `file:"importers"`
}

type importerConfig struct {
	Name    string           `yaml:"name"`
	Type    string           `yaml:"type"`
	Options []*config.Option `yaml:"options"`
}
