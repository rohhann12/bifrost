package configstore

import (
	"fmt"
	"os/exec"
)

// Migration migrates a SQLite database to Postgres using pgloader.
// sqlitePath: path to the SQLite .db file
// postgresLink: connection string for Postgres (e.g. "pgsql://user:pass@host/dbname")
func Migration(sqlitePath, postgresLink string) error {
	cmd := exec.Command("pgloader",
		fmt.Sprintf("sqlite:///%s", sqlitePath),
		postgresLink,
	)

	// Show pgloader output in console
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pgloader failed: %w", err)
	}
	return nil
}
