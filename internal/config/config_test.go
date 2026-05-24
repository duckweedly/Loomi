package config

import "testing"

func TestLoadAppliesLocalDefaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://loomi:secret@127.0.0.1:55432/loomi?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AppEnv != "local" {
		t.Fatalf("AppEnv = %q, want local", cfg.AppEnv)
	}
	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q", cfg.LogLevel)
	}
	if cfg.ReadinessTimeoutSeconds != 5 {
		t.Fatalf("ReadinessTimeoutSeconds = %d", cfg.ReadinessTimeoutSeconds)
	}
	if !cfg.WorkerQueueEnabled || cfg.WorkerQueuePaused || cfg.WorkerLeaseSeconds != 30 || cfg.WorkerMaxAttempts != 3 || cfg.WorkerPollMillis != 250 {
		t.Fatalf("worker defaults = %+v", cfg)
	}
}

func TestLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadRejectsMalformedDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "not-a-url")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error")
	}
}

func TestLoadRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
	}{
		{name: "app env", key: "APP_ENV", val: "production"},
		{name: "http addr", key: "HTTP_ADDR", val: "127.0.0.1"},
		{name: "log level", key: "LOG_LEVEL", val: "trace"},
		{name: "timeout text", key: "READINESS_TIMEOUT_SECONDS", val: "slow"},
		{name: "timeout too high", key: "READINESS_TIMEOUT_SECONDS", val: "30"},
		{name: "lease text", key: "LOOMI_WORKER_LEASE_SECONDS", val: "slow"},
		{name: "lease too high", key: "LOOMI_WORKER_LEASE_SECONDS", val: "301"},
		{name: "attempts zero", key: "LOOMI_WORKER_MAX_ATTEMPTS", val: "0"},
		{name: "poll too high", key: "LOOMI_WORKER_POLL_MILLIS", val: "10001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://loomi:secret@127.0.0.1:55432/loomi?sslmode=disable")
			t.Setenv(tt.key, tt.val)
			if _, err := Load(); err == nil {
				t.Fatal("Load() error = nil, want error")
			}
		})
	}
}

func TestLoadWorkerQueueConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://loomi:secret@127.0.0.1:55432/loomi?sslmode=disable")
	t.Setenv("LOOMI_WORKER_QUEUE_ENABLED", "false")
	t.Setenv("LOOMI_WORKER_QUEUE_PAUSED", "true")
	t.Setenv("LOOMI_WORKER_LEASE_SECONDS", "45")
	t.Setenv("LOOMI_WORKER_MAX_ATTEMPTS", "5")
	t.Setenv("LOOMI_WORKER_POLL_MILLIS", "500")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.WorkerQueueEnabled || !cfg.WorkerQueuePaused || cfg.WorkerLeaseSeconds != 45 || cfg.WorkerMaxAttempts != 5 || cfg.WorkerPollMillis != 500 {
		t.Fatalf("worker config = %+v", cfg)
	}
}

func TestLoadModelProviderConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://loomi:secret@127.0.0.1:55432/loomi?sslmode=disable")
	t.Setenv("LOOMI_CUSTOM_MODEL_API_KEY", "test-key")
	t.Setenv("LOOMI_CUSTOM_MODEL_BASE_URL", "https://example.test/v1")
	t.Setenv("LOOMI_CUSTOM_MODEL_NAME", "test-model")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.ModelProviders) != 1 {
		t.Fatalf("len(ModelProviders) = %d, want 1", len(cfg.ModelProviders))
	}
	provider := cfg.ModelProviders[0]
	if provider.ID != "custom" || provider.Family != "openai_compatible" || provider.Model != "test-model" || !provider.Enabled {
		t.Fatalf("ModelProviders[0] = %+v", provider)
	}
}

func TestRedactedDatabaseURL(t *testing.T) {
	cfg := Config{DatabaseURL: "postgres://loomi:secret@127.0.0.1:55432/loomi?sslmode=disable"}
	got := cfg.RedactedDatabaseURL()
	if got == "" || got == cfg.DatabaseURL {
		t.Fatalf("RedactedDatabaseURL() = %q", got)
	}
	if contains(got, "secret") {
		t.Fatalf("RedactedDatabaseURL() leaked secret: %q", got)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
