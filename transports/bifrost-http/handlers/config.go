package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/fasthttp/router"
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/configstore"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
	"github.com/maximhq/bifrost/transports/bifrost-http/utils"
	"github.com/valyala/fasthttp"
)

// ConfigHandler manages runtime configuration updates for Bifrost.
// It provides endpoints to update and retrieve settings persisted via the ConfigStore backed by sql database.
type ConfigHandler struct {
	client *bifrost.Bifrost
	logger schemas.Logger
	store  *lib.Config
	appDir string
}

// NewConfigHandler creates a new handler for configuration management.
// It requires the Bifrost client, a logger, and the config store.
func NewConfigHandler(client *bifrost.Bifrost, logger schemas.Logger, store *lib.Config) *ConfigHandler {
	return &ConfigHandler{
		client: client,
		logger: logger,
		store:  store,
	}
}

// RegisterRoutes registers the configuration-related routes.
// It adds the `PUT /api/config` endpoint.
func (h *ConfigHandler) RegisterRoutes(r *router.Router) {
	r.GET("/api/config", h.GetConfig)
	r.PUT("/api/config", h.handleUpdateConfig)
}

// GetConfig handles GET /config - Get the current configuration
func (h *ConfigHandler) GetConfig(ctx *fasthttp.RequestCtx) {

	var mapConfig = make(map[string]any)

	if query := string(ctx.QueryArgs().Peek("from_db")); query == "true" {
		if h.store.ConfigStore == nil {
			SendError(ctx, fasthttp.StatusServiceUnavailable, "config store not available", h.logger)
			return
		}
		cc, err := h.store.ConfigStore.GetClientConfig()
		if err != nil {
			SendError(ctx, fasthttp.StatusInternalServerError,
				fmt.Sprintf("failed to fetch config from db: %v", err), h.logger)
			return
		}
		if cc != nil {
			mapConfig["client_config"] = *cc
			h.logger.Info("cc type: %T", *cc)
		}
	} else {
		mapConfig["client_config"] = h.store.ClientConfig
	}

	mapConfig["is_db_connected"] = h.store.ConfigStore != nil
	mapConfig["is_cache_connected"] = h.store.VectorStore != nil
	mapConfig["is_logs_connected"] = h.store.LogsStore != nil
	h.logger.Info("Final ClientConfig: %+v", mapConfig)

	SendJSON(ctx, mapConfig, h.logger)
}

// handleUpdateConfig updates the core configuration settings.
// It supports hot-reloading of db config and client config.
func (h *ConfigHandler) handleUpdateConfig(ctx *fasthttp.RequestCtx) {
	if h.store.ConfigStore == nil {
		SendError(ctx, fasthttp.StatusInternalServerError, "Config store not initialized", h.logger)
		return
	}

	var req configstore.ClientConfig
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		SendError(ctx, fasthttp.StatusBadRequest, fmt.Sprintf("Invalid request format: %v", err), h.logger)
		return
	}

	// copy current config
	currentConfig := h.store.ClientConfig
	updatedConfig := currentConfig

	// hot-reload settings
	if req.DropExcessRequests != currentConfig.DropExcessRequests {
		h.client.UpdateDropExcessRequests(req.DropExcessRequests)
		updatedConfig.DropExcessRequests = req.DropExcessRequests
	}
	if !slices.Equal(req.PrometheusLabels, currentConfig.PrometheusLabels) {
		updatedConfig.PrometheusLabels = req.PrometheusLabels
	}
	if !slices.Equal(req.AllowedOrigins, currentConfig.AllowedOrigins) {
		updatedConfig.AllowedOrigins = req.AllowedOrigins
	}

	updatedConfig.InitialPoolSize = req.InitialPoolSize
	updatedConfig.EnableLogging = req.EnableLogging
	updatedConfig.EnableGovernance = req.EnableGovernance
	updatedConfig.EnforceGovernanceHeader = req.EnforceGovernanceHeader
	updatedConfig.AllowDirectKeys = req.AllowDirectKeys

	// ✅ update in-memory config
	h.store.ClientConfig = updatedConfig

	// ✅ persist in db
	if err := h.store.ConfigStore.UpdateClientConfig(&updatedConfig); err != nil {
		h.logger.Warn(fmt.Sprintf("failed to save configuration: %v", err))
		SendError(ctx, fasthttp.StatusInternalServerError, fmt.Sprintf("failed to save configuration: %v", err), h.logger)
		return
	}

	// ✅ also save to config.json
	configDir := utils.GetDefaultConfigDir(h.appDir)
	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(updatedConfig, "", "  ")
	if err != nil {
		h.logger.Warn(fmt.Sprintf("failed to marshal config.json: %v", err))
	} else {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			h.logger.Warn(fmt.Sprintf("failed to create config dir: %v", err))
		} else if err := os.WriteFile(configPath, data, 0644); err != nil {
			h.logger.Warn(fmt.Sprintf("failed to write config.json: %v", err))
		} else {
			h.logger.Info(fmt.Sprintf("config.json updated at %s", configPath))
		}
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	SendJSON(ctx, map[string]any{
		"status":  "success",
		"message": "configuration updated successfully",
	}, h.logger)
}
