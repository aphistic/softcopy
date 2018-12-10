package sqlite

import (
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3" // Import sqlite driver for our database
	migrate "github.com/rubenv/sql-migrate"

	"github.com/aphistic/softcopy/internal/pkg/storage"
)

const migrationPath = "../../internal/pkg/storage/sqlite/migrations"

type Client struct {
	dbPath string
	db     *sql.DB
}

var _ storage.Data = &Client{}

func NewClient(dbPath string) (*Client, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &Client{
		dbPath: dbPath,
		db:     db,
	}, nil
}

func (c *Client) Migrate() error {
	var migrations migrate.MigrationSource

	// If we're doing a migration first see if we have files in
	// the filesystem to load.
	fi, err := os.Stat(migrationPath)
	if os.IsNotExist(err) {
		// Directory doesn't exist, continue on
	} else if err != nil {
		return err
	} else if !fi.IsDir() {
		// The path isn't a directory, continue on
	} else {
		// The directory exists, use it for migrations
		migrations = &migrate.FileMigrationSource{
			Dir: migrationPath,
		}
	}

	_, err = migrate.Exec(c.db, "sqlite3", migrations, migrate.Up)
	if err != nil {
		return err
	}

	return nil
}
