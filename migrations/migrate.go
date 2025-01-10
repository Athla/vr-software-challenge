package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
)

//go:embed *.sql
var migrationFiles embed.FS

type Migration struct {
	Version  int
	Filename string
	Content  string
}

func LoadMigrations() ([]Migration, error) {
	entries, err := migrationFiles.ReadDir(".")
	if err != nil {
		log.Errorf("Unable to read migration dir due: %v", err)
		return nil, err
	}

	var migrations []Migration

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			content, err := migrationFiles.ReadFile(entry.Name())
			if err != nil {
				log.Errorf("Unable to open migration file '%v' due: %v", entry.Name(), err)
				return nil, err
			}

			var version int
			_, err = fmt.Sscanf(entry.Name(), "%d_", &version)
			if err != nil {
				log.Errorf("Unable to parse version from file '%s' due: %v", entry.Name(), err)
				return nil, err
			}

			migrations = append(migrations, Migration{
				Version:  version,
				Filename: entry.Name(),
				Content:  string(content),
			})
		}
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func Migrate(db *sql.DB, ctx context.Context) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		log.Errorf("Unable to check or create migration table due: %v", err)
		return err
	}

	migrations, err := LoadMigrations()
	if err != nil {
		log.Errorf("Unable to load migrations due: %v", err)
		return err
	}

	for _, migration := range migrations {
		var applied bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", migration.Version).Scan(&applied)
		if err != nil {
			log.Errorf("Unable to query 'schema_migrations' due: %v", err)
			return err
		}

		if !applied {
			log.Infof("Applying migration: %s", migration.Filename)
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				log.Errorf("Unable to start transaction due: %v", err)
				return err
			}

			if _, err := tx.Exec(migration.Content); err != nil {
				tx.Rollback()
				log.Errorf("Unable to execute transaction due: %v", err)
				return err
			}

			if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
				tx.Rollback()
				log.Errorf("Unable to insert new version of current migration into database due: %v", err)
				return err
			}

			if err := tx.Commit(); err != nil {
				log.Errorf("Unable to transaction commit due: %v", err)
				return err
			}
			log.Infof("Migration %s applied successfully!", migration.Filename)
		}
	}
	log.Info("All migrations have been finished.")

	return nil
}
