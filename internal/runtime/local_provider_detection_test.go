package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalProviderDetectsClaudePrimaryAPIKeyWithoutLeakingKey(t *testing.T) {
	home := t.TempDir()
	writeTestFile(t, filepath.Join(home, ".claude.json"), `{"primaryApiKey":"sk-ant-secret-value"}`)

	providers := DetectLocalProviders(LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}})
	claude := findLocalProvider(t, providers, "local_claude_code")

	if claude.Status != LocalProviderStatusAvailable || claude.AuthMode != LocalProviderAuthModeAPIKey || claude.Source != LocalProviderSourceLocalConfig {
		t.Fatalf("claude = %+v", claude)
	}
	assertNoLocalProviderSecret(t, claude, "sk-ant-secret-value")
}

func TestLocalProviderDetectsClaudeSettingsEnvAsSafeSummary(t *testing.T) {
	home := t.TempDir()
	claudeDir := filepath.Join(home, ".claude")
	writeTestFile(t, filepath.Join(claudeDir, "settings.json"), `{"env":{"ANTHROPIC_AUTH_TOKEN":"Bearer claude-oauth-token","ANTHROPIC_BASE_URL":"https://claude.example.test/v1/private","ANTHROPIC_MODEL":"claude-opus-4-1"}}`)

	providers := DetectLocalProviders(LocalProviderDetectionInput{HomeDir: home, ClaudeConfigDir: claudeDir, Env: map[string]string{}})
	claude := findLocalProvider(t, providers, "local_claude_code")

	if claude.Status != LocalProviderStatusAvailable || claude.AuthMode != LocalProviderAuthModeOAuth || claude.Source != LocalProviderSourceLocalConfig {
		t.Fatalf("claude = %+v", claude)
	}
	if !containsString(claude.ModelCandidates, "claude-opus-4-1") {
		t.Fatalf("model candidates = %+v", claude.ModelCandidates)
	}
	assertNoLocalProviderSecret(t, claude, "Bearer", "claude-oauth-token", "/private")
}

func TestLocalProviderDoesNotExecuteClaudeAPIKeyHelper(t *testing.T) {
	home := t.TempDir()
	writeTestFile(t, filepath.Join(home, ".claude.json"), `{"apiKeyHelper":"touch `+filepath.Join(home, "executed")+`"}`)

	providers := DetectLocalProviders(LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}})
	claude := findLocalProvider(t, providers, "local_claude_code")

	if claude.Status != LocalProviderStatusUnsupported || claude.AuthMode != LocalProviderAuthModeUnknown {
		t.Fatalf("claude = %+v", claude)
	}
	if _, err := os.Stat(filepath.Join(home, "executed")); !os.IsNotExist(err) {
		t.Fatalf("apiKeyHelper appears to have executed")
	}
}

func TestLocalProviderDetectsCodexAuthAPIKeyWithoutLeakingKey(t *testing.T) {
	home := t.TempDir()
	writeTestFile(t, filepath.Join(home, ".codex", "auth.json"), `{"OPENAI_API_KEY":"sk-codex-secret-value"}`)

	providers := DetectLocalProviders(LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{}})
	codex := findLocalProvider(t, providers, "local_codex")

	if codex.Status != LocalProviderStatusAvailable || codex.AuthMode != LocalProviderAuthModeAPIKey || codex.Source != LocalProviderSourceLocalConfig {
		t.Fatalf("codex = %+v", codex)
	}
	assertNoLocalProviderSecret(t, codex, "sk-codex-secret-value")
}

func TestLocalProviderDetectsCodexOAuthTokenWithoutLeakingToken(t *testing.T) {
	home := t.TempDir()
	codexHome := filepath.Join(home, "codex-fixture")
	writeTestFile(t, filepath.Join(codexHome, "auth.json"), `{"auth_mode":"chatgpt","tokens":{"access_token":"access-secret","refresh_token":"refresh-secret"}}`)

	providers := DetectLocalProviders(LocalProviderDetectionInput{HomeDir: home, CodexHome: codexHome, Env: map[string]string{}})
	codex := findLocalProvider(t, providers, "local_codex")

	if codex.Status != LocalProviderStatusAvailable || codex.AuthMode != LocalProviderAuthModeOAuth || codex.Source != LocalProviderSourceLocalConfig {
		t.Fatalf("codex = %+v", codex)
	}
	assertNoLocalProviderSecret(t, codex, "access-secret", "refresh-secret", "access_token", "refresh_token")
}

func TestLocalProviderCodexEnvAPIKeyTakesPrecedenceOverAuthFile(t *testing.T) {
	home := t.TempDir()
	writeTestFile(t, filepath.Join(home, ".codex", "auth.json"), `{"auth_mode":"chatgpt","tokens":{"access_token":"access-secret"}}`)

	providers := DetectLocalProviders(LocalProviderDetectionInput{HomeDir: home, Env: map[string]string{"CODEX_API_KEY": "sk-env-secret"}})
	codex := findLocalProvider(t, providers, "local_codex")

	if codex.Status != LocalProviderStatusAvailable || codex.AuthMode != LocalProviderAuthModeAPIKey || codex.Source != LocalProviderSourceEnv {
		t.Fatalf("codex = %+v", codex)
	}
	assertNoLocalProviderSecret(t, codex, "sk-env-secret", "access-secret")
}

func TestLocalProviderMissingFilesReturnsUnavailableWithTempHome(t *testing.T) {
	home := t.TempDir()

	providers := DetectLocalProviders(LocalProviderDetectionInput{HomeDir: home, CodexHome: filepath.Join(home, "codex"), ClaudeConfigDir: filepath.Join(home, "claude"), Env: map[string]string{}})

	for _, provider := range providers {
		if provider.Status != LocalProviderStatusUnavailable {
			t.Fatalf("provider = %+v", provider)
		}
		body := localProviderBody(provider)
		if strings.Contains(body, os.Getenv("HOME")) {
			t.Fatalf("provider leaked real HOME: %s", body)
		}
	}
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func findLocalProvider(t *testing.T, providers []LocalProviderCapability, providerID string) LocalProviderCapability {
	t.Helper()
	for _, provider := range providers {
		if provider.ProviderID == providerID {
			return provider
		}
	}
	t.Fatalf("provider %s not found in %+v", providerID, providers)
	return LocalProviderCapability{}
}

func assertNoLocalProviderSecret(t *testing.T, provider LocalProviderCapability, values ...string) {
	t.Helper()
	body := localProviderBody(provider)
	for _, value := range values {
		if value != "" && strings.Contains(body, value) {
			t.Fatalf("provider leaked %q in %s", value, body)
		}
	}
	if !provider.RedactionApplied {
		t.Fatalf("redaction_applied = false for %+v", provider)
	}
}

func localProviderBody(provider LocalProviderCapability) string {
	return provider.ProviderID + " " + provider.DisplayName + " " + string(provider.ProviderKind) + " " + string(provider.AuthMode) + " " + string(provider.Status) + " " + strings.Join(provider.ModelCandidates, " ") + " " + string(provider.Source) + " " + provider.Message
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
