package runtime

import (
	"encoding/json"
	"errors"
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

const localCodexDefaultModel = "gpt-5.5"

func LocalProviderRouteCapability(provider LocalProviderCapability) ProviderCapability {
	model := localCodexDefaultModel
	if len(provider.ModelCandidates) > 0 && strings.TrimSpace(provider.ModelCandidates[0]) != "" {
		model = strings.TrimSpace(provider.ModelCandidates[0])
	}
	if provider.ProviderID == "local_codex" && provider.ProviderKind == LocalProviderKindCodex && provider.Status == LocalProviderStatusAvailable {
		return ProviderCapability{
			ID:                  provider.ProviderID,
			Family:              ProviderFamilyOpenAICompatible,
			Model:               model,
			Status:              ProviderStatusAvailable,
			Message:             "Local Codex is enabled for this session.",
			LocalProvider:       true,
			SessionLocal:        true,
			CredentialReference: "redacted",
			ExecutionState:      "supported",
		}
	}
	return ProviderCapability{
		ID:                  provider.ProviderID,
		Family:              ProviderFamilyOpenAICompatible,
		Model:               model,
		Status:              ProviderStatusUnavailable,
		Message:             provider.DisplayName + " is enabled for this session, but execution is unsupported until the local provider execution bridge is implemented.",
		LocalProvider:       true,
		SessionLocal:        true,
		CredentialReference: "redacted",
		ExecutionState:      "unsupported",
	}
}

type LocalCodexCredentialSnapshot struct {
	AuthMode  LocalProviderAuthMode
	BaseURL   string
	APIKey    string
	Model     string
	AccountID string
}

func LoadLocalCodexCredentialSnapshot(input LocalProviderDetectionInput) (LocalCodexCredentialSnapshot, error) {
	env := input.Env
	if env == nil {
		env = map[string]string{}
	}
	model := strings.TrimSpace(env["CODEX_MODEL"])
	if model == "" {
		model = strings.TrimSpace(env["OPENAI_MODEL"])
	}
	if model == "" {
		model = localCodexDefaultModel
	}
	baseURL := firstNonEmpty(env["CODEX_BASE_URL"], env["OPENAI_BASE_URL"], env["LOOMI_LOCAL_CODEX_BASE_URL"])
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if key := firstNonEmpty(env["CODEX_API_KEY"], env["OPENAI_API_KEY"]); key != "" {
		return LocalCodexCredentialSnapshot{AuthMode: LocalProviderAuthModeAPIKey, BaseURL: baseURL, APIKey: key, Model: model}, nil
	}
	auth := readJSONObject(codexAuthPath(input))
	if authModel, ok := auth["model"].(string); ok && strings.TrimSpace(authModel) != "" {
		model = strings.TrimSpace(authModel)
	}
	if authBaseURL, ok := auth["base_url"].(string); ok && strings.TrimSpace(authBaseURL) != "" {
		baseURL = strings.TrimSpace(authBaseURL)
	}
	if key := firstNonEmpty(stringValue(auth, "CODEX_API_KEY"), stringValue(auth, "OPENAI_API_KEY"), stringValue(auth, "api_key")); key != "" {
		return LocalCodexCredentialSnapshot{AuthMode: LocalProviderAuthModeAPIKey, BaseURL: baseURL, APIKey: key, Model: model}, nil
	}
	if token := oauthAccessToken(auth); token != "" {
		oauthBaseURL := strings.TrimSpace(env["LOOMI_LOCAL_CODEX_BASE_URL"])
		if oauthBaseURL == "" {
			if authBaseURL, ok := auth["base_url"].(string); ok && strings.TrimSpace(authBaseURL) != "" {
				oauthBaseURL = strings.TrimSpace(authBaseURL)
			}
		}
		if oauthBaseURL == "" || oauthBaseURL == "https://api.openai.com/v1" {
			oauthBaseURL = "https://chatgpt.com/backend-api/codex"
		}
		return LocalCodexCredentialSnapshot{AuthMode: LocalProviderAuthModeOAuth, BaseURL: oauthBaseURL, APIKey: token, Model: model, AccountID: oauthAccountID(auth)}, nil
	}
	return LocalCodexCredentialSnapshot{}, errors.New("Local Codex login is unavailable.")
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
		ModelCandidates:  []string{localCodexDefaultModel},
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

func stringValue(values map[string]any, key string) string {
	value, ok := values[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func hasOAuthToken(values map[string]any) bool {
	return oauthAccessToken(values) != ""
}

func oauthAccessToken(values map[string]any) string {
	if len(values) == 0 {
		return ""
	}
	if nonEmptyString(values, "access_token") || nonEmptyString(values, "accessToken") {
		return firstNonEmpty(stringValue(values, "access_token"), stringValue(values, "accessToken"))
	}
	if tokens, ok := values["tokens"].(map[string]any); ok {
		if token := firstNonEmpty(stringValue(tokens, "access_token"), stringValue(tokens, "accessToken")); token != "" {
			return token
		}
	}
	if oauth, ok := values["oauth"].(map[string]any); ok {
		if token := firstNonEmpty(stringValue(oauth, "access_token"), stringValue(oauth, "accessToken")); token != "" {
			return token
		}
	}
	return ""
}

func oauthAccountID(values map[string]any) string {
	if len(values) == 0 {
		return ""
	}
	if accountID := firstNonEmpty(stringValue(values, "account_id"), stringValue(values, "accountId")); accountID != "" {
		return accountID
	}
	if tokens, ok := values["tokens"].(map[string]any); ok {
		if accountID := firstNonEmpty(stringValue(tokens, "account_id"), stringValue(tokens, "accountId")); accountID != "" {
			return accountID
		}
	}
	if oauth, ok := values["oauth"].(map[string]any); ok {
		if accountID := firstNonEmpty(stringValue(oauth, "account_id"), stringValue(oauth, "accountId")); accountID != "" {
			return accountID
		}
	}
	return ""
}
