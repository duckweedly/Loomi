package runtime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type LocalProviderAuthMode string

const (
	LocalProviderAuthModeAPIKey  LocalProviderAuthMode = "api_key"
	LocalProviderAuthModeOAuth   LocalProviderAuthMode = "oauth"
	LocalProviderAuthModeUnknown LocalProviderAuthMode = "unknown"
)

type LocalProviderStatus string

const (
	LocalProviderStatusAvailable   LocalProviderStatus = "available"
	LocalProviderStatusUnavailable LocalProviderStatus = "unavailable"
	LocalProviderStatusNeedsLogin  LocalProviderStatus = "needs_login"
	LocalProviderStatusUnsupported LocalProviderStatus = "unsupported"
	LocalProviderStatusDisabled    LocalProviderStatus = "disabled"
)

type LocalProviderSource string

const (
	LocalProviderSourceLocalConfig       LocalProviderSource = "local_config"
	LocalProviderSourceEnv               LocalProviderSource = "env"
	LocalProviderSourceKeychainUnchecked LocalProviderSource = "keychain_unchecked"
	LocalProviderSourceUnknown           LocalProviderSource = "unknown"
)

type LocalProviderKind string

const (
	LocalProviderKindClaudeCode LocalProviderKind = "claude_code"
	LocalProviderKindCodex      LocalProviderKind = "codex"
)

type LocalProviderDetectionInput struct {
	HomeDir         string
	CodexHome       string
	ClaudeConfigDir string
	Env             map[string]string
	Disabled        bool
}

type LocalProviderCapability struct {
	ProviderID       string                `json:"provider_id"`
	DisplayName      string                `json:"display_name"`
	ProviderKind     LocalProviderKind     `json:"provider_kind"`
	AuthMode         LocalProviderAuthMode `json:"auth_mode"`
	Status           LocalProviderStatus   `json:"status"`
	ModelCandidates  []string              `json:"model_candidates"`
	Source           LocalProviderSource   `json:"source"`
	RedactionApplied bool                  `json:"redaction_applied"`
	Message          string                `json:"message,omitempty"`
}

func DetectLocalProviders(input LocalProviderDetectionInput) []LocalProviderCapability {
	if input.Disabled {
		return []LocalProviderCapability{
			baseClaudeLocalProvider(LocalProviderStatusDisabled, LocalProviderAuthModeUnknown, LocalProviderSourceUnknown, "Local provider autodetect is disabled."),
			baseCodexLocalProvider(LocalProviderStatusDisabled, LocalProviderAuthModeUnknown, LocalProviderSourceUnknown, "Local provider autodetect is disabled."),
		}
	}
	return []LocalProviderCapability{
		detectClaudeCodeLocalProvider(input),
		detectCodexLocalProvider(input),
	}
}

func LocalProviderDetectionInputFromProcess() LocalProviderDetectionInput {
	env := map[string]string{}
	for _, pair := range os.Environ() {
		key, value, ok := strings.Cut(pair, "=")
		if ok {
			env[key] = value
		}
	}
	home, _ := os.UserHomeDir()
	return LocalProviderDetectionInput{HomeDir: home, CodexHome: env["CODEX_HOME"], ClaudeConfigDir: env["CLAUDE_CONFIG_DIR"], Env: env}
}

func detectClaudeCodeLocalProvider(input LocalProviderDetectionInput) LocalProviderCapability {
	home := strings.TrimSpace(input.HomeDir)
	claudeDir := strings.TrimSpace(input.ClaudeConfigDir)
	if claudeDir == "" && home != "" {
		claudeDir = filepath.Join(home, ".claude")
	}

	if home != "" {
		rootConfig := readJSONObject(filepath.Join(home, ".claude.json"))
		if nonEmptyString(rootConfig, "primaryApiKey") {
			return baseClaudeLocalProvider(LocalProviderStatusAvailable, LocalProviderAuthModeAPIKey, LocalProviderSourceLocalConfig, "Detected but not enabled. Explicit opt-in is required before use.")
		}
		if nonEmptyString(rootConfig, "apiKeyHelper") {
			return baseClaudeLocalProvider(LocalProviderStatusUnsupported, LocalProviderAuthModeUnknown, LocalProviderSourceLocalConfig, "apiKeyHelper is present but helper execution is unsupported.")
		}
	}

	settings := readJSONObject(filepath.Join(claudeDir, "settings.json"))
	if env, ok := settings["env"].(map[string]any); ok {
		models := []string{"claude-sonnet-4-5"}
		if model, ok := env["ANTHROPIC_MODEL"].(string); ok && strings.TrimSpace(model) != "" {
			models = []string{strings.TrimSpace(model)}
		}
		if nonEmptyString(env, "ANTHROPIC_AUTH_TOKEN") {
			provider := baseClaudeLocalProvider(LocalProviderStatusAvailable, LocalProviderAuthModeOAuth, LocalProviderSourceLocalConfig, "Detected settings env. Explicit opt-in is required before use; no secrets are shown.")
			provider.ModelCandidates = models
			return provider
		}
		if nonEmptyString(env, "ANTHROPIC_BASE_URL") || nonEmptyString(env, "ANTHROPIC_MODEL") {
			provider := baseClaudeLocalProvider(LocalProviderStatusNeedsLogin, LocalProviderAuthModeUnknown, LocalProviderSourceLocalConfig, "Settings were found but login material was not detected.")
			provider.ModelCandidates = models
			return provider
		}
	}

	credentials := readJSONObject(filepath.Join(claudeDir, ".credentials.json"))
	if hasOAuthToken(credentials) {
		return baseClaudeLocalProvider(LocalProviderStatusAvailable, LocalProviderAuthModeOAuth, LocalProviderSourceLocalConfig, "Detected OAuth credentials but not enabled. Explicit opt-in is required before use.")
	}
	if len(credentials) > 0 {
		return baseClaudeLocalProvider(LocalProviderStatusNeedsLogin, LocalProviderAuthModeOAuth, LocalProviderSourceLocalConfig, "OAuth credentials were found but usable token presence was not detected.")
	}

	return baseClaudeLocalProvider(LocalProviderStatusUnavailable, LocalProviderAuthModeUnknown, LocalProviderSourceUnknown, "Not detected.")
}

func detectCodexLocalProvider(input LocalProviderDetectionInput) LocalProviderCapability {
	env := input.Env
	if env == nil {
		env = map[string]string{}
	}
	if strings.TrimSpace(env["CODEX_API_KEY"]) != "" {
		return baseCodexLocalProvider(LocalProviderStatusAvailable, LocalProviderAuthModeAPIKey, LocalProviderSourceEnv, "Detected from environment. Explicit opt-in is required before use.")
	}
	if strings.TrimSpace(env["OPENAI_API_KEY"]) != "" {
		return baseCodexLocalProvider(LocalProviderStatusAvailable, LocalProviderAuthModeAPIKey, LocalProviderSourceEnv, "Detected from environment. Explicit opt-in is required before use.")
	}

	authPath := codexAuthPath(input)
	auth := readJSONObject(authPath)
	if nonEmptyString(auth, "CODEX_API_KEY") || nonEmptyString(auth, "OPENAI_API_KEY") || nonEmptyString(auth, "api_key") {
		return baseCodexLocalProvider(LocalProviderStatusAvailable, LocalProviderAuthModeAPIKey, LocalProviderSourceLocalConfig, "Detected but not enabled. Explicit opt-in is required before use.")
	}
	if hasOAuthToken(auth) {
		return baseCodexLocalProvider(LocalProviderStatusAvailable, LocalProviderAuthModeOAuth, LocalProviderSourceLocalConfig, "Detected OAuth token presence but not enabled. Explicit opt-in is required before use.")
	}
	if authMode, ok := auth["auth_mode"].(string); ok && strings.TrimSpace(authMode) != "" {
		return baseCodexLocalProvider(LocalProviderStatusNeedsLogin, LocalProviderAuthModeOAuth, LocalProviderSourceLocalConfig, "Auth file was found but usable token presence was not detected.")
	}
	return baseCodexLocalProvider(LocalProviderStatusUnavailable, LocalProviderAuthModeUnknown, LocalProviderSourceUnknown, "Not detected.")
}

func codexAuthPath(input LocalProviderDetectionInput) string {
	if strings.TrimSpace(input.CodexHome) != "" {
		return filepath.Join(strings.TrimSpace(input.CodexHome), "auth.json")
	}
	if strings.TrimSpace(input.HomeDir) == "" {
		return ""
	}
	return filepath.Join(strings.TrimSpace(input.HomeDir), ".codex", "auth.json")
}

func baseClaudeLocalProvider(status LocalProviderStatus, authMode LocalProviderAuthMode, source LocalProviderSource, message string) LocalProviderCapability {
	return LocalProviderCapability{
		ProviderID:       "local_claude_code",
		DisplayName:      "Local Claude Code",
		ProviderKind:     LocalProviderKindClaudeCode,
		AuthMode:         authMode,
		Status:           status,
		ModelCandidates:  []string{"claude-sonnet-4-5"},
		Source:           source,
		RedactionApplied: true,
		Message:          message,
	}
}

func baseCodexLocalProvider(status LocalProviderStatus, authMode LocalProviderAuthMode, source LocalProviderSource, message string) LocalProviderCapability {
	return LocalProviderCapability{
		ProviderID:       "local_codex",
		DisplayName:      "Local Codex",
		ProviderKind:     LocalProviderKindCodex,
		AuthMode:         authMode,
		Status:           status,
		ModelCandidates:  []string{"gpt-5"},
		Source:           source,
		RedactionApplied: true,
		Message:          message,
	}
}

func readJSONObject(path string) map[string]any {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil
	}
	return result
}

func nonEmptyString(values map[string]any, key string) bool {
	value, ok := values[key].(string)
	return ok && strings.TrimSpace(value) != ""
}

func hasOAuthToken(values map[string]any) bool {
	if len(values) == 0 {
		return false
	}
	if nonEmptyString(values, "access_token") || nonEmptyString(values, "accessToken") {
		return true
	}
	if tokens, ok := values["tokens"].(map[string]any); ok {
		return nonEmptyString(tokens, "access_token") || nonEmptyString(tokens, "accessToken")
	}
	if oauth, ok := values["oauth"].(map[string]any); ok {
		return nonEmptyString(oauth, "access_token") || nonEmptyString(oauth, "accessToken")
	}
	return false
}
