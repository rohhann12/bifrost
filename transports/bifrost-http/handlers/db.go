package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fasthttp/router"
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/configstore"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
	"github.com/maximhq/bifrost/transports/bifrost-http/utils"
	"github.com/valyala/fasthttp"
)

type DbHandler struct {
	client *bifrost.Bifrost
	logger schemas.Logger
	store  *lib.Config
}

func NewDbHandler(client *bifrost.Bifrost, logger schemas.Logger, store *lib.Config) *DbHandler {
	return &DbHandler{
		client: client,
		logger: logger,
		store:  store,
	}
}

func (h *DbHandler) RegisterRoutes(r *router.Router) {
	r.GET("/api/db", h.GetDbState)
	r.POST("/api/db", h.UpdateDbState)
}

// GetDbState returns the current database configuration from memory
func (h *DbHandler) GetDbState(ctx *fasthttp.RequestCtx) {
	configPath := filepath.Join(filepath.Dir(h.store.ConfigPath), "config.json")

	// Check if config.json exists
	if _, err := os.Stat(configPath); err == nil {
		// config.json exists - read from file
		data, err := os.ReadFile(configPath)
		if err != nil {
			SendError(ctx, fasthttp.StatusInternalServerError,
				fmt.Sprintf("failed to read config.json: %v", err), h.logger)
			return
		}

		var configData lib.ConfigData
		if err := json.Unmarshal(data, &configData); err != nil {
			SendError(ctx, fasthttp.StatusInternalServerError,
				fmt.Sprintf("failed to parse config.json: %v", err), h.logger)
			return
		}

		// Return config store config from file
		if configData.ConfigStoreConfig != nil {
			SendJSON(ctx, map[string]any{
				"enabled": configData.ConfigStoreConfig.Enabled,
				"type":    string(configData.ConfigStoreConfig.Type),
				"config":  configData.ConfigStoreConfig.Config,
			}, h.logger)
			return
		}
	}

	// No config.json exists - return SQLite config from memory (default)
	SendJSON(ctx, map[string]any{
		"enabled": true,
		"type":    "sqlite",
		"config": map[string]any{
			"path": h.store.ConfigPath,
		},
	}, h.logger)
}

// UpdateDbState creates or updates config.json with the provided Config struct
func (h *DbHandler) UpdateDbState(ctx *fasthttp.RequestCtx) {
	configPath := filepath.Join(filepath.Dir(h.store.ConfigPath), "config.json")

	var newConfig configstore.Config
	if err := json.Unmarshal(ctx.PostBody(), &newConfig); err != nil {
		SendError(ctx, fasthttp.StatusBadRequest,
			fmt.Sprintf("invalid config format: %v", err), h.logger)
		return
	}
	h.logger.Info("Successfully unmarshaled data: %+v", newConfig)

	var configData lib.ConfigData
	if _, err := os.Stat(configPath); err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			_ = json.Unmarshal(data, &configData)
		}
	}

	// Figure out current and target db types
	var oldConfig *configstore.Config
	if configData.ConfigStoreConfig != nil {
		oldConfig = configData.ConfigStoreConfig
	}

	oldType := ""
	if oldConfig != nil {
		oldType = string(oldConfig.Type)
	}

	if oldType == "" {
		if newConfig.Type == "postgres" {
			oldType = "sqlite"
			h.logger.Info("No old DB type found, defaulting to sqlite")
		}
	}

	newType := string(newConfig.Type)

	// Decide if migration is needed
	needsMigration := false
	if oldConfig == nil {
		if oldType != "" && oldType != newType {
			needsMigration = true
		}
	} else {
		if oldType != newType {
			needsMigration = true
		} else {
			// same type -> check if config values differ
			oldBytes, _ := json.Marshal(oldConfig.Config)
			newBytes, _ := json.Marshal(newConfig.Config)
			if !bytes.Equal(oldBytes, newBytes) {
				needsMigration = true
			}
		}
	}

	// Perform migration if needed
	if needsMigration {
		switch oldType {
		case "sqlite":
			var sqliteCfg configstore.SQLiteConfig
			if oldConfig != nil {
				b, _ := json.Marshal(oldConfig.Config)
				_ = json.Unmarshal(b, &sqliteCfg)
			} else {
				// fallback: oldConfig nil -> assume default sqlite path
				sqliteCfg.Path = filepath.Join(filepath.Dir(h.store.ConfigPath), "config.db")
			}

			var postgresCfg configstore.PostgresConfig
			b, _ := json.Marshal(newConfig.Config)
			_ = json.Unmarshal(b, &postgresCfg)

			sqliteDSN := sqliteCfg.Path
			postgresDSN := utils.CreatePostgresLink(
				postgresCfg.Host,
				postgresCfg.Port,
				postgresCfg.User,
				postgresCfg.Password,
				postgresCfg.DBName,
				postgresCfg.SSLMode,
			)

			if err := configstore.MigrateFromSql(sqliteDSN, postgresDSN, h.logger); err != nil {
				SendError(ctx, fasthttp.StatusInternalServerError,
					fmt.Sprintf("SQLite -> Postgres migration failed: %v", err), h.logger)
				return
			}

		case "postgres":
			var postgresCfg configstore.PostgresConfig
			b, _ := json.Marshal(oldConfig.Config)
			_ = json.Unmarshal(b, &postgresCfg)

			var sqliteCfg configstore.SQLiteConfig
			b, _ = json.Marshal(newConfig.Config)
			_ = json.Unmarshal(b, &sqliteCfg)

			postgresDSN := utils.CreatePostgresLink(
				postgresCfg.Host,
				postgresCfg.Port,
				postgresCfg.User,
				postgresCfg.Password,
				postgresCfg.DBName,
				postgresCfg.SSLMode,
			)
			sqliteDSN := sqliteCfg.Path

			if err := configstore.MigrateFromPostgres(sqliteDSN, postgresDSN, h.logger); err != nil {
				SendError(ctx, fasthttp.StatusInternalServerError,
					fmt.Sprintf("Postgres -> SQLite migration failed: %v", err), h.logger)
				return
			}
		}
	}

	// Update config and write back to config.json
	configData.ConfigStoreConfig = &newConfig

	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		SendError(ctx, fasthttp.StatusInternalServerError,
			"failed to encode config", h.logger)
		return
	}
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		SendError(ctx, fasthttp.StatusInternalServerError,
			"failed to write config.json", h.logger)
		return
	}

	SendJSON(ctx, map[string]any{
		"status":  "success",
		"message": "config.json updated successfully",
		"config":  newConfig,
	}, h.logger)
}
