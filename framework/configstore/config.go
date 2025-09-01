package configstore

import (
	"encoding/json"
	"fmt"
)

// ConfigStoreType represents the type of config store.
type ConfigStoreType string

// Supported config store types.
const (
	ConfigStoreTypeSQLite   ConfigStoreType = "sqlite"
	ConfigStoreTypePostgres ConfigStoreType = "postgres"
)

// Config represents the configuration for the config store.
type Config struct {
	Enabled bool            `json:"enabled"`
	Type    ConfigStoreType `json:"type"`
	Config  any             `json:"config"`
}

// SQLiteConfig represents configuration for SQLite.
type SQLiteConfig struct {
	Path string `json:"path"`
}

// PostgresConfig represents configuration for Postgres.
type PostgresConfig struct {
	ConnectionString string `json:"connectionString"`
}

// UnmarshalJSON implements custom unmarshaling for Config.
func (c *Config) UnmarshalJSON(data []byte) error {
	// Temporary struct to extract type and raw config
	type TempConfig struct {
		Enabled bool            `json:"enabled"`
		Type    ConfigStoreType `json:"type"`
		Config  json.RawMessage `json:"config"`
	}

	var temp TempConfig
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("failed to unmarshal config store config: %w", err)
	}

	// Assign basic fields
	c.Enabled = temp.Enabled
	c.Type = temp.Type

	// If disabled, no further parsing needed
	if !temp.Enabled {
		c.Config = nil
		return nil
	}

	// Parse based on type
	switch temp.Type {
	case ConfigStoreTypeSQLite:
		var sqliteConfig SQLiteConfig
		if err := json.Unmarshal(temp.Config, &sqliteConfig); err != nil {
			return fmt.Errorf("failed to unmarshal sqlite config: %w", err)
		}
		c.Config = &sqliteConfig

	case ConfigStoreTypePostgres:
		var pgConfig PostgresConfig
		if err := json.Unmarshal(temp.Config, &pgConfig); err != nil {
			return fmt.Errorf("failed to unmarshal postgres config: %w", err)
		}
		c.Config = &pgConfig

	default:
		return fmt.Errorf("unknown config store type: %s", temp.Type)
	}

	return nil
}
