package diagnostics

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestNewRequestID(t *testing.T) {
	id := NewRequestID()
	if id == "" {
		t.Fatal("NewRequestID() returned empty id")
	}
	if len(id) < 8 {
		t.Fatalf("NewRequestID() = %q, too short", id)
	}
}

func TestNewOperationID(t *testing.T) {
	id := NewOperationID("startup")
	if id == "" {
		t.Fatal("NewOperationID() returned empty id")
	}
	if id[:8] != "startup_" {
		t.Fatalf("NewOperationID() = %q, want startup_ prefix", id)
	}
}

func TestRedact(t *testing.T) {
	input := "postgres://loomi:secret@127.0.0.1:55433/loomi_m2?sslmode=disable"
	got := Redact(input)
	if got == input {
		t.Fatal("Redact() returned original secret")
	}
	if got != "[redacted]" {
		t.Fatalf("Redact() = %q", got)
	}
}

func TestParseLevel(t *testing.T) {
	tests := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	for input, want := range tests {
		got, err := ParseLevel(input)
		if err != nil {
			t.Fatalf("ParseLevel(%q) error = %v", input, err)
		}
		if got != want {
			t.Fatalf("ParseLevel(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestNewJSONLoggerRespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSONLogger(&buf, slog.LevelWarn)

	logger.Info("hidden", "operation_id", "op_1")
	logger.Warn("visible", "operation_id", "op_2")

	out := buf.String()
	if contains(out, "hidden") {
		t.Fatalf("info log should be filtered: %s", out)
	}
	if !contains(out, "visible") || !contains(out, "operation_id") {
		t.Fatalf("warn log missing structured fields: %s", out)
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
