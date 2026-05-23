package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/db"
	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	seedThreadID        = "thr_local_demo"
	seedThreadTitle     = "M3 local demo thread"
	seedMessageID       = "msg_local_demo_001"
	seedMessageContent  = "This is a durable local user message seeded for M3."
	seedClientMessageID = "seed-m3-local-demo-message"
)

type seedResult struct {
	ThreadID  string
	MessageID string
}

func main() {
	logger := diagnostics.NewJSONLogger(os.Stdout, slog.LevelInfo)
	opID := diagnostics.NewOperationID("seed")
	cfg, err := config.Load()
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	level, err := diagnostics.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	logger = diagnostics.NewJSONLogger(os.Stdout, level)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ReadinessTimeoutSeconds)*time.Second)
	defer cancel()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	result, err := runSeed(ctx, productdata.NewPostgresRepository(pool), identity.LocalDevIdentity())
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	logger.Info("m3 seed complete", "component", "seed", "operation_id", opID, "thread_id", result.ThreadID, "message_id", result.MessageID)
}

func runSeed(ctx context.Context, svc productdata.SeedService, ident identity.LocalIdentity) (seedResult, error) {
	if _, err := svc.CurrentIdentity(ctx, ident); err != nil {
		return seedResult{}, err
	}
	thread, err := svc.UpsertSeedThread(ctx, ident, productdata.SeedThreadInput{ID: seedThreadID, Title: seedThreadTitle, Mode: productdata.ThreadModeChat})
	if err != nil {
		return seedResult{}, err
	}
	message, err := svc.UpsertSeedMessage(ctx, ident, productdata.SeedMessageInput{ID: seedMessageID, ThreadID: thread.ID, Content: seedMessageContent, ClientMessageID: seedClientMessageID})
	if err != nil {
		return seedResult{}, err
	}
	return seedResult{ThreadID: thread.ID, MessageID: message.ID}, nil
}
