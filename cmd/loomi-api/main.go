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
	broadcaster := productruntime.NewBroadcaster()
	providers := productruntime.NewHTTPProviders(productruntime.ProviderConfigsFromConfig(cfg), http.DefaultClient)
	gateway := productruntime.NewGateway(product, broadcaster, providers)
	localRunner := productruntime.NewLocalRunner(product, broadcaster)
	if product != nil && cfg.WorkerQueueEnabled && !cfg.WorkerQueuePaused {
		worker := productruntime.NewWorker(product, broadcaster, productruntime.QueuedRunRouter{Local: localRunner, Gateway: gateway})
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
