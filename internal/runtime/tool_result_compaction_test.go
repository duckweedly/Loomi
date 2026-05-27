package runtime

import (
	"strings"
	"testing"
)

func TestCompactToolResultTextKeepsSignal(t *testing.T) {
	lines := []string{}
	for i := 0; i < 200; i++ {
		lines = append(lines, "repeated progress line")
	}
	lines = append(lines,
		"path: web/src/App.tsx",
		"status: 503",
		"error: provider unavailable",
	)

	compact := compactToolResultText(strings.Join(lines, "\n"), 700)

	if len(compact) > 700 {
		t.Fatalf("compact length = %d", len(compact))
	}
	for _, want := range []string{"web/src/App.tsx", "503", "provider unavailable", "tool output compacted"} {
		if !strings.Contains(compact, want) {
			t.Fatalf("compact missing %q: %s", want, compact)
		}
	}
}

func TestCompactToolResultTextLeavesSmallContentUntouched(t *testing.T) {
	input := `{"ok":true,"path":"web/src/App.tsx"}`

	if got := compactToolResultText(input, 700); got != input {
		t.Fatalf("compact = %q", got)
	}
}

func TestCompactToolResultPayloadCompactsNestedStrings(t *testing.T) {
	input := map[string]any{
		"nested": map[string]any{
			"output": strings.Repeat("noise\n", 200) + "path: web/src/App.tsx\nstatus: 503\n",
		},
	}

	got := compactToolResultPayload(input, 700)
	output, ok := got["nested"].(map[string]any)["output"].(string)
	if !ok {
		t.Fatalf("output = %+v", got)
	}
	if !strings.Contains(output, "web/src/App.tsx") || !strings.Contains(output, "503") {
		t.Fatalf("output = %q", output)
	}
}

func TestCompactToolResultPayloadPreservesReadableSummaryAfterRedaction(t *testing.T) {
	input := map[string]any{
		"operation": "exec_command",
		"summary":   "Tests passed after reading README and go.mod.",
		"output":    "stdout: ok\n普通总结: 项目使用 Go 和 Bun。\nsecret: sk-live-test",
	}

	got := compactToolResultPayload(input, 700)

	if got["summary"] != "Tests passed after reading README and go.mod." {
		t.Fatalf("summary was not preserved: %+v", got)
	}
	output, ok := got["output"].(string)
	if !ok {
		t.Fatalf("output = %+v", got)
	}
	if output == "[redacted]" || !strings.Contains(output, "普通总结") {
		t.Fatalf("readable output was over-redacted: %q", output)
	}
	if strings.Contains(output, "sk-live-test") {
		t.Fatalf("secret leaked in output: %q", output)
	}
}
