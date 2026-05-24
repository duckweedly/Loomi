package db

import (
	"context"
	"errors"
	"testing"
)

type fakeChecker struct {
	pingErr   error
	schemaErr error
}

func (f fakeChecker) Ping(context.Context) error        { return f.pingErr }
func (f fakeChecker) SchemaReady(context.Context) error { return f.schemaErr }

func TestReadyChecksPass(t *testing.T) {
	checks := CheckReadiness(context.Background(), fakeChecker{})
	if len(checks) != 3 {
		t.Fatalf("len(checks) = %d", len(checks))
	}
	for _, check := range checks {
		if check.Status != StatusOK {
			t.Fatalf("check %s status = %s", check.Name, check.Status)
		}
	}
}

func TestReadyChecksDatabaseFailure(t *testing.T) {
	checks := CheckReadiness(context.Background(), fakeChecker{pingErr: errors.New("dial tcp secret")})
	if checks[1].Name != CheckDatabase || checks[1].Status != StatusFailed {
		t.Fatalf("database check = %+v", checks[1])
	}
	if checks[1].Reason != "database ping failed" {
		t.Fatalf("reason = %q", checks[1].Reason)
	}
}

func TestReadyChecksSchemaFailure(t *testing.T) {
	checks := CheckReadiness(context.Background(), fakeChecker{schemaErr: errors.New("dirty")})
	if checks[2].Name != CheckSchema || checks[2].Status != StatusFailed {
		t.Fatalf("schema check = %+v", checks[2])
	}
	if checks[2].Reason != "schema version unavailable" {
		t.Fatalf("reason = %q", checks[2].Reason)
	}
}

func TestSchemaVersionReady(t *testing.T) {
	if err := schemaVersionReady(5, false); err != nil {
		t.Fatalf("schemaVersionReady(5, false) error = %v", err)
	}
}

func TestSchemaVersionRejectsM5OnlyBaseline(t *testing.T) {
	if err := schemaVersionReady(4, false); err == nil {
		t.Fatal("schemaVersionReady(4, false) error = nil, want error")
	}
}

func TestSchemaVersionRejectsM3OnlyBaseline(t *testing.T) {
	if err := schemaVersionReady(2, false); err == nil {
		t.Fatal("schemaVersionReady(2, false) error = nil, want error")
	}
}

func TestSchemaVersionRejectsM2OnlyBaseline(t *testing.T) {
	if err := schemaVersionReady(1, false); err == nil {
		t.Fatal("schemaVersionReady(1, false) error = nil, want error")
	}
}

func TestSchemaVersionRejectsMissingBaseline(t *testing.T) {
	if err := schemaVersionReady(0, false); err == nil {
		t.Fatal("schemaVersionReady(0, false) error = nil, want error")
	}
}

func TestSchemaVersionRejectsDirtyState(t *testing.T) {
	if err := schemaVersionReady(2, true); err == nil {
		t.Fatal("schemaVersionReady(2, true) error = nil, want error")
	}
}
