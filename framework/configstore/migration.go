package configstore

import (
	"fmt"
	"os/exec"

	"github.com/maximhq/bifrost/core/schemas" // adjust import path for your logger interface
)

// Migration migrates a SQLite database to Postgres using pgloader.
// sqlitePath: path to the SQLite .db file
// postgresLink: connection string for Postgres (e.g. "pgsql://user:pass@host/dbname")
func MigrateFromSql(sqlitePath, postgresLink string, logger schemas.Logger) error {
	logger.Info("Starting migration: SQLite → Postgres", sqlitePath, postgresLink)

	cmd := exec.Command("pgloader",
		fmt.Sprintf("sqlite:///%s", sqlitePath),
		postgresLink,
	)

	// Show pgloader output in console
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logger.Error("Migration failed", "error", err.Error())
		return fmt.Errorf("pgloader failed: %w", err)
	}

	logger.Info("Migration completed successfully: SQLite → Postgres")
	return nil
}

// Migration from Postgres back to SQLite
func MigrateFromPostgres(sqlitePath, postgresLink string, logger schemas.Logger) error {
	logger.Info("Starting migration: Postgres → SQLite")

	cmd := exec.Command("pgloader",
		postgresLink,
		fmt.Sprintf("sqlite:///%s", sqlitePath),
	)

	if err := cmd.Run(); err != nil {
		logger.Error("Migration failed", "error", err.Error())
		return fmt.Errorf("pgloader failed: %w", err)
	}

	logger.Info("Migration completed successfully: Postgres → SQLite")
	return nil
}
