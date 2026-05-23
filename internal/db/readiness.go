package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckName string
type CheckStatus string

const (
	CheckConfig   CheckName = "config"
	CheckDatabase CheckName = "database"
	CheckSchema   CheckName = "schema"
	StatusOK      CheckStatus = "ok"
	StatusFailed  CheckStatus = "failed"
)

type ReadinessCheck struct {
	Name   CheckName   `json:"name"`
	Status CheckStatus `json:"status"`
	Reason string      `json:"reason,omitempty"`
}

type Checker interface {
	Ping(context.Context) error
	SchemaReady(context.Context) error
}

type PostgresChecker struct {
	Pool *pgxpool.Pool
}

func (c PostgresChecker) Ping(ctx context.Context) error {
	if c.Pool == nil {
		return errors.New("database pool is nil")
	}
	return c.Pool.Ping(ctx)
}

func (c PostgresChecker) SchemaReady(ctx context.Context) error {
	if c.Pool == nil {
		return errors.New("database pool is nil")
	}
	// M3 必须看到 product schema，否则 API 会在运行时撞到缺表。
	var version int
	var dirty bool
	row := c.Pool.QueryRow(ctx, "select version, dirty from schema_migrations order by version desc limit 1")
	if err := row.Scan(&version, &dirty); err != nil {
		return err
	}
	return schemaVersionReady(version, dirty)
}

func schemaVersionReady(version int, dirty bool) error {
	if version < 2 || dirty {
		return errors.New("m3 schema unavailable")
	}
	return nil
}

func CheckReadiness(ctx context.Context, checker Checker) []ReadinessCheck {
	checks := make([]ReadinessCheck, 0, 3)
	checks = append(checks, ReadinessCheck{Name: CheckConfig, Status: StatusOK})
	if err := checker.Ping(ctx); err != nil {
		checks = append(checks, ReadinessCheck{Name: CheckDatabase, Status: StatusFailed, Reason: "database ping failed"})
	} else {
		checks = append(checks, ReadinessCheck{Name: CheckDatabase, Status: StatusOK})
	}
	if err := checker.SchemaReady(ctx); err != nil {
		checks = append(checks, ReadinessCheck{Name: CheckSchema, Status: StatusFailed, Reason: "schema version unavailable"})
	} else {
		checks = append(checks, ReadinessCheck{Name: CheckSchema, Status: StatusOK})
	}
	return checks
}
