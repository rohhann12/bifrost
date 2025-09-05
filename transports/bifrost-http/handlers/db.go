package handlers

import (
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

	// Parse the Config struct from request body (new target config)
	var configStoreConfig configstore.Config
	if err := json.Unmarshal(ctx.PostBody(), &configStoreConfig); err != nil {
		SendError(ctx, fasthttp.StatusBadRequest,
			fmt.Sprintf("invalid config format: %v", err), h.logger)
		return
	}

	// Load existing config.json if it exists (current running config)
	var configData lib.ConfigData
	if _, err := os.Stat(configPath); err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			_ = json.Unmarshal(data, &configData)
		}
	}
	// figure out current and target db types
	var currentType configstore.ConfigStoreType
	if configData.ConfigStoreConfig != nil {
		currentType = configData.ConfigStoreConfig.Type
	}

	// migrate only if db type is changing
	if currentType != "" {
		switch {
		case currentType == "sqlite":
			// unmarshal old SQLite config
			var sqliteCfg configstore.SQLiteConfig
			if b, err := json.Marshal(configData.ConfigStoreConfig.Config); err == nil {
				_ = json.Unmarshal(b, &sqliteCfg)
			}

			// unmarshal new Postgres config
			var postgresCfg configstore.PostgresConfig
			if b, err := json.Marshal(configStoreConfig.Config); err == nil {
				_ = json.Unmarshal(b, &postgresCfg)
			}

			sqliteDSN := fmt.Sprintf("sqlite://%s", sqliteCfg.Path)
			postgresDSN := utils.CreatePostgresLink(
				postgresCfg.Host,
				postgresCfg.Port,
				postgresCfg.User,
				postgresCfg.Password,
				postgresCfg.DBName,
				postgresCfg.SSLMode,
			)
			h.logger.Info("1", sqliteDSN)
			h.logger.Info("1", postgresDSN)
			// Add the required third argument (e.g., context.Background() if it's a context.Context)
			if err := configstore.MigrateFromSql(sqliteDSN, postgresDSN, h.logger); err != nil {
				h.logger.Error(fmt.Sprintf("SQLite -> Postgres migration failed: %v", err))
			}

		case currentType == "postgres":
			// unmarshal old Postgres config
			var postgresCfg configstore.PostgresConfig
			if b, err := json.Marshal(configData.ConfigStoreConfig.Config); err == nil {
				_ = json.Unmarshal(b, &postgresCfg)
			}

			// unmarshal new SQLite config
			var sqliteCfg configstore.SQLiteConfig
			if b, err := json.Marshal(configStoreConfig.Config); err == nil {
				_ = json.Unmarshal(b, &sqliteCfg)
			}

			postgresDSN := utils.CreatePostgresLink(
				postgresCfg.Host,
				postgresCfg.Port,
				postgresCfg.User,
				postgresCfg.Password,
				postgresCfg.DBName,
				postgresCfg.SSLMode,
			)
			sqliteDSN := fmt.Sprintf("sqlite://%s", sqliteCfg.Path)

			if err := configstore.MigrateFromPostgres(postgresDSN, sqliteDSN, h.logger); err != nil {
				h.logger.Error(fmt.Sprintf("Postgres -> SQLite migration failed: %v", err))

			}
		}
	}

	// Update the config store config
	configData.ConfigStoreConfig = &configStoreConfig

	// Write back to config.json
	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		SendError(ctx, fasthttp.StatusInternalServerError,
			"failed to encode config", h.logger)
		return
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		SendError(ctx, fasthttp.StatusInternalServerError,
			"failed to write config.json", h.logger)
		return
	}

	SendJSON(ctx, map[string]any{
		"status":  "success",
		"message": "config.json updated successfully",
		"config":  configStoreConfig,
	}, h.logger)
}
