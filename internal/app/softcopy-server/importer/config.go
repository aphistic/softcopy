package importer

import (
	"github.com/aphistic/softcopy/internal/pkg/config"
)

type importersConfig struct {
	Importers []*importerConfig `file:"importers"`
}

type importerConfig struct {
	Type    string           `yaml:"type"`
	Options []*config.Option `yaml:"options"`
}
