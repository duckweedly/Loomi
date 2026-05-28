package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	AppEnv                  string
	HTTPAddr                string
	DatabaseURL             string
	LogLevel                string
	ReadinessTimeoutSeconds int
	WorkerQueueEnabled      bool
	WorkerQueuePaused       bool
	WorkerLeaseSeconds      int
	WorkerMaxAttempts       int
	WorkerPollMillis        int
	ModelProviders          []ModelProvider
	TavilyAPIKey            string
	BraveSearchAPIKey       string
}

type ModelProvider struct {
	ID      string
	Family  string
	BaseURL string
	APIKey  string
	Model   string
	Enabled bool
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                  getenv("APP_ENV", "local"),
		HTTPAddr:                getenv("HTTP_ADDR", "127.0.0.1:18080"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		LogLevel:                getenv("LOG_LEVEL", "info"),
		ReadinessTimeoutSeconds: 5,
		WorkerQueueEnabled:      getenv("LOOMI_WORKER_QUEUE_ENABLED", "true") != "false",
		WorkerQueuePaused:       getenv("LOOMI_WORKER_QUEUE_PAUSED", "false") == "true",
		WorkerLeaseSeconds:      30,
		WorkerMaxAttempts:       3,
		WorkerPollMillis:        250,
		ModelProviders:          loadModelProviders(),
		TavilyAPIKey:            os.Getenv("LOOMI_TAVILY_API_KEY"),
		BraveSearchAPIKey:       os.Getenv("LOOMI_BRAVE_SEARCH_API_KEY"),
	}

	if raw := os.Getenv("READINESS_TIMEOUT_SECONDS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > 10 {
			return Config{}, errors.New("READINESS_TIMEOUT_SECONDS must be an integer from 1 to 10")
		}
		cfg.ReadinessTimeoutSeconds = value
	}
	if raw := os.Getenv("LOOMI_WORKER_LEASE_SECONDS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > 300 {
			return Config{}, errors.New("LOOMI_WORKER_LEASE_SECONDS must be an integer from 1 to 300")
		}
		cfg.WorkerLeaseSeconds = value
	}
	if raw := os.Getenv("LOOMI_WORKER_MAX_ATTEMPTS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > 10 {
			return Config{}, errors.New("LOOMI_WORKER_MAX_ATTEMPTS must be an integer from 1 to 10")
		}
		cfg.WorkerMaxAttempts = value
	}
	if raw := os.Getenv("LOOMI_WORKER_POLL_MILLIS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > 10000 {
			return Config{}, errors.New("LOOMI_WORKER_POLL_MILLIS must be an integer from 1 to 10000")
		}
		cfg.WorkerPollMillis = value
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	u, err := url.Parse(c.DatabaseURL)
	if err != nil || u.Scheme == "" || u.Host == "" || u.Scheme != "postgres" {
		return errors.New("DATABASE_URL must be a postgres connection URL")
	}
	if _, _, err := net.SplitHostPort(c.HTTPAddr); err != nil {
		return fmt.Errorf("HTTP_ADDR must be host:port: %w", err)
	}
	switch c.AppEnv {
	case "local", "test", "development":
	default:
		return errors.New("APP_ENV must be local, test, or development")
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return errors.New("LOG_LEVEL must be debug, info, warn, or error")
	}
	if c.WorkerLeaseSeconds <= 0 || c.WorkerLeaseSeconds > 300 {
		return errors.New("LOOMI_WORKER_LEASE_SECONDS must be an integer from 1 to 300")
	}
	if c.WorkerMaxAttempts <= 0 || c.WorkerMaxAttempts > 10 {
		return errors.New("LOOMI_WORKER_MAX_ATTEMPTS must be an integer from 1 to 10")
	}
	if c.WorkerPollMillis <= 0 || c.WorkerPollMillis > 10000 {
		return errors.New("LOOMI_WORKER_POLL_MILLIS must be an integer from 1 to 10000")
	}
	for _, provider := range c.ModelProviders {
		if provider.ID == "" {
			return errors.New("model provider id is required")
		}
		switch provider.Family {
		case "anthropic", "openai", "gemini", "openai_compatible":
		default:
			return errors.New("model provider family must be anthropic, openai, gemini, or openai_compatible")
		}
		if provider.Model == "" {
			return errors.New("model provider model is required")
		}
	}
	return nil
}

func (c Config) RedactedDatabaseURL() string {
	u, err := url.Parse(c.DatabaseURL)
	if err != nil {
		return "[redacted]"
	}
	if u.User != nil {
		username := u.User.Username()
		if username == "" {
			u.User = url.UserPassword("[redacted]", "[redacted]")
		} else {
			u.User = url.UserPassword(username, "[redacted]")
		}
	}
	return u.String()
}

func loadModelProviders() []ModelProvider {
	providers := make([]ModelProvider, 0, 4)
	if apiKey := os.Getenv("LOOMI_ANTHROPIC_API_KEY"); apiKey != "" {
		providers = append(providers, ModelProvider{ID: "anthropic", Family: "anthropic", APIKey: apiKey, Model: getenv("LOOMI_ANTHROPIC_MODEL", "claude-sonnet-4-6"), Enabled: true})
	}
	if apiKey := os.Getenv("LOOMI_OPENAI_API_KEY"); apiKey != "" {
		providers = append(providers, ModelProvider{ID: "openai", Family: "openai", APIKey: apiKey, Model: getenv("LOOMI_OPENAI_MODEL", "gpt-4.1"), Enabled: true})
	}
	if apiKey := os.Getenv("LOOMI_GEMINI_API_KEY"); apiKey != "" {
		providers = append(providers, ModelProvider{ID: "gemini", Family: "gemini", APIKey: apiKey, Model: getenv("LOOMI_GEMINI_MODEL", "gemini-3.5-flash"), Enabled: true})
	}
	if apiKey := os.Getenv("LOOMI_CUSTOM_MODEL_API_KEY"); apiKey != "" {
		providers = append(providers, ModelProvider{ID: getenv("LOOMI_CUSTOM_MODEL_PROVIDER_ID", "custom"), Family: "openai_compatible", BaseURL: os.Getenv("LOOMI_CUSTOM_MODEL_BASE_URL"), APIKey: apiKey, Model: os.Getenv("LOOMI_CUSTOM_MODEL_NAME"), Enabled: true})
	}
	return providers
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
