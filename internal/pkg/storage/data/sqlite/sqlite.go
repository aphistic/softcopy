package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3" // import sqlite driver
	"github.com/rubenv/sql-migrate"

	"github.com/aphistic/softcopy/internal/pkg/storage"
	"github.com/aphistic/softcopy/internal/pkg/storage/data/sqlite/migrations"
)

type Client struct {
	dbPath string
	db     *sql.DB
}

var _ storage.Data = &Client{}

func NewClient(dbPath string) (*Client, error) {
	dbRoot := path.Dir(dbPath)
	_, err := os.Stat(dbRoot)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dbRoot, 0755)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	fi, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		// This is fine, it'll be created when the DB is opened.
	} else if err != nil {
		return nil, err
	} else if fi.IsDir() {
		return nil, fmt.Errorf("database path is a directory, not a file")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Make sure foreign key constraints are checked
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return nil, err
	}

	return &Client{
		dbPath: dbPath,
		db:     db,
	}, nil
}

func (c *Client) Migrate() error {
	ms, err := migrations.NewVaultMigrationSource()
	if err != nil {
		return err
	}

	_, err = migrate.Exec(c.db, "sqlite3", ms, migrate.Up)
	if err != nil {
		return err
	}

	return nil
}
