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
		client = NewClient(cfg.Host)
	}
	report := DoctorReport{OK: true, Config: cfg}
	report.add(okCheck("config", fmt.Sprintf("path=%s found=%v mode=%s provider=%s script=%s", cfg.Path, cfg.Found, cfg.Mode, cfg.Provider, cfg.Script)))
	if err := client.CheckReady(ctx); err != nil {
		report.add(failCheck("api", err.Error(), "start loomi-api or set LOOMI_HOST / loomi config set host"))
		return report
	}
	report.add(okCheck("api", client.BaseURL()))

	providers, err := client.ListModelProviders(ctx)
	if err != nil {
		report.add(failCheck("providers", err.Error(), "check /v1/model-providers and provider settings"))
	} else {
		report.add(providerCheck(ctx, client, cfg, providers))
	}

	tools, err := client.ListTools(ctx)
	if err != nil {
		report.add(failCheck("tools", err.Error(), "check /v1/tools/catalog"))
	} else {
		report.add(toolCatalogCheck(tools))
	}
	return report
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
		return warnCheck("providers", detail, "fix provider configuration or upstream completion before live loomi run")
	}
	if configured == "local_codex" {
		return warnCheck("providers", "default provider local_codex is not registered", "set LOOMI_PROVIDER or run loomi config set provider <id>; use loomi models list to see registered providers")
	}
	return warnCheck("providers", "configured provider not found: "+configured, "run loomi models list or set loomi config provider")
}

func providerReady(provider ProviderCapability) bool {
	if provider.ExecutionState == "unsupported" {
		return false
	}
	return provider.Status == "available" || provider.Status == "configured" || provider.Status == "reachable" || provider.Status == "completion-ok"
}

func providerDetail(provider ProviderCapability) string {
	detail := fmt.Sprintf("%s status=%s execution=%s model=%s", provider.ID, provider.Status, provider.ExecutionState, provider.Model)
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
