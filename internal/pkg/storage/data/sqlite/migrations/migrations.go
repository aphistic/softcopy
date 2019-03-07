package migrations

import (
	"bytes"
	"sort"

	"github.com/aphistic/goblin"
	"github.com/rubenv/sql-migrate"
)

//go:generate goblin --name migrations --include ./*.sql

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
	for filename := range vms.vault.Files() {
		sortedFiles = append(sortedFiles, filename)
	}
	sort.Strings(sortedFiles)

	for _, filename := range sortedFiles {
		file, _ := vms.vault.File(filename)
		migration, err := migrate.ParseMigration(filename, bytes.NewReader(file))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}
