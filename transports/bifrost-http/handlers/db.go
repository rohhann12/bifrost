package handlers

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fasthttp/router"
	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"
	"github.com/maximhq/bifrost/framework/configstore"
	"github.com/maximhq/bifrost/transports/bifrost-http/lib"
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

// Get current state of the DbType - by default we check config.json.
// If that exists, return the parsed config. Otherwise return the response from last configuration done as the user can change path
// for sqlite as well it is not that everyimt it will have
// config.db
func (h *DbHandler) GetDbState(ctx *fasthttp.RequestCtx) {
	configPath := "config.json"
	var cfg configstore.Config

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		ctx.Response.Header.Set("Content-Type", "application/json")
		_ = json.NewEncoder(ctx).Encode(map[string]string{
			"db_type": "sqlite", // fallback default
			"path":    string(cfg.Type),
		})
		h.logger.Info(fmt.Sprintf("cfg.Type: %s", cfg.Type))

		cfgJson, _ := json.Marshal(cfg.Config)
		h.logger.Info(fmt.Sprintf("cfg.Config: %s", string(cfgJson)))

		h.logger.Info(fmt.Sprintf("cfg.Enabled: %v", cfg.Enabled))
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBody([]byte(`{"error":"failed to read config.json"}`))
		return
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBody([]byte(`{"error":"invalid config.json"}`))
		return
	}

	var resp map[string]any
	switch cfg.Type {
	case configstore.ConfigStoreTypeSQLite:
		sqliteCfg, ok := cfg.Config.(map[string]any)
		if !ok {
			ctx.SetStatusCode(500)
			ctx.SetBody([]byte(`{"error":"invalid sqlite config"}`))

			h.logger.Info("hitting this")
			return
		}
		resp = map[string]any{
			"db_type": "sqlite",
			"path":    sqliteCfg["path"],
		}

	case configstore.ConfigStoreTypePostgres:
		pgCfg, ok := cfg.Config.(map[string]any)
		if !ok {
			ctx.SetStatusCode(500)
			ctx.SetBody([]byte(`{"error":"invalid postgres config"}`))
			return
		}
		resp = map[string]any{
			"db_type":  "postgres",
			"host":     pgCfg["host"],
			"port":     pgCfg["port"],
			"user":     pgCfg["user"],
			"dbname":   pgCfg["dbName"],
			"ssl_mode": pgCfg["sslMode"],
		}

	default:
		resp = map[string]any{"db_type": "unknown"}
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	_ = json.NewEncoder(ctx).Encode(resp)
}

func (h *DbHandler) UpdateDbState(ctx *fasthttp.RequestCtx) {
	configPath := "config.json"

	// expected input: { "db_type": "postgres" }
	var body struct {
		DbType string `json:"db_type"`
	}
	if err := json.Unmarshal(ctx.PostBody(), &body); err != nil {
		ctx.SetStatusCode(400)
		ctx.SetBody([]byte(`{"error":"invalid request body"}`))
		return
	}

	// load existing config if present
	var cfg lib.Config
	if _, err := os.Stat(configPath); err == nil {
		if data, err := os.ReadFile(configPath); err == nil {
			_ = json.Unmarshal(data, &cfg) // best effort
		}
	}

	// update only DbType
	cfg.DbType = body.DbType

	// write back to file
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBody([]byte(`{"error":"failed to encode config"}`))
		return
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		ctx.SetStatusCode(500)
		ctx.SetBody([]byte(`{"error":"failed to write config.json"}`))
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json")
	_ = json.NewEncoder(ctx).Encode(map[string]string{"db_type": cfg.DbType})
}
