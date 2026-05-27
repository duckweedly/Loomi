package cli

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Path     string `json:"path"`
	Found    bool   `json:"found"`
	Host     string `json:"host"`
	Mode     string `json:"mode"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Persona  string `json:"persona"`
	Script   string `json:"script"`
}

func LoadConfigFromEnv() (Config, error) {
	cfg, err := LoadConfigFileFromEnv()
	if err != nil {
		return Config{}, err
	}
	applyEnvDefault(&cfg.Host, "LOOMI_HOST")
	applyEnvDefault(&cfg.Mode, "LOOMI_MODE")
	applyEnvDefault(&cfg.Provider, "LOOMI_PROVIDER")
	applyEnvDefault(&cfg.Model, "LOOMI_MODEL")
	applyEnvDefault(&cfg.Persona, "LOOMI_PERSONA")
	applyEnvDefault(&cfg.Script, "LOOMI_SCRIPT")
	if strings.TrimSpace(cfg.Host) == "" {
		cfg.Host = DefaultBaseURL
	}
	if strings.TrimSpace(cfg.Mode) == "" {
		cfg.Mode = "work"
	}
	if strings.TrimSpace(cfg.Provider) == "" && strings.TrimSpace(cfg.Script) == "" {
		cfg.Provider = "local_codex"
	}
	return cfg, nil
}

func LoadConfigFileFromEnv() (Config, error) {
	return LoadConfigFile(defaultConfigPath())
}

func LoadConfigFile(path string) (Config, error) {
	cfg := defaultConfig(path)
	if cfg.Path == "" {
		return cfg, nil
	}
	raw, err := os.ReadFile(cfg.Path)
	if err == nil {
		cfg.Found = true
		path := cfg.Path
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return Config{}, err
		}
		cfg.Path = path
		cfg.Found = true
		return normalizeConfig(cfg), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}
	return cfg, nil
}

func SaveConfigFile(cfg Config) error {
	path := strings.TrimSpace(cfg.Path)
	if path == "" {
		return errors.New("loomi config path is unavailable")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	cfg = normalizeConfig(cfg)
	raw, err := json.MarshalIndent(struct {
		Host     string `json:"host,omitempty"`
		Mode     string `json:"mode,omitempty"`
		Provider string `json:"provider,omitempty"`
		Model    string `json:"model,omitempty"`
		Persona  string `json:"persona,omitempty"`
		Script   string `json:"script,omitempty"`
	}{
		Host:     cfg.Host,
		Mode:     cfg.Mode,
		Provider: cfg.Provider,
		Model:    cfg.Model,
		Persona:  cfg.Persona,
		Script:   cfg.Script,
	}, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0o600)
}

func SetConfigValue(cfg *Config, key string, value string) error {
	if cfg == nil {
		return errors.New("loomi config is unavailable")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("loomi config value cannot be empty")
	}
	return setConfigValue(cfg, key, value)
}

func UnsetConfigValue(cfg *Config, key string) error {
	if cfg == nil {
		return errors.New("loomi config is unavailable")
	}
	return setConfigValue(cfg, key, "")
}

func setConfigValue(cfg *Config, key string, value string) error {
	switch strings.TrimSpace(key) {
	case "host":
		cfg.Host = value
	case "mode":
		cfg.Mode = value
	case "provider":
		cfg.Provider = value
	case "model":
		cfg.Model = value
	case "persona":
		cfg.Persona = value
	case "script":
		cfg.Script = value
	default:
		return errors.New("supported config keys: host, mode, provider, model, persona, script")
	}
	return nil
}

func defaultConfig(path string) Config {
	return Config{
		Path: strings.TrimSpace(path),
	}
}

func normalizeConfig(cfg Config) Config {
	cfg.Host = strings.TrimSpace(cfg.Host)
	cfg.Mode = strings.TrimSpace(cfg.Mode)
	cfg.Provider = strings.TrimSpace(cfg.Provider)
	cfg.Model = strings.TrimSpace(cfg.Model)
	cfg.Persona = strings.TrimSpace(cfg.Persona)
	cfg.Script = strings.TrimSpace(cfg.Script)
	return cfg
}

func defaultConfigPath() string {
	if path := strings.TrimSpace(os.Getenv("LOOMI_CONFIG")); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".loomi", "config.json")
}

func applyEnvDefault(value *string, env string) {
	if raw := strings.TrimSpace(os.Getenv(env)); raw != "" {
		*value = raw
	}
}
