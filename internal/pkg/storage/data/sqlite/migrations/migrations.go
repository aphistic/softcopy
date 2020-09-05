package migrations

import (
	"bytes"
	"sort"

	"github.com/aphistic/goblin"
	migrate "github.com/rubenv/sql-migrate"
)

//go:generate /home/aphistic/dev/goblin/cmd/goblin/goblin --name migrations -i *.sql

type VaultMigrationSource struct {
	vault goblin.Vault
}

var _ migrate.MigrationSource = &VaultMigrationSource{}

func NewVaultMigrationSource() (*VaultMigrationSource, error) {
	vault, err := loadVaultMigrations()
	if err != nil {
		return nil, err
	}

	return &VaultMigrationSource{
		vault: vault,
	}, nil
}

func (vms *VaultMigrationSource) FindMigrations() ([]*migrate.Migration, error) {
	var migrations []*migrate.Migration

	var sortedFiles []string

	files, err := vms.vault.ReadDir(".")
	if err != nil {
		return nil, err
	}

	for _, fInfo := range files {
		if fInfo.IsDir() {
			continue
		}
		sortedFiles = append(sortedFiles, fInfo.Name())
	}
	sort.Strings(sortedFiles)

	for _, filename := range sortedFiles {
		file, _ := vms.vault.ReadFile(filename)
		migration, err := migrate.ParseMigration(filename, bytes.NewReader(file))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}
