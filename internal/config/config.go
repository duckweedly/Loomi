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
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                  getenv("APP_ENV", "local"),
		HTTPAddr:                getenv("HTTP_ADDR", "127.0.0.1:8080"),
		DatabaseURL:             os.Getenv("DATABASE_URL"),
		LogLevel:                getenv("LOG_LEVEL", "info"),
		ReadinessTimeoutSeconds: 5,
	}

	if raw := os.Getenv("READINESS_TIMEOUT_SECONDS"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > 10 {
			return Config{}, errors.New("READINESS_TIMEOUT_SECONDS must be an integer from 1 to 10")
		}
		cfg.ReadinessTimeoutSeconds = value
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

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
