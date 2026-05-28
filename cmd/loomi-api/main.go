package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/db"
	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/httpapi"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func main() {
	logger := diagnostics.NewJSONLogger(os.Stdout, slog.LevelInfo)
	opID := diagnostics.NewOperationID("startup")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("configuration failed", "operation_id", opID, "error", err.Error())
		os.Exit(1)
	}
	level, err := diagnostics.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Error("log level parsing failed", "operation_id", opID, "error", err.Error())
		os.Exit(1)
	}
	logger = diagnostics.NewJSONLogger(os.Stdout, level)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ReadinessTimeoutSeconds)*time.Second)
	defer cancel()

	// M2 starts even when Postgres is down; readiness exposes dependency state.
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Warn("database pool unavailable", "operation_id", opID, "database_url", cfg.RedactedDatabaseURL(), "error", "database pool creation failed")
	}
	if pool != nil {
		defer pool.Close()
	}

	product := productServiceForPool(pool)
	if product != nil {
		if _, err := product.SyncBuiltInPersonas(ctx, identityLocalDev(), productdata.BuiltInPersonas()); err != nil {
			logger.Warn("built-in persona sync failed", "operation_id", opID, "error", err.Error())
		}
		applySavedWebSearchConfig(ctx, product, &cfg)
	}
	broadcaster := productruntime.NewBroadcaster()
	providerConfigs := append(productruntime.ProviderConfigsFromConfig(cfg), savedProviderConfigs(ctx, product)...)
	providers := productruntime.NewHTTPProviders(providerConfigs, http.DefaultClient)
	gateway := productruntime.NewGateway(product, broadcaster, providers)
	localRunner := productruntime.NewLocalRunner(product, broadcaster)
	mcpConfigs := mcpServerConfigs(ctx, product)
	if product != nil && cfg.WorkerQueueEnabled && !cfg.WorkerQueuePaused {
		worker := productruntime.NewWorker(product, broadcaster, productruntime.QueuedRunRouter{Local: localRunner, Gateway: gateway, SandboxStore: sandboxProcessStore(product), MCPExecutor: productruntime.StdioMCPToolExecutor{Configs: mcpConfigs, ConfigLoader: func(loaderCtx context.Context) (map[string]productruntime.MCPServerConfig, error) {
			return mcpServerConfigs(loaderCtx, product), nil
		}}, WebExecutor: productruntime.WebToolExecutor{TavilyAPIKey: cfg.TavilyAPIKey, BraveAPIKey: cfg.BraveSearchAPIKey}})
		worker.LeaseSeconds = cfg.WorkerLeaseSeconds
		worker.PollInterval = time.Duration(cfg.WorkerPollMillis) * time.Millisecond
		worker.Start(context.Background())
	}
	server := httpapi.NewServerWithRuntimes(cfg, db.PostgresChecker{Pool: pool}, product, broadcaster, localRunner, gateway)
	logger.Info("loomi api starting", "operation_id", opID, "addr", cfg.HTTPAddr, "env", cfg.AppEnv)

	if err := http.ListenAndServe(cfg.HTTPAddr, server); err != nil {
		logger.Error("loomi api stopped", "operation_id", opID, "error", err.Error())
		os.Exit(1)
	}
}

func productServiceForPool(pool *pgxpool.Pool) productdata.Service {
	if pool == nil {
		return nil
	}
	return productdata.NewPostgresRepository(pool)
}

func sandboxProcessStore(product productdata.Service) *productruntime.SandboxProcessStore {
	repo, ok := product.(productdata.SandboxProcessRepository)
	if !ok {
		return nil
	}
	return productruntime.NewSandboxProcessStoreWithRepository(repo, productruntime.SandboxProcessStoreOptions{})
}

func identityLocalDev() identity.LocalIdentity {
	return identity.LocalDevIdentity()
}

func savedProviderConfigs(ctx context.Context, product productdata.Service) []productruntime.ProviderConfig {
	store, ok := product.(productdata.ModelProviderConfigStore)
	if !ok {
		return nil
	}
	saved, err := store.ListModelProviderConfigs(ctx, identityLocalDev())
	if err != nil {
		return nil
	}
	providers := make([]productruntime.ProviderConfig, 0, len(saved))
	for _, provider := range saved {
		providers = append(providers, productruntime.ProviderConfig{ID: provider.ID, Family: productruntime.ProviderFamily(provider.Family), BaseURL: provider.BaseURL, APIKey: provider.APIKey, Model: provider.Model, Enabled: provider.Enabled})
	}
	return providers
}

func applySavedWebSearchConfig(ctx context.Context, product productdata.Service, cfg *config.Config) {
	store, ok := product.(productdata.ModelProviderConfigStore)
	if !ok || cfg == nil {
		return
	}
	saved, err := store.GetWebSearchConfig(ctx, identityLocalDev())
	if err != nil {
		return
	}
	if saved.TavilyAPIKey != "" {
		cfg.TavilyAPIKey = saved.TavilyAPIKey
		_ = os.Setenv("LOOMI_TAVILY_API_KEY", saved.TavilyAPIKey)
	}
	if saved.BraveAPIKey != "" {
		cfg.BraveSearchAPIKey = saved.BraveAPIKey
		_ = os.Setenv("LOOMI_BRAVE_SEARCH_API_KEY", saved.BraveAPIKey)
	}
}

func mcpServerConfigs(ctx context.Context, product productdata.Service) map[string]productruntime.MCPServerConfig {
	configs, err := productruntime.MCPServerConfigsFromEnv()
	if err != nil || configs == nil {
		configs = map[string]productruntime.MCPServerConfig{}
	}
	for slug, config := range savedMCPServerConfigs(ctx, product) {
		configs[slug] = config
	}
	return configs
}

func savedMCPServerConfigs(ctx context.Context, product productdata.Service) map[string]productruntime.MCPServerConfig {
	store, ok := product.(productdata.MCPServerConfigStore)
	if !ok {
		return nil
	}
	saved, err := store.ListMCPServerConfigs(ctx, identityLocalDev())
	if err != nil {
		return nil
	}
	configs := map[string]productruntime.MCPServerConfig{}
	for _, record := range saved {
		config := productruntime.MCPServerConfig{Slug: record.Slug, DisplayName: record.DisplayName, Enabled: record.Enabled, Transport: productruntime.MCPTransport(record.Transport), Command: record.Command, Args: record.Args, Env: record.Env, TimeoutMS: record.TimeoutMS}
		validated, err := productruntime.ValidateMCPServerConfig(config)
		if err == nil {
			configs[validated.Slug] = validated
		}
	}
	return configs
}
