package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type DoctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Detail  string `json:"detail,omitempty"`
	Remedy  string `json:"remedy,omitempty"`
	Warning bool   `json:"warning,omitempty"`
}

type DoctorReport struct {
	OK     bool          `json:"ok"`
	Config Config        `json:"config"`
	Checks []DoctorCheck `json:"checks"`
}

func RunDoctor(ctx context.Context, client *Client, cfg Config) DoctorReport {
	if client == nil {
		client = NewClientFromConfig(cfg)
	}
	report := DoctorReport{OK: true, Config: cfg}
	report.add(okCheck("config", fmt.Sprintf("path=%s found=%v api_token_set=%v mode=%s provider=%s script=%s", cfg.Path, cfg.Found, strings.TrimSpace(cfg.APIToken) != "", cfg.Mode, cfg.Provider, cfg.Script)))
	if err := client.CheckReady(ctx); err != nil {
		report.add(failCheck("api", err.Error(), "start loomi-api or set LOOMI_HOST / loomi config set host"))
		return report
	}
	report.add(okCheck("api", client.BaseURL()))

	providers, err := client.ListModelProviders(ctx)
	if err != nil {
		report.add(failCheck("providers", err.Error(), authRemedy(err, "check /v1/model-providers and provider settings")))
	} else {
		report.add(providerCheck(ctx, client, cfg, providers))
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		report.add(failCheck("tools", err.Error(), authRemedy(err, "check /v1/tools/catalog")))
	} else {
		report.add(toolCatalogCheck(tools))
	}
	return report
}

func authRemedy(err error, fallback string) string {
	if err == nil {
		return fallback
	}
	detail := strings.ToLower(err.Error())
	if strings.Contains(detail, "401") && strings.Contains(detail, "missing bearer token") {
		return "set LOOMI_API_TOKEN or run loomi config set api_token <token>, then run loomi doctor again"
	}
	if strings.Contains(detail, "401") || strings.Contains(detail, "403") {
		return "refresh LOOMI_API_TOKEN or run loomi config set api_token <token>, then run loomi doctor again"
	}
	return fallback
}

func (r *DoctorReport) add(check DoctorCheck) {
	r.Checks = append(r.Checks, check)
	if check.Status == "fail" {
		r.OK = false
	}
}

func okCheck(name string, detail string) DoctorCheck {
	return DoctorCheck{Name: name, Status: "ok", Detail: detail}
}

func warnCheck(name string, detail string, remedy string) DoctorCheck {
	return DoctorCheck{Name: name, Status: "warn", Detail: detail, Remedy: remedy, Warning: true}
}

func failCheck(name string, detail string, remedy string) DoctorCheck {
	return DoctorCheck{Name: name, Status: "fail", Detail: detail, Remedy: remedy}
}

func providerCheck(ctx context.Context, client *Client, cfg Config, providers []ProviderCapability) DoctorCheck {
	if strings.TrimSpace(cfg.Script) != "" {
		return okCheck("providers", "script="+cfg.Script)
	}
	configured := strings.TrimSpace(cfg.Provider)
	if configured == "" {
		configured = "local_codex"
	}
	for _, provider := range providers {
		if provider.ID != configured {
			continue
		}
		var err error
		if provider.Status == "configured" || provider.Status == "reachable" {
			provider, err = client.CheckModelProvider(ctx, provider.ID)
		}
		detail := providerDetail(provider)
		if err != nil {
			return warnCheck("providers", detail+" check_error="+err.Error(), "run loomi models list and check provider settings")
		}
		if providerReady(provider) {
			return okCheck("providers", detail)
		}
		return warnCheck("providers", detail, providerCheckRemedy(provider))
	}
	if configured == "local_codex" {
		if local, ok := localProviderDetection(ctx, client, configured); ok {
			return warnCheck("providers", localProviderDetail(local), "enable Local Codex in Settings > Providers or POST /v1/local-provider-detections/local_codex/enable, then run loomi doctor --provider local_codex")
		}
		return warnCheck("providers", "default provider local_codex is not registered", "set LOOMI_PROVIDER or run loomi config set provider <id>; use loomi models list to see registered providers")
	}
	return warnCheck("providers", "configured provider not found: "+configured, "run loomi models list or set loomi config provider")
}

func localProviderDetection(ctx context.Context, client *Client, providerID string) (LocalProviderCapability, bool) {
	localProviders, err := client.ListLocalProviderDetections(ctx)
	if err != nil {
		return LocalProviderCapability{}, false
	}
	for _, provider := range localProviders {
		if provider.ProviderID == providerID {
			return provider, true
		}
	}
	return LocalProviderCapability{}, false
}

func localProviderDetail(provider LocalProviderCapability) string {
	detail := fmt.Sprintf("%s blocked status=%s auth=%s", provider.ProviderID, provider.Status, provider.AuthMode)
	if provider.Message != "" {
		detail += " message=" + provider.Message
	}
	return detail
}

func providerReady(provider ProviderCapability) bool {
	if provider.ExecutionState == "unsupported" {
		return false
	}
	return provider.Status == "available" || provider.Status == "configured" || provider.Status == "reachable" || provider.Status == "completion-ok"
}

func providerDetail(provider ProviderCapability) string {
	detail := fmt.Sprintf("%s status=%s execution=%s model=%s", provider.ID, provider.Status, provider.ExecutionState, provider.Model)
	if provider.CheckStage != "" {
		detail += " check_stage=" + provider.CheckStage
	}
	if provider.CheckCode != "" {
		detail += " check=" + provider.CheckCode
	}
	if provider.HTTPStatus != 0 {
		detail += fmt.Sprintf(" http=%d", provider.HTTPStatus)
	}
	if provider.Message != "" {
		detail += " message=" + provider.Message
	}
	return detail
}

func providerCheckRemedy(provider ProviderCapability) string {
	switch provider.HTTPStatus {
	case 401, 403:
		return "Refresh the provider API token, then run loomi doctor again."
	case 429:
		return "Wait for quota reset or set LOOMI_PROVIDER / loomi config set provider to another configured provider."
	case 503:
		return "Retry later or set LOOMI_PROVIDER / loomi config set provider to another configured provider."
	}
	if provider.CheckStage != "" || provider.CheckCode != "" {
		return "Fix provider configuration or upstream completion before live loomi run."
	}
	return "fix provider configuration or upstream completion before live loomi run"
}

func toolCatalogCheck(tools []ToolCatalogEntry) DoctorCheck {
	if len(tools) == 0 {
		return warnCheck("tools", "catalog is empty", "check tool catalog registration")
	}
	enabled := 0
	groups := map[string]struct{}{}
	for _, tool := range tools {
		if tool.Enabled {
			enabled++
		}
		group := strings.TrimSpace(tool.Group)
		if group == "" {
			group = "other"
		}
		groups[group] = struct{}{}
	}
	detail := fmt.Sprintf("%d tools, %d enabled, %d groups", len(tools), enabled, len(groups))
	if enabled == 0 {
		return warnCheck("tools", detail, "enable Work-mode tools before code-agent dogfood")
	}
	return okCheck("tools", detail)
}

func (r Renderer) PrintDoctor(report DoctorReport) error {
	status := "ok"
	if !report.OK {
		status = "fail"
	}
	if _, err := fmt.Fprintf(r.out(), "doctor %s\n", status); err != nil {
		return err
	}
	for _, check := range report.Checks {
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\n", check.Status, check.Name, check.Detail); err != nil {
			return err
		}
		if check.Remedy != "" {
			if _, err := fmt.Fprintf(r.out(), "fix\t%s\t%s\n", check.Name, check.Remedy); err != nil {
				return err
			}
		}
	}
	return nil
}

func PrintDoctor(w io.Writer, report DoctorReport) error {
	return Renderer{Out: w}.PrintDoctor(report)
}
