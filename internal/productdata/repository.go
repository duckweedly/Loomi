package productdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sheridiany/loomi/internal/identity"
)

type PostgresRepository struct {
	Pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{Pool: pool}
}

type modelProviderConfigScanner interface {
	Scan(dest ...any) error
}

type webSearchConfigScanner interface {
	Scan(dest ...any) error
}

type workspaceRootConfigScanner interface {
	Scan(dest ...any) error
}

type memoryProviderConfigScanner interface {
	Scan(dest ...any) error
}

type mcpServerConfigScanner interface {
	Scan(dest ...any) error
}

type sandboxProcessScanner interface {
	Scan(dest ...any) error
}

func scanModelProviderConfig(row modelProviderConfigScanner) (ModelProviderConfig, error) {
	var provider ModelProviderConfig
	err := row.Scan(&provider.ID, &provider.UserID, &provider.Family, &provider.BaseURL, &provider.APIKey, &provider.Model, &provider.Enabled)
	return provider, err
}

func scanWebSearchConfig(row webSearchConfigScanner) (WebSearchConfig, error) {
	var config WebSearchConfig
	err := row.Scan(&config.UserID, &config.TavilyAPIKey, &config.BraveAPIKey)
	return config, err
}

func scanWorkspaceRootConfig(row workspaceRootConfigScanner) (WorkspaceRootConfig, error) {
	var config WorkspaceRootConfig
	err := row.Scan(&config.UserID, &config.Path)
	config.DisplayName = WorkspaceDisplayNameFromPath(config.Path)
	return config, err
}

func scanMemoryProviderConfig(row memoryProviderConfigScanner) (MemoryProviderConfig, error) {
	var config MemoryProviderConfig
	err := row.Scan(
		&config.UserID,
		&config.Enabled,
		&config.Provider,
		&config.CommitAfterRun,
		&config.SemanticEndpoint,
		&config.OpenViking.BaseURL,
		&config.OpenViking.RootAPIKey,
		&config.OpenViking.EmbeddingSelector,
		&config.OpenViking.EmbeddingProvider,
		&config.OpenViking.EmbeddingModel,
		&config.OpenViking.EmbeddingAPIKey,
		&config.OpenViking.EmbeddingAPIBase,
		&config.OpenViking.EmbeddingDimension,
		&config.OpenViking.VLMSelector,
		&config.OpenViking.VLMProvider,
		&config.OpenViking.VLMModel,
		&config.OpenViking.VLMAPIKey,
		&config.OpenViking.VLMAPIBase,
		&config.OpenViking.RerankSelector,
		&config.OpenViking.RerankProvider,
		&config.OpenViking.RerankModel,
		&config.OpenViking.RerankAPIKey,
		&config.OpenViking.RerankAPIBase,
		&config.Nowledge.BaseURL,
		&config.Nowledge.APIKey,
		&config.Nowledge.RequestTimeoutMS,
		&config.Diagnostic,
		&config.UpdatedAt,
	)
	config.OpenViking.RootAPIKeySet = config.OpenViking.RootAPIKey != ""
	config.OpenViking.EmbeddingAPIKeySet = config.OpenViking.EmbeddingAPIKey != ""
	config.OpenViking.VLMAPIKeySet = config.OpenViking.VLMAPIKey != ""
	config.OpenViking.RerankAPIKeySet = config.OpenViking.RerankAPIKey != ""
	config.Nowledge.APIKeySet = config.Nowledge.APIKey != ""
	return config, err
}

func scanMCPServerConfig(row mcpServerConfigScanner) (MCPServerConfigRecord, error) {
	var record MCPServerConfigRecord
	var argsRaw []byte
	var envRaw []byte
	if err := row.Scan(&record.UserID, &record.Slug, &record.DisplayName, &record.Enabled, &record.Transport, &record.Command, &argsRaw, &envRaw, &record.TimeoutMS); err != nil {
		return MCPServerConfigRecord{}, err
	}
	_ = json.Unmarshal(argsRaw, &record.Args)
	_ = json.Unmarshal(envRaw, &record.Env)
	if record.Args == nil {
		record.Args = []string{}
	}
	if record.Env == nil {
		record.Env = map[string]string{}
	}
	return record, nil
}

func scanSandboxProcess(row sandboxProcessScanner) (SandboxProcessRecord, error) {
	var record SandboxProcessRecord
	var argvRaw []byte
	var endedAt pgtype.Timestamptz
	var exitCode pgtype.Int4
	if err := row.Scan(
		&record.RunID,
		&record.ProcessID,
		&argvRaw,
		&record.CwdAlias,
		&record.Status,
		&record.Cursor,
		&record.StdoutTail,
		&record.StdoutCursor,
		&record.StderrTail,
		&record.StderrCursor,
		&record.StdoutBytes,
		&record.StderrBytes,
		&record.StdinOpen,
		&record.InputSeq,
		&record.TimedOut,
		&record.StartedAt,
		&record.UpdatedAt,
		&endedAt,
		&exitCode,
		&record.TerminalSummary,
		&record.OutputLimit,
	); err != nil {
		return SandboxProcessRecord{}, err
	}
	_ = json.Unmarshal(argvRaw, &record.ArgvSummary)
	if record.ArgvSummary == nil {
		record.ArgvSummary = []string{}
	}
	if endedAt.Valid {
		value := endedAt.Time
		record.EndedAt = &value
	}
	if exitCode.Valid {
		value := int(exitCode.Int32)
		record.ExitCode = &value
	}
	return normalizeSandboxProcessRecord(record), nil
}

func (r *PostgresRepository) CurrentIdentity(ctx context.Context, ident identity.LocalIdentity) (User, error) {
	return r.ensureUser(ctx, ident)
}

func (r *PostgresRepository) SaveModelProviderConfig(ctx context.Context, ident identity.LocalIdentity, input ModelProviderConfig) (ModelProviderConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ModelProviderConfig{}, err
	}
	provider := normalizeModelProviderConfig(input)
	if provider.ID == "" || provider.Model == "" || provider.APIKey == "" {
		return ModelProviderConfig{}, NewError(CodeProviderMisconfigured, "Provider configuration is incomplete.")
	}
	provider.UserID = user.ID
	row := r.Pool.QueryRow(ctx, `insert into model_provider_configs (id, user_id, family, base_url, api_key, model, enabled) values ($1, $2, $3, $4, $5, $6, $7) on conflict (user_id, id) do update set family=excluded.family, base_url=excluded.base_url, api_key=excluded.api_key, model=excluded.model, enabled=excluded.enabled, updated_at=now() returning id, user_id, family, base_url, api_key, model, enabled`, provider.ID, provider.UserID, provider.Family, provider.BaseURL, provider.APIKey, provider.Model, provider.Enabled)
	return scanModelProviderConfig(row)
}

func (r *PostgresRepository) ListModelProviderConfigs(ctx context.Context, ident identity.LocalIdentity) ([]ModelProviderConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select id, user_id, family, base_url, api_key, model, enabled from model_provider_configs where user_id=$1 order by id asc`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	providers := []ModelProviderConfig{}
	for rows.Next() {
		provider, err := scanModelProviderConfig(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}
	return providers, rows.Err()
}

func (r *PostgresRepository) SaveWebSearchConfig(ctx context.Context, ident identity.LocalIdentity, input WebSearchConfig) (WebSearchConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return WebSearchConfig{}, err
	}
	next := normalizeWebSearchConfig(input)
	row := r.Pool.QueryRow(ctx, `insert into web_search_configs (user_id, tavily_api_key, brave_api_key) values ($1, $2, $3) on conflict (user_id) do update set tavily_api_key=case when excluded.tavily_api_key<>'' then excluded.tavily_api_key else web_search_configs.tavily_api_key end, brave_api_key=case when excluded.brave_api_key<>'' then excluded.brave_api_key else web_search_configs.brave_api_key end, updated_at=now() returning user_id, tavily_api_key, brave_api_key`, user.ID, next.TavilyAPIKey, next.BraveAPIKey)
	return scanWebSearchConfig(row)
}

func (r *PostgresRepository) GetWebSearchConfig(ctx context.Context, ident identity.LocalIdentity) (WebSearchConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return WebSearchConfig{}, err
	}
	config, err := scanWebSearchConfig(r.Pool.QueryRow(ctx, `select user_id, tavily_api_key, brave_api_key from web_search_configs where user_id=$1`, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return WebSearchConfig{UserID: user.ID}, nil
	}
	return config, err
}

func (r *PostgresRepository) SaveWorkspaceRootConfig(ctx context.Context, ident identity.LocalIdentity, input WorkspaceRootConfig) (WorkspaceRootConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return WorkspaceRootConfig{}, err
	}
	next := normalizeWorkspaceRootConfig(input)
	if next.Path == "" {
		return WorkspaceRootConfig{}, NewError(CodeInvalidRequest, "Workspace folder is required.")
	}
	row := r.Pool.QueryRow(ctx, `insert into workspace_root_configs (user_id, root_path) values ($1, $2) on conflict (user_id) do update set root_path=excluded.root_path, updated_at=now() returning user_id, root_path`, user.ID, next.Path)
	return scanWorkspaceRootConfig(row)
}

func (r *PostgresRepository) GetWorkspaceRootConfig(ctx context.Context, ident identity.LocalIdentity) (WorkspaceRootConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return WorkspaceRootConfig{}, err
	}
	config, err := scanWorkspaceRootConfig(r.Pool.QueryRow(ctx, `select user_id, root_path from workspace_root_configs where user_id=$1`, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkspaceRootConfig{UserID: user.ID}, nil
	}
	return config, err
}

func (r *PostgresRepository) SaveMemoryProviderConfig(ctx context.Context, ident identity.LocalIdentity, input MemoryProviderConfig) (MemoryProviderConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryProviderConfig{}, err
	}
	next := normalizeMemoryProviderConfig(input, time.Now())
	row := r.Pool.QueryRow(ctx, `insert into memory_provider_configs (user_id, enabled, provider, commit_after_run, semantic_endpoint, openviking_base_url, openviking_root_api_key, openviking_embedding_selector, openviking_embedding_provider, openviking_embedding_model, openviking_embedding_api_key, openviking_embedding_api_base, openviking_embedding_dimension, openviking_vlm_selector, openviking_vlm_provider, openviking_vlm_model, openviking_vlm_api_key, openviking_vlm_api_base, openviking_rerank_selector, openviking_rerank_provider, openviking_rerank_model, openviking_rerank_api_key, openviking_rerank_api_base, nowledge_base_url, nowledge_api_key, nowledge_request_timeout_ms, diagnostic) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27) on conflict (user_id) do update set enabled=excluded.enabled, provider=excluded.provider, commit_after_run=excluded.commit_after_run, semantic_endpoint=excluded.semantic_endpoint, openviking_base_url=excluded.openviking_base_url, openviking_root_api_key=case when excluded.openviking_root_api_key<>'' then excluded.openviking_root_api_key else memory_provider_configs.openviking_root_api_key end, openviking_embedding_selector=excluded.openviking_embedding_selector, openviking_embedding_provider=excluded.openviking_embedding_provider, openviking_embedding_model=excluded.openviking_embedding_model, openviking_embedding_api_key=case when excluded.openviking_embedding_api_key<>'' then excluded.openviking_embedding_api_key else memory_provider_configs.openviking_embedding_api_key end, openviking_embedding_api_base=excluded.openviking_embedding_api_base, openviking_embedding_dimension=excluded.openviking_embedding_dimension, openviking_vlm_selector=excluded.openviking_vlm_selector, openviking_vlm_provider=excluded.openviking_vlm_provider, openviking_vlm_model=excluded.openviking_vlm_model, openviking_vlm_api_key=case when excluded.openviking_vlm_api_key<>'' then excluded.openviking_vlm_api_key else memory_provider_configs.openviking_vlm_api_key end, openviking_vlm_api_base=excluded.openviking_vlm_api_base, openviking_rerank_selector=excluded.openviking_rerank_selector, openviking_rerank_provider=excluded.openviking_rerank_provider, openviking_rerank_model=excluded.openviking_rerank_model, openviking_rerank_api_key=case when excluded.openviking_rerank_api_key<>'' then excluded.openviking_rerank_api_key else memory_provider_configs.openviking_rerank_api_key end, openviking_rerank_api_base=excluded.openviking_rerank_api_base, nowledge_base_url=excluded.nowledge_base_url, nowledge_api_key=case when excluded.nowledge_api_key<>'' then excluded.nowledge_api_key else memory_provider_configs.nowledge_api_key end, nowledge_request_timeout_ms=excluded.nowledge_request_timeout_ms, diagnostic=excluded.diagnostic, updated_at=now() returning user_id, enabled, provider, commit_after_run, coalesce(semantic_endpoint,''), coalesce(openviking_base_url,''), coalesce(openviking_root_api_key,''), coalesce(openviking_embedding_selector,''), coalesce(openviking_embedding_provider,''), coalesce(openviking_embedding_model,''), coalesce(openviking_embedding_api_key,''), coalesce(openviking_embedding_api_base,''), coalesce(openviking_embedding_dimension,0), coalesce(openviking_vlm_selector,''), coalesce(openviking_vlm_provider,''), coalesce(openviking_vlm_model,''), coalesce(openviking_vlm_api_key,''), coalesce(openviking_vlm_api_base,''), coalesce(openviking_rerank_selector,''), coalesce(openviking_rerank_provider,''), coalesce(openviking_rerank_model,''), coalesce(openviking_rerank_api_key,''), coalesce(openviking_rerank_api_base,''), coalesce(nowledge_base_url,''), coalesce(nowledge_api_key,''), coalesce(nowledge_request_timeout_ms,0), diagnostic, updated_at`, user.ID, next.Enabled, next.Provider, next.CommitAfterRun, next.SemanticEndpoint, next.OpenViking.BaseURL, next.OpenViking.RootAPIKey, next.OpenViking.EmbeddingSelector, next.OpenViking.EmbeddingProvider, next.OpenViking.EmbeddingModel, next.OpenViking.EmbeddingAPIKey, next.OpenViking.EmbeddingAPIBase, next.OpenViking.EmbeddingDimension, next.OpenViking.VLMSelector, next.OpenViking.VLMProvider, next.OpenViking.VLMModel, next.OpenViking.VLMAPIKey, next.OpenViking.VLMAPIBase, next.OpenViking.RerankSelector, next.OpenViking.RerankProvider, next.OpenViking.RerankModel, next.OpenViking.RerankAPIKey, next.OpenViking.RerankAPIBase, next.Nowledge.BaseURL, next.Nowledge.APIKey, next.Nowledge.RequestTimeoutMS, next.Diagnostic)
	return scanMemoryProviderConfig(row)
}

func (r *PostgresRepository) GetMemoryProviderStatus(ctx context.Context, ident identity.LocalIdentity) (MemoryProviderStatus, error) {
	config, err := r.GetMemoryProviderConfig(ctx, ident)
	if err != nil {
		return MemoryProviderStatus{}, err
	}
	return memoryProviderStatus(config, time.Now()), nil
}

func (r *PostgresRepository) GetMemoryProviderConfig(ctx context.Context, ident identity.LocalIdentity) (MemoryProviderConfig, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryProviderConfig{}, err
	}
	config, err := scanMemoryProviderConfig(r.Pool.QueryRow(ctx, `select user_id, enabled, provider, commit_after_run, coalesce(semantic_endpoint,''), coalesce(openviking_base_url,''), coalesce(openviking_root_api_key,''), coalesce(openviking_embedding_selector,''), coalesce(openviking_embedding_provider,''), coalesce(openviking_embedding_model,''), coalesce(openviking_embedding_api_key,''), coalesce(openviking_embedding_api_base,''), coalesce(openviking_embedding_dimension,0), coalesce(openviking_vlm_selector,''), coalesce(openviking_vlm_provider,''), coalesce(openviking_vlm_model,''), coalesce(openviking_vlm_api_key,''), coalesce(openviking_vlm_api_base,''), coalesce(openviking_rerank_selector,''), coalesce(openviking_rerank_provider,''), coalesce(openviking_rerank_model,''), coalesce(openviking_rerank_api_key,''), coalesce(openviking_rerank_api_base,''), coalesce(nowledge_base_url,''), coalesce(nowledge_api_key,''), coalesce(nowledge_request_timeout_ms,0), coalesce(diagnostic,''), updated_at from memory_provider_configs where user_id=$1`, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		config = defaultMemoryProviderConfig(user.ID, time.Now())
		err = nil
	}
	if err != nil {
		return MemoryProviderConfig{}, err
	}
	return config, nil
}

func (r *PostgresRepository) workspaceRootPathForUserTx(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	var root string
	err := tx.QueryRow(ctx, `select root_path from workspace_root_configs where user_id=$1`, userID).Scan(&root)
	if errors.Is(err, pgx.ErrNoRows) {
		return strings.TrimSpace(os.Getenv("LOOMI_WORKSPACE_ROOT")), nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(root), nil
}

func (r *PostgresRepository) workspaceRootPathForRunTx(ctx context.Context, tx pgx.Tx, userID string, runID string) (string, error) {
	rows, err := tx.Query(ctx, `select metadata from background_jobs where user_id=$1 and run_id=$2 order by created_at asc, id asc`, userID, runID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return "", err
		}
		metadata := map[string]any{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &metadata)
		}
		if root := metadataStringValue(metadata, "workspace_root_path"); root != "" {
			return root, nil
		}
	}
	return "", rows.Err()
}

func (r *PostgresRepository) ListMemoryProviderErrors(ctx context.Context, ident identity.LocalIdentity, limit int) ([]MemoryProviderErrorEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	items := []MemoryProviderErrorEvent{}
	config, err := r.GetMemoryProviderConfig(ctx, ident)
	if err != nil {
		return nil, err
	}
	status := memoryProviderStatus(config, time.Now())
	if status.Diagnostic.Code != "" && status.Diagnostic.Code != "ok" {
		checkedAt := time.Now()
		if status.CheckedAt != nil {
			checkedAt = *status.CheckedAt
		}
		items = append(items, MemoryProviderErrorEvent{Code: status.Diagnostic.Code, Message: status.Diagnostic.Message, Provider: status.Provider, State: status.State, CheckedAt: checkedAt})
	}
	rows, err := r.Pool.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where user_id=$1 and type in ($2,$3) order by created_at desc, id desc limit $4`, user.ID, EventMemoryExternalSnapshotFailed, "memory_provider_commit_failed", limitMemoryProviderErrorQueryLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, memoryProviderErrorFromRunEvent(event))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CheckedAt.After(items[j].CheckedAt)
	})
	return limitMemoryProviderErrors(items, limit), nil
}

func (r *PostgresRepository) GetMemoryOverviewSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryOverviewSnapshot, error) {
	return r.memoryOverviewSnapshot(ctx, ident, false)
}

func (r *PostgresRepository) RebuildMemoryOverviewSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryOverviewSnapshot, error) {
	return r.memoryOverviewSnapshot(ctx, ident, true)
}

func (r *PostgresRepository) GetMemoryImpressionSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryImpressionSnapshot, error) {
	return r.memoryImpressionSnapshot(ctx, ident, false)
}

func (r *PostgresRepository) RebuildMemoryImpressionSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryImpressionSnapshot, error) {
	return r.memoryImpressionSnapshot(ctx, ident, true)
}

func (r *PostgresRepository) memoryOverviewSnapshot(ctx context.Context, ident identity.LocalIdentity, rebuilt bool) (MemoryOverviewSnapshot, error) {
	output, err := r.SearchMemory(ctx, ident, MemorySearchInput{Limit: 7, Purpose: "snapshot"})
	if err != nil {
		return MemoryOverviewSnapshot{}, err
	}
	return buildMemoryOverviewSnapshot(semanticMemorySnapshotItems(output.Items), time.Now(), rebuilt), nil
}

func (r *PostgresRepository) memoryImpressionSnapshot(ctx context.Context, ident identity.LocalIdentity, rebuilt bool) (MemoryImpressionSnapshot, error) {
	output, err := r.SearchMemory(ctx, ident, MemorySearchInput{Limit: 7, Purpose: "impression"})
	if err != nil {
		return MemoryImpressionSnapshot{}, err
	}
	return buildMemoryImpressionSnapshot(semanticMemorySnapshotItems(output.Items), time.Now(), rebuilt), nil
}

func (r *PostgresRepository) SaveMCPServerConfig(ctx context.Context, ident identity.LocalIdentity, input MCPServerConfigRecord) (MCPServerConfigRecord, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MCPServerConfigRecord{}, err
	}
	record := normalizeMCPServerConfigRecord(input)
	if record.Slug == "" {
		return MCPServerConfigRecord{}, NewError(CodeInvalidRequest, "MCP server slug is required.")
	}
	record.UserID = user.ID
	argsRaw, err := json.Marshal(record.Args)
	if err != nil {
		return MCPServerConfigRecord{}, err
	}
	envRaw, err := json.Marshal(record.Env)
	if err != nil {
		return MCPServerConfigRecord{}, err
	}
	row := r.Pool.QueryRow(ctx, `insert into mcp_server_configs (user_id, slug, display_name, enabled, transport, command, args_json, env_json, timeout_ms) values ($1,$2,$3,$4,$5,$6,$7,$8,$9) on conflict (user_id, slug) do update set display_name=excluded.display_name, enabled=excluded.enabled, transport=excluded.transport, command=excluded.command, args_json=excluded.args_json, env_json=excluded.env_json, timeout_ms=excluded.timeout_ms, updated_at=now() returning user_id, slug, display_name, enabled, transport, command, args_json, env_json, timeout_ms`, record.UserID, record.Slug, record.DisplayName, record.Enabled, record.Transport, record.Command, argsRaw, envRaw, record.TimeoutMS)
	return scanMCPServerConfig(row)
}

func (r *PostgresRepository) ListMCPServerConfigs(ctx context.Context, ident identity.LocalIdentity) ([]MCPServerConfigRecord, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select user_id, slug, display_name, enabled, transport, command, args_json, env_json, timeout_ms from mcp_server_configs where user_id=$1 order by slug asc`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := []MCPServerConfigRecord{}
	for rows.Next() {
		record, err := scanMCPServerConfig(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (r *PostgresRepository) DeleteMCPServerConfig(ctx context.Context, ident identity.LocalIdentity, slug string) error {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return err
	}
	_, err = r.Pool.Exec(ctx, `delete from mcp_server_configs where user_id=$1 and slug=$2`, user.ID, strings.TrimSpace(slug))
	return err
}

func (r *PostgresRepository) SaveSandboxProcess(ctx context.Context, input SandboxProcessRecord) error {
	record := normalizeSandboxProcessRecord(input)
	if record.RunID == "" || record.ProcessID == "" {
		return NewError(CodeInvalidRequest, "Sandbox process record is incomplete.")
	}
	argvRaw, err := json.Marshal(record.ArgvSummary)
	if err != nil {
		return err
	}
	_, err = r.Pool.Exec(ctx, `insert into sandbox_process_records (
		run_id, process_id, argv_summary, cwd_alias, status, cursor, stdout_tail, stdout_cursor,
		stderr_tail, stderr_cursor, stdout_bytes, stderr_bytes, stdin_open, input_seq, timed_out,
		started_at, updated_at, ended_at, exit_code, terminal_summary, output_limit
	) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
	on conflict (process_id) do update set
		run_id=excluded.run_id,
		argv_summary=excluded.argv_summary,
		cwd_alias=excluded.cwd_alias,
		status=excluded.status,
		cursor=excluded.cursor,
		stdout_tail=excluded.stdout_tail,
		stdout_cursor=excluded.stdout_cursor,
		stderr_tail=excluded.stderr_tail,
		stderr_cursor=excluded.stderr_cursor,
		stdout_bytes=excluded.stdout_bytes,
		stderr_bytes=excluded.stderr_bytes,
		stdin_open=excluded.stdin_open,
		input_seq=excluded.input_seq,
		timed_out=excluded.timed_out,
		started_at=excluded.started_at,
		updated_at=excluded.updated_at,
		ended_at=excluded.ended_at,
		exit_code=excluded.exit_code,
		terminal_summary=excluded.terminal_summary,
		output_limit=excluded.output_limit`,
		record.RunID,
		record.ProcessID,
		argvRaw,
		record.CwdAlias,
		record.Status,
		record.Cursor,
		record.StdoutTail,
		record.StdoutCursor,
		record.StderrTail,
		record.StderrCursor,
		record.StdoutBytes,
		record.StderrBytes,
		record.StdinOpen,
		record.InputSeq,
		record.TimedOut,
		record.StartedAt,
		record.UpdatedAt,
		record.EndedAt,
		record.ExitCode,
		record.TerminalSummary,
		record.OutputLimit,
	)
	return err
}

func (r *PostgresRepository) ListSandboxProcesses(ctx context.Context) ([]SandboxProcessRecord, error) {
	rows, err := r.Pool.Query(ctx, `select run_id, process_id, argv_summary, cwd_alias, status, cursor, stdout_tail, stdout_cursor, stderr_tail, stderr_cursor, stdout_bytes, stderr_bytes, stdin_open, input_seq, timed_out, started_at, updated_at, ended_at, exit_code, terminal_summary, output_limit from sandbox_process_records order by updated_at asc, process_id asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records := []SandboxProcessRecord{}
	for rows.Next() {
		record, err := scanSandboxProcess(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (r *PostgresRepository) DeleteSandboxProcessesUpdatedBefore(ctx context.Context, before time.Time) (int, error) {
	tag, err := r.Pool.Exec(ctx, `delete from sandbox_process_records where updated_at < $1`, before)
	return int(tag.RowsAffected()), err
}

func (r *PostgresRepository) CreateThread(ctx context.Context, ident identity.LocalIdentity, input CreateThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Thread{}, err
	}
	threadID := NewThreadID()
	personaID := strings.TrimSpace(input.PersonaID)
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Thread{}, err
	}
	defer tx.Rollback(ctx)
	if err := validatePersonaReferenceTx(ctx, tx, personaID); err != nil {
		return Thread{}, err
	}
	row := tx.QueryRow(ctx, `insert into threads (id, user_id, title, mode, lifecycle_status, persona_id) values ($1, $2, $3, $4, $5, nullif($6, '')) returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, threadID, user.ID, title, input.Mode, ThreadLifecycleActive, personaID)
	thread, err := scanThread(row)
	if err != nil {
		return Thread{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func (r *PostgresRepository) UpsertSeedThread(ctx context.Context, ident identity.LocalIdentity, input SeedThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Thread{}, err
	}
	row := r.Pool.QueryRow(ctx, `insert into threads (id, user_id, title, mode, lifecycle_status) values ($1, $2, $3, $4, 'active') on conflict (id) do update set title=excluded.title, mode=excluded.mode, lifecycle_status='active', archived_at=null, updated_at=now() returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, input.ID, user.ID, title, input.Mode)
	return scanThread(row)
}

func (r *PostgresRepository) ListThreads(ctx context.Context, ident identity.LocalIdentity, includeArchived bool) ([]Thread, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	query := `select id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at from threads where user_id=$1`
	args := []any{user.ID}
	if !includeArchived {
		query += ` and lifecycle_status='active'`
	}
	query += ` order by updated_at desc, id desc`
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var threads []Thread
	for rows.Next() {
		thread, err := scanThread(rows)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	return threads, rows.Err()
}

func (r *PostgresRepository) GetThread(ctx context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Thread{}, err
	}
	row := r.Pool.QueryRow(ctx, `select id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at from threads where id=$1 and user_id=$2`, threadID, user.ID)
	thread, err := scanThread(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	return thread, err
}

func (r *PostgresRepository) UpdateThread(ctx context.Context, ident identity.LocalIdentity, threadID string, input UpdateThreadInput) (Thread, error) {
	current, err := r.GetThread(ctx, ident, threadID)
	if err != nil {
		return Thread{}, err
	}
	title := current.Title
	mode := current.Mode
	personaID := current.PersonaID
	if input.Title != nil {
		normalized, err := NormalizeThreadTitle(*input.Title)
		if err != nil {
			return Thread{}, err
		}
		title = normalized
	}
	if input.Mode != nil {
		if err := ValidateThreadMode(*input.Mode); err != nil {
			return Thread{}, err
		}
		mode = *input.Mode
	}
	if input.PersonaID != nil {
		personaID = strings.TrimSpace(*input.PersonaID)
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Thread{}, err
	}
	defer tx.Rollback(ctx)
	if err := validatePersonaReferenceTx(ctx, tx, personaID); err != nil {
		return Thread{}, err
	}
	row := tx.QueryRow(ctx, `update threads set title=$1, mode=$2, persona_id=nullif($5, ''), updated_at=now() where id=$3 and user_id=$4 returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, title, mode, threadID, current.UserID, personaID)
	thread, err := scanThread(row)
	if err != nil {
		return Thread{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func (r *PostgresRepository) ArchiveThread(ctx context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	current, err := r.GetThread(ctx, ident, threadID)
	if err != nil {
		return Thread{}, err
	}
	row := r.Pool.QueryRow(ctx, `update threads set lifecycle_status='archived', archived_at=coalesce(archived_at, now()), updated_at=now() where id=$1 and user_id=$2 returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, threadID, current.UserID)
	return scanThread(row)
}

func (r *PostgresRepository) CreateContextSource(ctx context.Context, ident identity.LocalIdentity, input CreateContextSourceInput) (ContextSource, error) {
	normalized, err := NormalizeContextSourceInput(input)
	if err != nil {
		return ContextSource{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ContextSource{}, err
	}
	if _, err := r.GetThread(ctx, ident, normalized.ThreadID); err != nil {
		return ContextSource{}, err
	}
	row := r.Pool.QueryRow(ctx, `insert into context_sources (id, user_id, thread_id, kind, title, locator, summary, status, metadata) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id, thread_id, user_id, kind, title, locator, summary, status, metadata, created_at, updated_at`,
		NewContextSourceID(),
		user.ID,
		normalized.ThreadID,
		normalized.Kind,
		normalized.Title,
		normalized.Locator,
		normalized.Summary,
		ContextSourceStatusRegistered,
		mustJSON(normalized.Metadata),
	)
	return scanContextSource(row)
}

func (r *PostgresRepository) ListContextSources(ctx context.Context, ident identity.LocalIdentity, input ListContextSourcesInput) ([]ContextSource, error) {
	threadID := strings.TrimSpace(input.ThreadID)
	if threadID == "" {
		return nil, NewError(CodeInvalidRequest, "Thread id is required.")
	}
	if _, err := r.GetThread(ctx, ident, threadID); err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 || limit > 50 {
		limit = 50
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select id, thread_id, user_id, kind, title, locator, summary, status, metadata, created_at, updated_at from context_sources where user_id=$1 and thread_id=$2 order by created_at asc, id asc limit $3`, user.ID, threadID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sources := []ContextSource{}
	for rows.Next() {
		source, err := scanContextSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, ident identity.LocalIdentity, threadID string, input CreateMessageInput) (Message, bool, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, false, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, false, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Message{}, false, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Message{}, false, err
	}
	defer tx.Rollback(ctx)
	var threadUserID string
	if err := tx.QueryRow(ctx, `select user_id from threads where id=$1 and user_id=$2`, threadID, user.ID).Scan(&threadUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Message{}, false, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Message{}, false, err
	}
	if clientMessageID != nil {
		message, err := scanMessage(tx.QueryRow(ctx, `select id, thread_id, user_id, role, content, metadata, client_message_id, created_at from messages where thread_id=$1 and user_id=$2 and client_message_id=$3`, threadID, user.ID, *clientMessageID))
		if err == nil {
			return message, false, tx.Commit(ctx)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Message{}, false, err
		}
	}
	messageID := NewMessageID()
	message, err := scanMessage(tx.QueryRow(ctx, `insert into messages (id, thread_id, user_id, role, content, metadata, client_message_id) values ($1, $2, $3, 'user', $4, '{}'::jsonb, $5) returning id, thread_id, user_id, role, content, metadata, client_message_id, created_at`, messageID, threadID, user.ID, content, clientMessageID))
	if err != nil {
		return Message{}, false, err
	}
	if _, err := tx.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, threadID, user.ID); err != nil {
		return Message{}, false, err
	}
	return message, true, tx.Commit(ctx)
}

func (r *PostgresRepository) AppendAssistantMessage(ctx context.Context, ident identity.LocalIdentity, threadID string, input AppendAssistantMessageInput) (Message, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Message{}, err
	}
	metadata := RedactEventMetadata(input.Metadata)
	if runID, ok := metadata["run_id"].(string); ok && runID != "" {
		var existing string
		err := r.Pool.QueryRow(ctx, `select id from messages where thread_id=$1 and user_id=$2 and role='assistant' and metadata->>'run_id'=$3 limit 1`, threadID, user.ID, runID).Scan(&existing)
		if err == nil {
			return Message{}, NewError(CodeInvalidRequest, "Assistant message already exists for run.")
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return Message{}, err
		}
	}
	message, err := scanMessage(r.Pool.QueryRow(ctx, `insert into messages (id, thread_id, user_id, role, content, metadata, client_message_id) values ($1, $2, $3, 'assistant', $4, $5, null) returning id, thread_id, user_id, role, content, metadata, client_message_id, created_at`, NewMessageID(), threadID, user.ID, content, mustJSON(metadata)))
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			return Message{}, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Message{}, err
	}
	if _, err := r.Pool.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, threadID, user.ID); err != nil {
		return Message{}, err
	}
	return message, nil
}

func (r *PostgresRepository) UpsertSeedMessage(ctx context.Context, ident identity.LocalIdentity, input SeedMessageInput) (Message, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Message{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Message{}, err
	}
	defer tx.Rollback(ctx)
	var threadUserID string
	if err := tx.QueryRow(ctx, `select user_id from threads where id=$1 and user_id=$2`, input.ThreadID, user.ID).Scan(&threadUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Message{}, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Message{}, err
	}
	message, err := scanMessage(tx.QueryRow(ctx, `insert into messages (id, thread_id, user_id, role, content, metadata, client_message_id) values ($1, $2, $3, 'user', $4, '{}'::jsonb, $5) on conflict (id) do update set content=excluded.content, client_message_id=excluded.client_message_id returning id, thread_id, user_id, role, content, metadata, client_message_id, created_at`, input.ID, input.ThreadID, user.ID, content, clientMessageID))
	if err != nil {
		return Message{}, err
	}
	if _, err := tx.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, input.ThreadID, user.ID); err != nil {
		return Message{}, err
	}
	return message, tx.Commit(ctx)
}

func (r *PostgresRepository) ListMessages(ctx context.Context, ident identity.LocalIdentity, threadID string) ([]Message, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	var exists bool
	if err := r.Pool.QueryRow(ctx, `select exists(select 1 from threads where id=$1 and user_id=$2)`, threadID, user.ID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, NewError(CodeThreadNotFound, "Thread not found.")
	}
	rows, err := r.Pool.Query(ctx, `select id, thread_id, user_id, role, content, metadata, client_message_id, created_at from messages where thread_id=$1 and user_id=$2 order by created_at asc, id asc`, threadID, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *PostgresRepository) StartRun(ctx context.Context, ident identity.LocalIdentity, threadID string, input StartRunInput) (Run, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Run{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Run{}, err
	}
	defer tx.Rollback(ctx)
	run, err := r.startRunTx(ctx, tx, user, threadID, input)
	if err != nil {
		return Run{}, err
	}
	return run, tx.Commit(ctx)
}

func (r *PostgresRepository) startRunTx(ctx context.Context, tx pgx.Tx, user User, threadID string, input StartRunInput) (Run, error) {
	var threadUserID, threadPersonaID string
	if err := tx.QueryRow(ctx, `select user_id, coalesce(persona_id, '') from threads where id=$1 and user_id=$2 and lifecycle_status='active'`, threadID, user.ID).Scan(&threadUserID, &threadPersonaID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Run{}, err
	}
	source, err := NormalizeRunSource(input.Source)
	if err != nil {
		return Run{}, err
	}
	snapshot, err := r.resolvePersonaSnapshotTx(ctx, tx, threadPersonaID, input.PersonaID)
	if err != nil {
		return Run{}, err
	}
	runID := NewRunID()
	run, err := scanRun(tx.QueryRow(ctx, `insert into runs (id, thread_id, user_id, status, source, title, persona_id) values ($1, $2, $3, 'queued', $4, $5, nullif($6, '')) returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, runID, threadID, user.ID, source, TitleForRunSource(source), snapshot.ID))
	if err != nil {
		if strings.Contains(err.Error(), "runs_one_active_per_thread_idx") {
			return Run{}, NewError(CodeActiveRunExists, "Thread already has an active run.")
		}
		return Run{}, err
	}
	run.PersonaID = snapshot.ID
	jobID := NewBackgroundJobID()
	metadata := map[string]any{"source": string(source), "job_id": jobID}
	workspaceRoot, err := r.workspaceRootPathForUserTx(ctx, tx, user.ID)
	if err != nil {
		return Run{}, err
	}
	if workspaceRoot != "" {
		metadata["workspace_root_path"] = workspaceRoot
		metadata["workspace_label"] = WorkspaceDisplayNameFromPath(workspaceRoot)
	}
	if source == RunSourceLocalSimulated {
		metadata["script_name"] = NormalizeScriptName(input.ScriptName)
	} else {
		metadata["message_id"] = input.MessageID
		metadata["provider_id"] = runProviderID(input.ProviderID, snapshot)
		metadata["model"] = runModel(input.ProviderID, input.Model, snapshot)
	}
	if snapshot.ID != "" {
		metadata["persona_id"] = snapshot.ID
		metadata["persona_version"] = snapshot.Version
		metadata["persona_name"] = snapshot.Name
		metadata["persona_resolved_from"] = string(snapshot.ResolvedFrom)
		if err := insertPersonaSnapshot(ctx, tx, run.ID, snapshot); err != nil {
			return Run{}, err
		}
	}
	created, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, metadata) values ($1, $2, $3, $4, 1, 'lifecycle', 'run_created', 'Run created', $5) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, threadID, user.ID, mustJSON(RedactEventMetadata(metadata))))
	if err != nil {
		return Run{}, err
	}
	queued, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, metadata) values ($1, $2, $3, $4, 2, 'lifecycle', 'run_queued', 'Run queued', $5) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, threadID, user.ID, mustJSON(RedactEventMetadata(map[string]any{"job_id": jobID}))))
	if err != nil {
		return Run{}, err
	}
	if err := updateRunStepStateProjectionTx(ctx, tx, created); err != nil {
		return Run{}, err
	}
	if err := updateRunStepStateProjectionTx(ctx, tx, queued); err != nil {
		return Run{}, err
	}
	if _, err := tx.Exec(ctx, `insert into background_jobs (id, run_id, thread_id, user_id, kind, status, max_attempts, metadata) values ($1, $2, $3, $4, 'run_execution', 'queued', 3, $5)`, jobID, run.ID, threadID, user.ID, mustJSON(metadata)); err != nil {
		return Run{}, err
	}
	if _, err := tx.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, threadID, user.ID); err != nil {
		return Run{}, err
	}
	return run, nil
}

func (r *PostgresRepository) GetRun(ctx context.Context, ident identity.LocalIdentity, runID string) (Run, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Run{}, err
	}
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, NewError(CodeRunNotFound, "Run not found.")
	}
	return run, err
}

func (r *PostgresRepository) GetCurrentRun(ctx context.Context, ident identity.LocalIdentity, threadID string) (Run, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Run{}, err
	}
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where thread_id=$1 and user_id=$2 order by case when status in ('pending','queued','running','recovering','blocked_on_tool_approval','stopping','retrying') then 0 else 1 end, updated_at desc, id desc limit 1`, threadID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, NewError(CodeRunNotFound, "Run not found.")
	}
	return run, err
}

func (r *PostgresRepository) ListRunEvents(ctx context.Context, ident identity.LocalIdentity, runID string, afterSequence int) ([]RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	var exists bool
	if err := r.Pool.QueryRow(ctx, `select exists(select 1 from runs where id=$1 and user_id=$2)`, runID, user.ID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, NewError(CodeRunNotFound, "Run not found.")
	}
	rows, err := r.Pool.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where run_id=$1 and user_id=$2 and sequence>$3 order by sequence asc, id asc`, runID, user.ID, afterSequence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []RunEvent
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *PostgresRepository) HasRunEventType(ctx context.Context, ident identity.LocalIdentity, runID string, eventType string) (bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return false, err
	}
	var exists bool
	if err := r.Pool.QueryRow(ctx, `select exists(select 1 from run_events where run_id=$1 and user_id=$2 and type=$3)`, runID, user.ID, strings.TrimSpace(eventType)).Scan(&exists); err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	var runExists bool
	if err := r.Pool.QueryRow(ctx, `select exists(select 1 from runs where id=$1 and user_id=$2)`, runID, user.ID).Scan(&runExists); err != nil {
		return false, err
	}
	if !runExists {
		return false, NewError(CodeRunNotFound, "Run not found.")
	}
	return false, nil
}

func (r *PostgresRepository) PrepareRunContext(ctx context.Context, ident identity.LocalIdentity, job BackgroundJob) (RunContext, error) {
	run, err := r.GetRun(ctx, ident, job.RunID)
	if err != nil {
		return RunContext{}, err
	}
	thread, err := r.GetThread(ctx, ident, run.ThreadID)
	if err != nil {
		return RunContext{}, err
	}
	if job.ID == "" || job.RunID != run.ID || job.ThreadID != thread.ID || job.UserID != run.UserID {
		return RunContext{}, NewError(CodeInvalidRequest, "Run context job boundary is invalid.")
	}
	messages, err := r.ListMessages(ctx, ident, thread.ID)
	if err != nil {
		return RunContext{}, err
	}
	state, stateErr := r.GetRunStepState(ctx, ident, run.ID)
	useStateContext := stateErr == nil && runStepStateCanPrepareContext(run, state)
	if !useStateContext {
		if stateErr != nil {
			return RunContext{}, stateErr
		}
		return RunContext{}, NewError(CodeInvalidRequest, "Run context state is incomplete.")
	}
	context, err := buildRunContextFromState(run, thread, messages, job, state)
	if err != nil {
		return RunContext{}, err
	}
	snapshot, err := r.getPersonaSnapshot(ctx, run.ID)
	if err == nil {
		context.Persona = snapshot
		applyPersonaToRunContextFromState(&context, state)
	}
	memories, err := r.SearchMemory(ctx, ident, MemorySearchInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, Limit: 5, Purpose: "run_context"})
	if err != nil {
		context.MemorySnapshot = MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Limit: 5, LoadStatus: "unavailable"}
		return context, nil
	}
	status := "loaded"
	if len(memories.Items) == 0 {
		status = "empty"
	}
	context.MemorySnapshot = MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Entries: memories.Items, Limit: 5, TotalCandidates: len(memories.Items), LoadStatus: status, RedactionApplied: true}
	notebook, err := r.SearchMemory(ctx, ident, MemorySearchInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, SourceType: "notebook", Limit: 5, Purpose: "run_context_notebook"})
	if err != nil {
		context.NotebookSnapshot = MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Limit: 5, LoadStatus: "unavailable"}
	} else {
		notebookStatus := "loaded"
		if len(notebook.Items) == 0 {
			notebookStatus = "empty"
		}
		context.NotebookSnapshot = MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Entries: notebook.Items, Limit: 5, TotalCandidates: len(notebook.Items), LoadStatus: notebookStatus, RedactionApplied: true}
	}
	sources, err := r.ListContextSources(ctx, ident, ListContextSourcesInput{ThreadID: thread.ID, Limit: 10})
	if err == nil {
		context.ContextSources = sources
		_, _ = r.AppendRunEvent(ctx, ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: EventContextSourcesLoaded, Summary: "Context sources loaded", Metadata: contextSourcesEventMetadata(sources)})
	}
	_, _ = r.AppendRunEvent(ctx, ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: EventMemorySnapshotLoaded, Summary: "Memory snapshot loaded", Metadata: memorySnapshotEventMetadata(context.MemorySnapshot)})
	return context, nil
}

func (r *PostgresRepository) ListToolCatalog(ctx context.Context, ident identity.LocalIdentity) ([]ToolCatalogEntry, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where user_id=$1 and type in ('mcp_discovery_succeeded','mcp_discovery_failed','mcp_discovery_rejected') order by created_at asc, id asc`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []RunEvent
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return SafeToolCatalogFromEvents(events), nil
}

func (r *PostgresRepository) ListMCPDiscoveryEvents(ctx context.Context, ident identity.LocalIdentity) ([]RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where user_id=$1 and type in ('mcp_discovery_succeeded','mcp_discovery_failed','mcp_discovery_rejected') order by created_at asc, id asc`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []RunEvent
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *PostgresRepository) CreateArtifact(ctx context.Context, ident identity.LocalIdentity, input CreateArtifactInput) (Artifact, error) {
	title := strings.TrimSpace(input.Title)
	content := input.Content
	if title == "" || strings.TrimSpace(content) == "" {
		return Artifact{}, NewError(CodeInvalidRequest, "Artifact title and content are required.")
	}
	if !utf8.ValidString(content) {
		return Artifact{}, NewError(CodeInvalidRequest, "Artifact content must be valid UTF-8.")
	}
	limit := boundedArtifactBytes(input.MaxBytes)
	if len([]byte(content)) > limit {
		return Artifact{}, NewError(CodeInvalidRequest, "Artifact content is too large.")
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Artifact{}, err
	}
	thread, err := scanThread(r.Pool.QueryRow(ctx, `select id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at from threads where id=$1 and user_id=$2`, strings.TrimSpace(input.ThreadID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Artifact{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if err != nil {
		return Artifact{}, err
	}
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 and thread_id=$3`, strings.TrimSpace(input.RunID), user.ID, thread.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Artifact{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return Artifact{}, err
	}
	return scanArtifact(r.Pool.QueryRow(ctx, `insert into artifacts (id, user_id, thread_id, run_id, title, artifact_type, content, content_bytes, text_excerpt, truncated) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) returning id, thread_id, run_id, title, artifact_type, content, content_bytes, text_excerpt, truncated, created_at, updated_at`, NewArtifactID(), user.ID, thread.ID, run.ID, title, normalizedArtifactType(input.ArtifactType), content, len([]byte(content)), artifactExcerpt(content, limit), false))
}

func (r *PostgresRepository) ReadArtifact(ctx context.Context, ident identity.LocalIdentity, input ReadArtifactInput) (Artifact, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Artifact{}, err
	}
	artifact, err := scanArtifact(r.Pool.QueryRow(ctx, `select id, thread_id, run_id, title, artifact_type, content, content_bytes, text_excerpt, truncated, created_at, updated_at from artifacts where id=$1 and user_id=$2 and thread_id=$3`, strings.TrimSpace(input.ArtifactID), user.ID, strings.TrimSpace(input.ThreadID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return Artifact{}, NewError(CodeArtifactNotFound, "Artifact not found.")
	}
	if err != nil {
		return Artifact{}, err
	}
	limit := boundedArtifactBytes(input.MaxBytes)
	artifact.TextExcerpt = artifactExcerpt(artifact.Content, limit)
	artifact.Truncated = len([]byte(artifact.Content)) > limit
	return artifact, nil
}

func (r *PostgresRepository) ListArtifacts(ctx context.Context, ident identity.LocalIdentity, input ListArtifactsInput) ([]Artifact, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.Pool.Query(ctx, `select a.id, a.thread_id, a.run_id, a.title, a.artifact_type, '' as content, a.content_bytes, a.text_excerpt, a.truncated, a.created_at, a.updated_at from artifacts a join threads t on t.id=a.thread_id and t.user_id=a.user_id where a.user_id=$1 and a.thread_id=$2 order by a.created_at asc, a.id asc limit $3`, user.ID, strings.TrimSpace(input.ThreadID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	artifacts := []Artifact{}
	for rows.Next() {
		artifact, err := scanArtifact(rows)
		if err != nil {
			return nil, err
		}
		artifact.TextExcerpt = artifactExcerpt(artifact.TextExcerpt, 512)
		artifacts = append(artifacts, artifact)
	}
	return artifacts, rows.Err()
}

func (r *PostgresRepository) SpawnAgentTask(ctx context.Context, ident identity.LocalIdentity, input SpawnAgentTaskInput) (AgentTask, error) {
	role := strings.TrimSpace(input.Role)
	goal := strings.TrimSpace(input.Goal)
	if !isSupportedAgentRole(role) || goal == "" || len([]rune(goal)) > 4000 {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task role and goal are invalid.")
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return AgentTask{}, err
	}
	thread, err := scanThread(r.Pool.QueryRow(ctx, `select id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at from threads where id=$1 and user_id=$2`, strings.TrimSpace(input.ThreadID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 and thread_id=$3`, strings.TrimSpace(input.RunID), user.ID, thread.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	return scanAgentTask(r.Pool.QueryRow(ctx, `insert into agent_tasks (id, user_id, thread_id, run_id, role, goal, status) values ($1,$2,$3,$4,$5,$6,$7) returning id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at`, NewAgentTaskID(), user.ID, thread.ID, run.ID, role, goal, AgentTaskStatusSpawned))
}

func (r *PostgresRepository) ListAgentTasks(ctx context.Context, ident identity.LocalIdentity, input ListAgentTasksInput) ([]AgentTask, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.Pool.Query(ctx, `select at.id, at.thread_id, at.run_id, at.role, at.goal, at.status, at.result_summary, coalesce(at.child_thread_id, ''), coalesce(at.child_run_id, ''), coalesce(at.parent_tool_call_id, ''), at.delegated_at, at.created_at, at.updated_at from agent_tasks at join threads t on t.id=at.thread_id and t.user_id=at.user_id where at.user_id=$1 and at.thread_id=$2 order by at.created_at asc, at.id asc limit $3`, user.ID, strings.TrimSpace(input.ThreadID), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []AgentTask{}
	for rows.Next() {
		task, err := scanAgentTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (r *PostgresRepository) StartAgentTask(ctx context.Context, ident identity.LocalIdentity, input StartAgentTaskInput) (AgentTask, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return AgentTask{}, err
	}
	task, err := scanAgentTask(r.Pool.QueryRow(ctx, `update agent_tasks set status=$1, updated_at=now() where id=$2 and user_id=$3 and thread_id=$4 and status=$5 returning id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at`, AgentTaskStatusInProgress, strings.TrimSpace(input.TaskID), user.ID, strings.TrimSpace(input.ThreadID), AgentTaskStatusSpawned))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task not found or not startable.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	return task, nil
}

func (r *PostgresRepository) DelegateAgentTask(ctx context.Context, ident identity.LocalIdentity, input DelegateAgentTaskInput) (AgentTask, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return AgentTask{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return AgentTask{}, err
	}
	defer tx.Rollback(ctx)
	parentThread, err := scanThread(tx.QueryRow(ctx, `select id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at from threads where id=$1 and user_id=$2 and mode='work' and lifecycle_status='active'`, strings.TrimSpace(input.ThreadID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	task, err := scanAgentTask(tx.QueryRow(ctx, `select id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at from agent_tasks where id=$1 and user_id=$2 and thread_id=$3 for update`, strings.TrimSpace(input.TaskID), user.ID, parentThread.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task not found.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	if task.Status != AgentTaskStatusSpawned && task.Status != AgentTaskStatusInProgress {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task is already terminal.")
	}
	parentToolCallID := strings.TrimSpace(input.ParentToolCallID)
	if task.ChildRunID != "" || task.ChildThreadID != "" {
		if parentToolCallID != "" && parentToolCallID == strings.TrimSpace(task.ParentToolCallID) {
			return task, tx.Commit(ctx)
		}
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task is already delegated.")
	}
	parentRun, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 and thread_id=$3`, task.RunID, user.ID, parentThread.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	route, err := r.providerRouteForRunTx(ctx, tx, parentRun.ID)
	if err != nil {
		return AgentTask{}, err
	}
	if parentRun.Source == RunSourceModelGateway && route.ProviderID == "" {
		return AgentTask{}, NewError(CodeInvalidRequest, "Parent run provider route is unavailable.")
	}
	childThread, err := scanThread(tx.QueryRow(ctx, `insert into threads (id, user_id, title, mode, lifecycle_status, persona_id) values ($1, $2, $3, 'work', 'active', nullif($4, '')) returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, NewThreadID(), user.ID, agentChildThreadTitle(task), parentThread.PersonaID))
	if err != nil {
		return AgentTask{}, err
	}
	childMessage, err := scanMessage(tx.QueryRow(ctx, `insert into messages (id, thread_id, user_id, role, content, metadata, client_message_id) values ($1, $2, $3, 'user', $4, $5, null) returning id, thread_id, user_id, role, content, metadata, client_message_id, created_at`, NewMessageID(), childThread.ID, user.ID, agentChildPrompt(parentThread, task), mustJSON(map[string]any{"parent_thread_id": parentThread.ID, "parent_run_id": parentRun.ID, "parent_agent_task_id": task.ID})))
	if err != nil {
		return AgentTask{}, err
	}
	childRun, err := r.startRunTx(ctx, tx, user, childThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: childMessage.ID, ProviderID: route.ProviderID, Model: route.Model, PersonaID: parentThread.PersonaID})
	if err != nil {
		return AgentTask{}, err
	}
	task, err = scanAgentTask(tx.QueryRow(ctx, `update agent_tasks set status=$1, child_thread_id=$2, child_run_id=$3, parent_tool_call_id=$4, delegated_at=now(), updated_at=now() where id=$5 and user_id=$6 and thread_id=$7 returning id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at`, AgentTaskStatusInProgress, childThread.ID, childRun.ID, parentToolCallID, task.ID, user.ID, parentThread.ID))
	if err != nil {
		return AgentTask{}, err
	}
	return task, tx.Commit(ctx)
}

func (r *PostgresRepository) ReconcileAgentTaskChildRuns(ctx context.Context, ident identity.LocalIdentity, limit int) ([]AgentTaskChildRunReconciliation, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `select id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at from agent_tasks where user_id=$1 and status=$2 and child_run_id is not null and parent_tool_call_id is not null order by updated_at asc, id asc for update skip locked limit $3`, user.ID, AgentTaskStatusInProgress, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []AgentTask{}
	for rows.Next() {
		task, err := scanAgentTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	reconciled := []AgentTaskChildRunReconciliation{}
	for _, task := range tasks {
		childRun, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2`, task.ChildRunID, user.ID))
		if err != nil || !IsRunTerminal(childRun.Status) {
			continue
		}
		parentRun, call, err := scopedPostgresToolCall(ctx, tx, user.ID, task.ThreadID, task.RunID, task.ParentToolCallID)
		if err != nil || IsRunTerminal(parentRun.Status) || call.ExecutionStatus != ToolCallExecutionExecuting {
			continue
		}
		resultText, err := postgresChildRunResultSummary(ctx, tx, childRun)
		if err != nil {
			return nil, err
		}
		taskStatus := AgentTaskStatusFailed
		if childRun.Status == RunStatusCompleted {
			taskStatus = AgentTaskStatusCompleted
		}
		task, err = scanAgentTask(tx.QueryRow(ctx, `update agent_tasks set status=$1, result_summary=$2, updated_at=now() where id=$3 and user_id=$4 returning id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at`, taskStatus, RedactEventText(resultText), task.ID, user.ID))
		if err != nil {
			return nil, err
		}
		result := agentTaskResultSummary(ToolNameAgentDelegate, task, childRun, resultText)
		call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='succeeded', result_summary=$1, updated_at=now() where id=$2 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, mustJSON(result), call.ID))
		if err != nil {
			return nil, err
		}
		queuedRun, err := scanRun(tx.QueryRow(ctx, `update runs set status='queued', completed_at=null, updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, parentRun.ID, user.ID))
		if err != nil {
			return nil, err
		}
		toolMetadata, err := toolCallEventMetadataForPostgresState(ctx, tx, queuedRun, call)
		if err != nil {
			return nil, err
		}
		succeeded, err := insertRunEvent(ctx, tx, queuedRun, RunEventCategoryProgress, EventToolCallSucceeded, "Delegated agent child run completed", nil, toolMetadata)
		if err != nil {
			return nil, err
		}
		jobID := NewBackgroundJobID()
		metadata := map[string]any{"source": string(queuedRun.Source), "job_id": jobID, "tool_call_id": call.ToolCallID, "resume_reason": "agent_child_run_completed", "child_run_id": childRun.ID, "agent_task_id": task.ID}
		workspaceRoot, err := r.workspaceRootPathForRunTx(ctx, tx, user.ID, queuedRun.ID)
		if err != nil {
			return nil, err
		}
		if workspaceRoot != "" {
			metadata["workspace_root_path"] = workspaceRoot
			metadata["workspace_label"] = WorkspaceDisplayNameFromPath(workspaceRoot)
		}
		queued, err := insertRunEvent(ctx, tx, queuedRun, RunEventCategoryProgress, EventRunQueued, "Run queued", nil, metadata)
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `insert into background_jobs (id, run_id, thread_id, user_id, kind, status, priority, max_attempts, scheduled_at, metadata) values ($1, $2, $3, $4, $5, 'queued', 50, 3, now(), $6)`, jobID, queuedRun.ID, queuedRun.ThreadID, user.ID, BackgroundJobKindRunExecution, mustJSON(metadata)); err != nil {
			return nil, err
		}
		reconciled = append(reconciled, AgentTaskChildRunReconciliation{Task: task, Run: queuedRun, Events: []RunEvent{succeeded, queued}})
	}
	return reconciled, tx.Commit(ctx)
}

func postgresChildRunResultSummary(ctx context.Context, tx pgx.Tx, childRun Run) (string, error) {
	var content string
	err := tx.QueryRow(ctx, `select content from messages where thread_id=$1 and user_id=$2 and role='assistant' order by created_at desc, id desc limit 1`, childRun.ThreadID, childRun.UserID).Scan(&content)
	if errors.Is(err, pgx.ErrNoRows) {
		return "Child run " + string(childRun.Status) + ".", nil
	}
	if err != nil {
		return "", err
	}
	content = RedactEventText(strings.TrimSpace(content))
	if len([]rune(content)) > 1000 {
		content = string([]rune(content)[:1000])
	}
	if content == "" {
		content = "Child run " + string(childRun.Status) + "."
	}
	return content, nil
}

func (r *PostgresRepository) CompleteAgentTask(ctx context.Context, ident identity.LocalIdentity, input CompleteAgentTaskInput) (AgentTask, error) {
	summary := strings.TrimSpace(input.ResultSummary)
	if summary == "" || len([]rune(summary)) > 4000 {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent result summary is invalid.")
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return AgentTask{}, err
	}
	task, err := scanAgentTask(r.Pool.QueryRow(ctx, `update agent_tasks set status=$1, result_summary=$2, updated_at=now() where id=$3 and user_id=$4 and thread_id=$5 and status in ($6,$7) returning id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at`, AgentTaskStatusCompleted, RedactEventText(summary), strings.TrimSpace(input.TaskID), user.ID, strings.TrimSpace(input.ThreadID), AgentTaskStatusSpawned, AgentTaskStatusInProgress))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task not found or already terminal.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	return task, nil
}

func (r *PostgresRepository) FailAgentTask(ctx context.Context, ident identity.LocalIdentity, input FailAgentTaskInput) (AgentTask, error) {
	summary := strings.TrimSpace(input.ResultSummary)
	if summary == "" || len([]rune(summary)) > 4000 {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent result summary is invalid.")
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return AgentTask{}, err
	}
	task, err := scanAgentTask(r.Pool.QueryRow(ctx, `update agent_tasks set status=$1, result_summary=$2, updated_at=now() where id=$3 and user_id=$4 and thread_id=$5 and status in ($6,$7) returning id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at`, AgentTaskStatusFailed, RedactEventText(summary), strings.TrimSpace(input.TaskID), user.ID, strings.TrimSpace(input.ThreadID), AgentTaskStatusSpawned, AgentTaskStatusInProgress))
	if errors.Is(err, pgx.ErrNoRows) {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task not found or already terminal.")
	}
	if err != nil {
		return AgentTask{}, err
	}
	return task, nil
}

func (r *PostgresRepository) CreateMemoryEntry(ctx context.Context, ident identity.LocalIdentity, input CreateMemoryEntryInput) (MemoryEntry, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryEntry{}, err
	}
	scopeType, scopeID, err := normalizeMemoryScope(user.ID, input.ScopeType, input.ScopeID)
	if err != nil {
		return MemoryEntry{}, err
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Content)
	if err != nil {
		return MemoryEntry{}, err
	}
	status := MemoryEntryApproved
	if safety == MemorySafetyBlocked {
		status = MemoryEntryDisabled
	}
	row := r.Pool.QueryRow(ctx, `insert into memory_entries (id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, source_thread_id, source_run_id, source_event_id, content_hash) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,nullif($10,''),nullif($11,''),nullif($12,''),$13) returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'')`,
		NewMemoryEntryID(), user.ID, scopeType, scopeID, title, summary, content, status, safety, strings.TrimSpace(input.SourceThreadID), strings.TrimSpace(input.SourceRunID), strings.TrimSpace(input.SourceEventID), memoryContentHash(scopeType, scopeID, content))
	return scanMemoryEntry(row)
}

func (r *PostgresRepository) ListMemoryEntries(ctx context.Context, ident identity.LocalIdentity, input MemorySearchInput) (MemorySearchOutput, error) {
	return r.SearchMemory(ctx, ident, input)
}

func (r *PostgresRepository) SearchMemory(ctx context.Context, ident identity.LocalIdentity, input MemorySearchInput) (MemorySearchOutput, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemorySearchOutput{}, err
	}
	limit := memoryLimit(input.Limit)
	scopeType := input.ScopeType
	scopeID := strings.TrimSpace(input.ScopeID)
	if scopeType == "" {
		scopeType = MemoryScopeUser
	}
	query := strings.ToLower(strings.TrimSpace(input.Query))
	sql := `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where user_id=$1 and safety_state <> 'blocked'`
	args := []any{user.ID}
	if input.IncludeTombstoned {
		sql += ` and status in ('approved','tombstoned')`
	} else {
		sql += ` and status='approved'`
	}
	if scopeType == MemoryScopeThread {
		args = append(args, scopeID)
		sql += ` and ((scope_type='user' and scope_id=$1) or (scope_type='thread' and scope_id=$2))`
	} else {
		sql += ` and scope_type='user' and scope_id=$1`
	}
	if query != "" {
		args = append(args, "%"+query+"%")
		sql += ` and (lower(title) like $` + intPlaceholder(len(args)) + ` or lower(summary) like $` + intPlaceholder(len(args)) + ` or lower(content) like $` + intPlaceholder(len(args)) + `)`
	}
	if sourceRunID := strings.TrimSpace(input.SourceRunID); sourceRunID != "" {
		args = append(args, sourceRunID)
		sql += ` and source_run_id=$` + intPlaceholder(len(args))
	}
	if sourceThreadID := strings.TrimSpace(input.SourceThreadID); sourceThreadID != "" {
		args = append(args, sourceThreadID)
		sql += ` and source_thread_id=$` + intPlaceholder(len(args))
	}
	switch strings.TrimSpace(input.SourceType) {
	case "", "any":
	case "run":
		sql += ` and source_run_id is not null`
	case "thread":
		sql += ` and source_thread_id is not null`
	case "notebook":
		sql += ` and source_event_id='notebook'`
	case "manual":
		sql += ` and source_run_id is null and source_thread_id is null and coalesce(source_event_id,'') <> 'notebook'`
	default:
		return MemorySearchOutput{}, NewError(CodeInvalidRequest, "Memory source type is invalid.")
	}
	args = append(args, limit)
	sql += ` order by updated_at desc, id desc limit $` + intPlaceholder(len(args))
	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return MemorySearchOutput{}, err
	}
	defer rows.Close()
	var items []MemorySearchResult
	for rows.Next() {
		entry, err := scanMemoryEntry(rows)
		if err != nil {
			return MemorySearchOutput{}, err
		}
		items = append(items, memorySearchResult(entry))
	}
	return MemorySearchOutput{Items: items}, rows.Err()
}

func (r *PostgresRepository) GetMemoryEntry(ctx context.Context, ident identity.LocalIdentity, entryID string, input MemoryEntryAccessInput) (MemoryEntry, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryEntry{}, err
	}
	entry, err := scanMemoryEntry(r.Pool.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where id=$1 and user_id=$2 and status in ('approved','tombstoned') and safety_state <> 'blocked'`, strings.TrimSpace(entryID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryEntry{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if err != nil {
		return MemoryEntry{}, err
	}
	if !memoryEntryReadableTo(entry, user.ID, input) {
		return MemoryEntry{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	entry.Content = ""
	return entry, err
}

func (r *PostgresRepository) ListMemoryAudit(ctx context.Context, ident identity.LocalIdentity, input MemoryAuditInput) (MemoryAuditOutput, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryAuditOutput{}, err
	}
	limit := memoryLimit(input.Limit)
	sql := `select id, run_id, thread_id, user_id, 0, 'progress', type, summary, null, metadata, created_at from memory_audit_events where user_id=$1 and type = any($2)`
	args := []any{user.ID, []string{EventMemorySnapshotLoaded, EventMemoryWriteProposed, EventMemoryWriteApproved, EventMemoryWriteDenied, EventMemoryEntryDeleted}}
	if threadID := strings.TrimSpace(input.ThreadID); threadID != "" {
		args = append(args, threadID)
		sql += ` and thread_id=$` + intPlaceholder(len(args))
	}
	if runID := strings.TrimSpace(input.SourceRunID); runID != "" {
		args = append(args, runID)
		sql += ` and run_id=$` + intPlaceholder(len(args))
	}
	if eventType := strings.TrimSpace(input.EventType); eventType != "" {
		if eventType == "memory_deleted" {
			eventType = EventMemoryEntryDeleted
		}
		args = append(args, eventType)
		sql += ` and type=$` + intPlaceholder(len(args))
	}
	args = append(args, limit)
	sql += ` order by created_at desc, id desc limit $` + intPlaceholder(len(args))
	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return MemoryAuditOutput{}, err
	}
	defer rows.Close()
	var items []MemoryAuditItem
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return MemoryAuditOutput{}, err
		}
		items = append(items, memoryAuditItem(event))
	}
	return MemoryAuditOutput{Items: items}, rows.Err()
}

func (r *PostgresRepository) DeleteMemoryEntry(ctx context.Context, ident identity.LocalIdentity, entryID string, input DeleteMemoryEntryInput) (MemoryTombstone, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryTombstone{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return MemoryTombstone{}, err
	}
	defer tx.Rollback(ctx)
	entry, err := scanMemoryEntry(tx.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where id=$1 and user_id=$2 for update`, strings.TrimSpace(entryID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryTombstone{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if err != nil {
		return MemoryTombstone{}, err
	}
	if !memoryEntryReadableTo(entry, user.ID, MemoryEntryAccessInput{ScopeType: input.ScopeType, ScopeID: input.ScopeID, SourceThreadID: input.SourceThreadID, SourceRunID: input.SourceRunID}) {
		return MemoryTombstone{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if entry.Status == MemoryEntryTombstoned && entry.DeletedAt != nil {
		if err := tx.Commit(ctx); err != nil {
			return MemoryTombstone{}, err
		}
		return MemoryTombstone{EntryID: entry.ID, Status: string(MemoryEntryTombstoned), DeletedAt: *entry.DeletedAt}, nil
	}
	entry, err = scanMemoryEntry(tx.QueryRow(ctx, `update memory_entries set status='tombstoned', content='', summary='[deleted]', deleted_at=now(), deleted_by_user_id=$3, delete_reason=$4, updated_at=now() where id=$1 and user_id=$2 and status <> 'tombstoned' returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'')`, entry.ID, user.ID, user.ID, RedactEventText(strings.TrimSpace(input.Reason))))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryTombstone{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if err != nil {
		return MemoryTombstone{}, err
	}
	if err := r.appendMemoryAuditEventTx(ctx, tx, user, entry.SourceRunID, EventMemoryEntryDeleted, "Memory entry deleted", memoryEntryAuditMetadata(entry, "")); err != nil {
		return MemoryTombstone{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MemoryTombstone{}, err
	}
	return MemoryTombstone{EntryID: strings.TrimSpace(entryID), Status: string(MemoryEntryTombstoned), DeletedAt: *entry.DeletedAt}, nil
}

func (r *PostgresRepository) ProposeMemoryWrite(ctx context.Context, ident identity.LocalIdentity, input ProposeMemoryWriteInput) (MemoryWriteProposal, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	if key := strings.TrimSpace(input.IdempotencyKey); key != "" {
		proposal, err := scanMemoryProposal(r.Pool.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'') from memory_write_proposals where user_id=$1 and idempotency_key=$2`, user.ID, key))
		if err == nil {
			return proposal, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return MemoryWriteProposal{}, err
		}
	}
	scopeType, scopeID, err := normalizeMemoryScope(user.ID, input.ScopeType, input.ScopeID)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Content)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	status := MemoryWritePending
	if safety == MemorySafetyBlocked {
		status = MemoryWriteDenied
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `insert into memory_write_proposals (id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, source_thread_id, source_run_id, source_event_id, idempotency_key) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,nullif($10,''),nullif($11,''),nullif($12,''),$13) returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'')`,
		NewMemoryProposalID(), user.ID, scopeType, scopeID, title, summary, content, status, safety, strings.TrimSpace(input.SourceThreadID), strings.TrimSpace(input.SourceRunID), strings.TrimSpace(input.SourceEventID), strings.TrimSpace(input.IdempotencyKey))
	proposal, err := scanMemoryProposal(row)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	if err := r.appendMemoryAuditEventTx(ctx, tx, user, proposal.SourceRunID, EventMemoryWriteProposed, "Memory write proposed", memoryProposalAuditMetadata(proposal, "")); err != nil {
		return MemoryWriteProposal{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MemoryWriteProposal{}, err
	}
	return proposal, nil
}

func (r *PostgresRepository) ListMemoryWriteProposals(ctx context.Context, ident identity.LocalIdentity, input MemoryWriteProposalListInput) (MemoryWriteProposalListOutput, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteProposalListOutput{}, err
	}
	limit := memoryLimit(input.Limit)
	status := input.Status
	if status == "" {
		status = MemoryWritePending
	}
	conditions := []string{"user_id=$1", "status=$2"}
	args := []any{user.ID, status}
	if input.ScopeType != "" {
		args = append(args, input.ScopeType)
		conditions = append(conditions, "scope_type=$"+strconv.Itoa(len(args)))
	}
	if strings.TrimSpace(input.ScopeID) != "" {
		args = append(args, strings.TrimSpace(input.ScopeID))
		conditions = append(conditions, "scope_id=$"+strconv.Itoa(len(args)))
	}
	if strings.TrimSpace(input.SourceRunID) != "" {
		args = append(args, strings.TrimSpace(input.SourceRunID))
		conditions = append(conditions, "source_run_id=$"+strconv.Itoa(len(args)))
	}
	args = append(args, limit)
	query := `select id, user_id, scope_type, scope_id, title, summary, '' as content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), '' as idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, '' as decided_by_user_id, coalesce(decision_reason,'') from memory_write_proposals where ` + strings.Join(conditions, " and ") + ` order by created_at desc limit $` + strconv.Itoa(len(args))
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return MemoryWriteProposalListOutput{}, err
	}
	defer rows.Close()
	items := []MemoryWriteProposal{}
	for rows.Next() {
		proposal, err := scanMemoryProposal(rows)
		if err != nil {
			return MemoryWriteProposalListOutput{}, err
		}
		items = append(items, proposal)
	}
	if err := rows.Err(); err != nil {
		return MemoryWriteProposalListOutput{}, err
	}
	return MemoryWriteProposalListOutput{Items: items}, nil
}

func (r *PostgresRepository) UpdateMemoryWriteProposal(ctx context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteProposalUpdateInput) (MemoryWriteProposal, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Summary)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	if safety == MemorySafetyBlocked {
		return MemoryWriteProposal{}, NewError(CodeInvalidRequest, "Memory proposal edit contains sensitive content.")
	}
	proposal, err := scanMemoryProposal(r.Pool.QueryRow(ctx, `update memory_write_proposals set title=$3, summary=$4, content=$5, safety_state=$6 where id=$1 and user_id=$2 and status='pending' and safety_state <> 'blocked' returning id, user_id, scope_type, scope_id, title, summary, '' as content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), '' as idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, '' as decided_by_user_id, coalesce(decision_reason,'')`, strings.TrimSpace(proposalID), user.ID, title, summary, content, safety))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryWriteProposal{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	return proposal, nil
}

func (r *PostgresRepository) ApproveMemoryWrite(ctx context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteDecisionInput) (MemoryWriteDecision, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	defer tx.Rollback(ctx)
	proposal, err := scanMemoryProposal(tx.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'') from memory_write_proposals where id=$1 and user_id=$2 for update`, strings.TrimSpace(proposalID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryWriteDecision{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	if proposal.Status == MemoryWriteApproved && proposal.CreatedEntryID != "" {
		entry, err := scanMemoryEntry(tx.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where id=$1`, proposal.CreatedEntryID))
		if err != nil {
			return MemoryWriteDecision{}, err
		}
		entry.Content = ""
		if err := tx.Commit(ctx); err != nil {
			return MemoryWriteDecision{}, err
		}
		return MemoryWriteDecision{Proposal: proposal, Entry: entry}, nil
	}
	if proposal.Status != MemoryWritePending || proposal.SafetyState == MemorySafetyBlocked {
		return MemoryWriteDecision{}, NewError(CodeInvalidRequest, "Memory write cannot be approved.")
	}
	entry, err := scanMemoryEntry(tx.QueryRow(ctx, `insert into memory_entries (id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, source_thread_id, source_run_id, source_event_id, content_hash) values ($1,$2,$3,$4,$5,$6,$7,'approved',$8,nullif($9,''),nullif($10,''),nullif($11,''),$12) returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'')`,
		NewMemoryEntryID(), user.ID, proposal.ScopeType, proposal.ScopeID, proposal.Title, proposal.Summary, proposal.Content, proposal.SafetyState, proposal.SourceThreadID, proposal.SourceRunID, proposal.SourceEventID, memoryContentHash(proposal.ScopeType, proposal.ScopeID, proposal.Content)))
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	proposal, err = scanMemoryProposal(tx.QueryRow(ctx, `update memory_write_proposals set status='approved', created_entry_id=$1, decided_at=now(), decided_by_user_id=$2, decision_reason=$3 where id=$4 returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'')`, entry.ID, user.ID, RedactEventText(strings.TrimSpace(input.Reason)), proposal.ID))
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	entry.Content = ""
	if err := r.appendMemoryAuditEventTx(ctx, tx, user, proposal.SourceRunID, EventMemoryWriteApproved, "Memory write approved", memoryProposalAuditMetadata(proposal, entry.ID)); err != nil {
		return MemoryWriteDecision{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MemoryWriteDecision{}, err
	}
	return MemoryWriteDecision{Proposal: proposal, Entry: entry}, nil
}

func (r *PostgresRepository) DenyMemoryWrite(ctx context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteDecisionInput) (MemoryWriteDecision, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	defer tx.Rollback(ctx)
	proposal, err := scanMemoryProposal(tx.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'') from memory_write_proposals where id=$1 and user_id=$2 for update`, strings.TrimSpace(proposalID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryWriteDecision{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	if proposal.Status == MemoryWriteDenied {
		if err := tx.Commit(ctx); err != nil {
			return MemoryWriteDecision{}, err
		}
		return MemoryWriteDecision{Proposal: proposal}, nil
	}
	if proposal.Status == MemoryWriteApproved {
		return MemoryWriteDecision{}, NewError(CodeInvalidRequest, "Approved memory write cannot be denied.")
	}
	proposal, err = scanMemoryProposal(tx.QueryRow(ctx, `update memory_write_proposals set status='denied', decided_at=now(), decided_by_user_id=$3, decision_reason=$4 where id=$1 and user_id=$2 and status='pending' returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'')`, proposal.ID, user.ID, user.ID, RedactEventText(strings.TrimSpace(input.Reason))))
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	if err := r.appendMemoryAuditEventTx(ctx, tx, user, proposal.SourceRunID, EventMemoryWriteDenied, "Memory write denied", memoryProposalAuditMetadata(proposal, "")); err != nil {
		return MemoryWriteDecision{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MemoryWriteDecision{}, err
	}
	return MemoryWriteDecision{Proposal: proposal}, nil
}

func (r *PostgresRepository) appendMemoryAuditEvent(ctx context.Context, ident identity.LocalIdentity, runID string, eventType string, summary string, metadata map[string]any) error {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := r.appendMemoryAuditEventTx(ctx, tx, user, runID, eventType, summary, metadata); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) appendMemoryAuditEventTx(ctx context.Context, tx pgx.Tx, user User, runID string, eventType string, summary string, metadata map[string]any) error {
	runID = strings.TrimSpace(runID)
	threadID := ""
	var run Run
	runFound := false
	if runID != "" {
		run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
		if err == nil {
			threadID = run.ThreadID
			runFound = true
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}
	if !runFound {
		_, err := tx.Exec(ctx, `insert into memory_audit_events (id, user_id, thread_id, run_id, type, summary, metadata) values ($1,$2,$3,$4,$5,$6,$7)`, NewRunEventID(), user.ID, threadID, runID, eventType, RedactEventText(summary), mustJSON(RedactEventMetadata(metadata)))
		return err
	}
	if _, err := tx.Exec(ctx, `insert into memory_audit_events (id, user_id, thread_id, run_id, type, summary, metadata) values ($1,$2,$3,$4,$5,$6,$7)`, NewRunEventID(), user.ID, threadID, runID, eventType, RedactEventText(summary), mustJSON(RedactEventMetadata(metadata))); err != nil {
		return err
	}
	if _, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, eventType, summary, nil, metadata); err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) AppendRunEvent(ctx context.Context, ident identity.LocalIdentity, runID string, input AppendRunEventInput) (RunEvent, error) {
	input, err := NormalizeRunEventInput(input)
	if err != nil {
		return RunEvent{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return RunEvent{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return RunEvent{}, err
	}
	defer tx.Rollback(ctx)
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return RunEvent{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return RunEvent{}, err
	}
	if IsRunTerminal(run.Status) && !isTerminalRunEventAppendAllowed(input.Type) {
		return RunEvent{}, NewError(CodeInvalidRequest, "Terminal run cannot accept new events.")
	}
	if err := lockRunEventSequenceTx(ctx, tx, run.ID); err != nil {
		return RunEvent{}, err
	}
	var nextSequence int
	if err := tx.QueryRow(ctx, `select coalesce(max(sequence), 0) + 1 from run_events where run_id=$1`, run.ID).Scan(&nextSequence); err != nil {
		return RunEvent{}, err
	}
	event, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, run.ThreadID, user.ID, nextSequence, input.Category, input.Type, input.Summary, input.Content, mustJSON(input.Metadata)))
	if err != nil {
		return RunEvent{}, err
	}
	if isMemoryAuditEvent(event.Type) {
		if _, err := tx.Exec(ctx, `insert into memory_audit_events (id, user_id, thread_id, run_id, type, summary, metadata, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)`, NewRunEventID(), user.ID, run.ThreadID, run.ID, event.Type, event.Summary, mustJSON(event.Metadata), event.CreatedAt); err != nil {
			return RunEvent{}, err
		}
	}
	if err := updateRunStepStateProjectionTx(ctx, tx, event); err != nil {
		return RunEvent{}, err
	}
	status := run.Status
	completedAtSQL := `completed_at`
	errorCode := run.ErrorCode
	errorMessage := run.ErrorMessage
	if input.Category == RunEventCategoryFinal {
		status = statusFromFinalType(input.Type)
		completedAtSQL = `now()`
		if input.ErrorCode != "" {
			errorCode = &input.ErrorCode
		}
		if input.ErrorMessage != "" {
			errorMessage = &input.ErrorMessage
		}
	}
	if _, err := tx.Exec(ctx, `update runs set status=$1, updated_at=now(), completed_at=`+completedAtSQL+`, error_code=$4, error_message=$5 where id=$2 and user_id=$3`, status, run.ID, user.ID, errorCode, errorMessage); err != nil {
		return RunEvent{}, err
	}
	return event, tx.Commit(ctx)
}

func (r *PostgresRepository) GetRunStepState(ctx context.Context, ident identity.LocalIdentity, runID string) (RunStepState, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return RunStepState{}, err
	}
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return RunStepState{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return RunStepState{}, err
	}
	var raw []byte
	var lastSequence int
	err = r.Pool.QueryRow(ctx, `select last_sequence, state from run_step_state_projections where run_id=$1 and user_id=$2`, runID, user.ID).Scan(&lastSequence, &raw)
	if err == nil {
		var state RunStepState
		if err := json.Unmarshal(raw, &state); err != nil {
			return r.rebuildRunStepStateProjection(ctx, ident, runID, user.ID)
		}
		if !validRunStepStateProjection(run, lastSequence, state) {
			return r.rebuildRunStepStateProjection(ctx, ident, runID, user.ID)
		}
		events, err := r.ListRunEvents(ctx, ident, runID, lastSequence)
		if err != nil {
			return RunStepState{}, err
		}
		if len(events) == 0 {
			return state, nil
		}
		for _, event := range events {
			state = AdvanceRunStepState(state, event)
		}
		tag, err := r.Pool.Exec(ctx, `update run_step_state_projections set last_sequence=$1, state=$2, updated_at=now() where run_id=$3 and user_id=$4 and last_sequence <= $1`, state.LastEventSequence, mustJSON(state), runID, user.ID)
		if err != nil {
			return RunStepState{}, err
		}
		if tag.RowsAffected() == 0 {
			return r.GetRunStepState(ctx, ident, runID)
		}
		return state, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return RunStepState{}, err
	}
	return r.rebuildRunStepStateProjection(ctx, ident, runID, user.ID)
}

func (r *PostgresRepository) EnsureRunStepStateProjection(ctx context.Context, ident identity.LocalIdentity, runID string) error {
	_, err := r.GetRunStepState(ctx, ident, runID)
	return err
}

func (r *PostgresRepository) ClaimToolContinuation(ctx context.Context, ident identity.LocalIdentity, input ClaimToolContinuationInput) (RunEvent, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return RunEvent{}, false, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return RunEvent{}, false, err
	}
	defer tx.Rollback(ctx)
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 and thread_id=$3 for update`, strings.TrimSpace(input.RunID), user.ID, strings.TrimSpace(input.ThreadID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return RunEvent{}, false, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return RunEvent{}, false, err
	}
	if IsRunTerminal(run.Status) {
		if err := tx.Commit(ctx); err != nil {
			return RunEvent{}, false, err
		}
		return RunEvent{}, false, nil
	}
	state, err := runStepStateProjectionForTx(ctx, tx, run)
	if err != nil {
		return RunEvent{}, false, err
	}
	claimNow := time.Now().UTC()
	if !toolContinuationClaimAllowed(state, input.ToolCallID, input.JobID, claimNow, activePostgresContinuationJob(ctx, tx, run, user.ID)) {
		if err := tx.Commit(ctx); err != nil {
			return RunEvent{}, false, err
		}
		return RunEvent{}, false, nil
	}
	event, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, "model_request_started", "Model request started", nil, toolContinuationClaimMetadata(input, claimNow))
	if err != nil {
		return RunEvent{}, false, err
	}
	if _, err := tx.Exec(ctx, `update runs set updated_at=now() where id=$1 and user_id=$2`, run.ID, user.ID); err != nil {
		return RunEvent{}, false, err
	}
	return event, true, tx.Commit(ctx)
}

func (r *PostgresRepository) rebuildRunStepStateProjection(ctx context.Context, ident identity.LocalIdentity, runID string, userID string) (RunStepState, error) {
	events, err := r.ListRunEvents(ctx, ident, runID, 0)
	if err != nil {
		return RunStepState{}, err
	}
	state := RebuildRunStepState(events)
	tag, err := r.Pool.Exec(ctx, `insert into run_step_state_projections (run_id, thread_id, user_id, last_sequence, state) select id, thread_id, user_id, $2, $3 from runs where id=$1 and user_id=$4 on conflict (run_id) do update set last_sequence=excluded.last_sequence, state=excluded.state, updated_at=now() where run_step_state_projections.last_sequence <= excluded.last_sequence`, runID, state.LastEventSequence, mustJSON(state), userID)
	if err != nil {
		return RunStepState{}, err
	}
	if tag.RowsAffected() == 0 {
		return r.GetRunStepState(ctx, ident, runID)
	}
	return state, nil
}

func (r *PostgresRepository) GetToolCall(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, err
	}
	call, err := scanToolCall(r.Pool.QueryRow(ctx, `select tc.id, tc.thread_id, tc.run_id, tc.tool_call_id, tc.tool_name, tc.candidate_schema_hash, tc.arguments_summary, tc.approval_status, tc.execution_status, tc.result_summary, tc.error_code, tc.error_message, tc.requested_at, tc.updated_at from tool_calls tc join runs r on r.id=tc.run_id where tc.thread_id=$1 and tc.run_id=$2 and tc.tool_call_id=$3 and r.user_id=$4`, threadID, runID, strings.TrimSpace(toolCallID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return ToolCall{}, err
	}
	return call, nil
}

func (r *PostgresRepository) RecordToolCallRequest(ctx context.Context, ident identity.LocalIdentity, runID string, input RecordToolCallRequestInput) (ToolCall, []RunEvent, error) {
	input, err := ValidateToolCallRequestInput(input)
	if err != nil {
		return ToolCall{}, nil, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ToolCall{}, nil, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return ToolCall{}, nil, err
	}
	if IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Terminal runs cannot request tools.")
	}
	rows, err := tx.Query(ctx, `select tool_call_id, tool_name, approval_status, execution_status from tool_calls where run_id=$1 and execution_status in ('blocked', 'not_started', 'executing')`, run.ID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var existingToolCallID, existingToolName string
		var existingApprovalStatus ToolCallApprovalStatus
		var existingExecutionStatus ToolCallExecutionStatus
		if err := rows.Scan(&existingToolCallID, &existingToolName, &existingApprovalStatus, &existingExecutionStatus); err != nil {
			return ToolCall{}, nil, err
		}
		if existingToolCallID != input.ToolCallID && !pendingToolCallRequestCanCoexist(existingToolCallID, existingToolName, existingApprovalStatus, existingExecutionStatus, input) {
			return ToolCall{}, nil, NewError(CodeInvalidRequest, "Another tool call is already pending or executing.")
		}
	}
	if err := rows.Err(); err != nil {
		return ToolCall{}, nil, err
	}
	arguments := RedactEventMetadata(input.ArgumentsSummary)
	call, err := scanToolCall(tx.QueryRow(ctx, `insert into tool_calls (id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, arguments_hash, approval_status, execution_status) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) on conflict (run_id, tool_call_id) do nothing returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, NewToolCallID(), run.ThreadID, run.ID, input.ToolCallID, input.ToolName, input.CandidateSchemaHash, mustJSON(arguments), input.ArgumentsHash, input.ApprovalStatus, input.ExecutionStatus))
	if errors.Is(err, pgx.ErrNoRows) {
		var existingHash string
		existing, err := scanToolCallWithHash(tx.QueryRow(ctx, `select id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, coalesce(arguments_hash, ''), approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at from tool_calls where run_id=$1 and tool_call_id=$2`, run.ID, input.ToolCallID), &existingHash)
		if err != nil {
			return ToolCall{}, nil, err
		}
		if !toolCallRequestMatchesExisting(existing, existingHash, input, arguments) {
			return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call id was already used for a different request.")
		}
		return existing, nil, tx.Commit(ctx)
	}
	if err != nil {
		return ToolCall{}, nil, err
	}
	autoApproved := call.ApprovalStatus == ToolCallApprovalApproved && call.ExecutionStatus == ToolCallExecutionNotStarted
	if autoApproved {
		run.Status = RunStatusQueued
	} else {
		run.Status = RunStatusBlockedOnToolApproval
	}
	if _, err := tx.Exec(ctx, `update runs set status=$1, updated_at=now() where id=$2 and user_id=$3`, run.Status, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'retrying')`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	metadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
	if err != nil {
		return ToolCall{}, nil, err
	}
	requested, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallRequested, "Tool call requested", nil, metadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if autoApproved {
		approved, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallApproved, "Tool call auto-approved", nil, metadata)
		if err != nil {
			return ToolCall{}, nil, err
		}
		return call, []RunEvent{requested, approved}, tx.Commit(ctx)
	}
	required, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallApprovalRequired, "Tool approval required", nil, metadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{requested, required}, tx.Commit(ctx)
}

func (r *PostgresRepository) ApproveToolCall(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ApprovalStatus == ToolCallApprovalApproved {
		if call.ExecutionStatus == ToolCallExecutionNotStarted || call.ExecutionStatus == ToolCallExecutionExecuting || call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed {
			return call, nil, tx.Commit(ctx)
		}
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be approved.")
	}
	if call.ApprovalStatus != ToolCallApprovalRequired || call.ExecutionStatus != ToolCallExecutionBlocked || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be approved.")
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set approval_status='approved', execution_status='not_started', updated_at=now() where id=$1 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	run.Status = RunStatusQueued
	if _, err := tx.Exec(ctx, `update runs set status='queued', updated_at=now() where id=$1 and user_id=$2`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	workspaceRoot, err := r.workspaceRootPathForRunTx(ctx, tx, user.ID, run.ID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'leased', 'retrying')`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	jobID := NewBackgroundJobID()
	metadata := map[string]any{"source": string(run.Source), "job_id": jobID, "tool_call_id": call.ToolCallID, "resume_reason": "tool_call_approved"}
	if workspaceRoot != "" {
		metadata["workspace_root_path"] = workspaceRoot
		metadata["workspace_label"] = WorkspaceDisplayNameFromPath(workspaceRoot)
	}
	if _, err := tx.Exec(ctx, `insert into background_jobs (id, run_id, thread_id, user_id, kind, status, priority, max_attempts, scheduled_at, metadata) values ($1, $2, $3, $4, $5, 'queued', 50, 3, now(), $6)`, jobID, run.ID, run.ThreadID, user.ID, BackgroundJobKindRunExecution, mustJSON(metadata)); err != nil {
		return ToolCall{}, nil, err
	}
	toolMetadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
	if err != nil {
		return ToolCall{}, nil, err
	}
	event, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallApproved, "Tool call approved", nil, toolMetadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{event}, tx.Commit(ctx)
}

func (r *PostgresRepository) DenyToolCall(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ApprovalStatus == ToolCallApprovalDenied {
		return call, nil, tx.Commit(ctx)
	}
	if call.ApprovalStatus != ToolCallApprovalRequired || call.ExecutionStatus != ToolCallExecutionBlocked || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be denied.")
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set approval_status='denied', execution_status='cancelled', updated_at=now() where id=$1 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'leased', 'retrying')`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	stopped, err := scanRun(tx.QueryRow(ctx, `update runs set status='stopped', completed_at=now(), updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	toolMetadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
	if err != nil {
		return ToolCall{}, nil, err
	}
	denied, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryProgress, EventToolCallDenied, "Tool call denied by user", nil, toolMetadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	cancelled, err := cancelPostgresUnresolvedToolCallsTx(ctx, tx, stopped, call.ToolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	final, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryFinal, EventRunStopped, "Run stopped after tool denial", nil, map[string]any{"tool_call_id": call.ToolCallID, "reason": "tool_call_denied"})
	if err != nil {
		return ToolCall{}, nil, err
	}
	events := append([]RunEvent{denied}, cancelled...)
	events = append(events, final)
	return call, events, tx.Commit(ctx)
}

func cancelPostgresUnresolvedToolCallsTx(ctx context.Context, tx pgx.Tx, run Run, exceptToolCallID string) ([]RunEvent, error) {
	rows, err := tx.Query(ctx, `select id from tool_calls where run_id=$1 and tool_call_id<>$2 and execution_status in ('blocked', 'not_started', 'executing') for update`, run.ID, strings.TrimSpace(exceptToolCallID))
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	events := []RunEvent{}
	for _, id := range ids {
		call, err := scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='cancelled', updated_at=now() where id=$1 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, id))
		if err != nil {
			return nil, err
		}
		metadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
		if err != nil {
			return nil, err
		}
		event, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallCancelled, "Tool call cancelled", nil, metadata)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

func (r *PostgresRepository) StartToolCallExecution(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionExecuting || call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed || call.ExecutionStatus == ToolCallExecutionCancelled {
		return call, nil, tx.Commit(ctx)
	}
	if call.ApprovalStatus != ToolCallApprovalApproved || call.ExecutionStatus != ToolCallExecutionNotStarted || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot execute.")
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='executing', updated_at=now() where id=$1 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	toolMetadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
	if err != nil {
		return ToolCall{}, nil, err
	}
	event, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallExecuting, "Tool call executing", nil, toolMetadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{event}, tx.Commit(ctx)
}

func (r *PostgresRepository) CompleteToolCallSuccess(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, resultSummary map[string]any) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionSucceeded {
		return call, nil, tx.Commit(ctx)
	}
	if call.ExecutionStatus != ToolCallExecutionExecuting || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot succeed.")
	}
	result := RedactEventMetadata(resultSummary)
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='succeeded', result_summary=$1, updated_at=now() where id=$2 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, mustJSON(result), call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	nextStatus, err := postgresRunStatusAfterToolSuccess(ctx, tx, run.ID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	running, err := scanRun(tx.QueryRow(ctx, `update runs set status=$1, completed_at=null, updated_at=now() where id=$2 and user_id=$3 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, nextStatus, run.ID, user.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	toolMetadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
	if err != nil {
		return ToolCall{}, nil, err
	}
	succeeded, err := insertRunEvent(ctx, tx, running, RunEventCategoryProgress, EventToolCallSucceeded, "Tool call succeeded", nil, toolMetadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{succeeded}, tx.Commit(ctx)
}

func postgresRunStatusAfterToolSuccess(ctx context.Context, tx pgx.Tx, runID string) (RunStatus, error) {
	rows, err := tx.Query(ctx, `select approval_status, execution_status from tool_calls where run_id=$1 and execution_status in ('blocked', 'not_started', 'executing')`, runID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	hasReady := false
	hasBlocked := false
	for rows.Next() {
		var approvalStatus ToolCallApprovalStatus
		var executionStatus ToolCallExecutionStatus
		if err := rows.Scan(&approvalStatus, &executionStatus); err != nil {
			return "", err
		}
		if executionStatus == ToolCallExecutionExecuting {
			return RunStatusRunning, nil
		}
		if approvalStatus == ToolCallApprovalApproved && executionStatus == ToolCallExecutionNotStarted {
			hasReady = true
		}
		if approvalStatus == ToolCallApprovalRequired && executionStatus == ToolCallExecutionBlocked {
			hasBlocked = true
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if hasReady {
		return RunStatusQueued, nil
	}
	if hasBlocked {
		return RunStatusBlockedOnToolApproval, nil
	}
	return RunStatusRunning, nil
}

func (r *PostgresRepository) FailToolCallExecution(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, errorCode string, errorMessage string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionFailed {
		return call, nil, tx.Commit(ctx)
	}
	if call.ExecutionStatus != ToolCallExecutionExecuting || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot fail.")
	}
	code := strings.TrimSpace(errorCode)
	if code == "" {
		code = "tool_execution_failed"
	}
	message := RedactEventText(strings.TrimSpace(errorMessage))
	if message == "" {
		message = "Tool execution failed."
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='failed', error_code=$1, error_message=$2, updated_at=now() where id=$3 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, code, message, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	failedRun, err := scanRun(tx.QueryRow(ctx, `update runs set status='failed', completed_at=now(), updated_at=now(), error_code=$1, error_message=$2 where id=$3 and user_id=$4 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, code, message, run.ID, user.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	toolMetadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
	if err != nil {
		return ToolCall{}, nil, err
	}
	failed, err := insertRunEvent(ctx, tx, failedRun, RunEventCategoryError, EventToolCallFailed, message, nil, toolMetadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	final, err := insertRunEvent(ctx, tx, failedRun, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"tool_call_id": call.ToolCallID, "error_code": code})
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{failed, final}, tx.Commit(ctx)
}

func (r *PostgresRepository) RecordToolCallExecutionFailure(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, errorCode string, errorMessage string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionFailed {
		return call, nil, tx.Commit(ctx)
	}
	if call.ExecutionStatus != ToolCallExecutionExecuting || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot fail.")
	}
	code := strings.TrimSpace(errorCode)
	if code == "" {
		code = "tool_execution_failed"
	}
	message := RedactEventText(strings.TrimSpace(errorMessage))
	if message == "" {
		message = "Tool execution failed."
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='failed', error_code=$1, error_message=$2, updated_at=now() where id=$3 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, code, message, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	nextStatus, err := postgresRunStatusAfterToolSuccess(ctx, tx, run.ID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	running, err := scanRun(tx.QueryRow(ctx, `update runs set status=$1, completed_at=null, updated_at=now() where id=$2 and user_id=$3 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, nextStatus, run.ID, user.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	toolMetadata, err := toolCallEventMetadataForPostgresState(ctx, tx, run, call)
	if err != nil {
		return ToolCall{}, nil, err
	}
	failed, err := insertRunEvent(ctx, tx, running, RunEventCategoryError, EventToolCallFailed, message, nil, toolMetadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{failed}, tx.Commit(ctx)
}

func (r *PostgresRepository) ClaimBackgroundJob(ctx context.Context, ident identity.LocalIdentity, input ClaimBackgroundJobInput) (BackgroundJob, Run, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	workerID := strings.TrimSpace(input.WorkerID)
	if workerID == "" {
		return BackgroundJob{}, Run{}, false, NewError(CodeInvalidRequest, "Worker id is required.")
	}
	leaseSeconds := input.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	defer tx.Rollback(ctx)
	job, err := scanBackgroundJob(tx.QueryRow(ctx, `select bj.id, bj.run_id, bj.thread_id, bj.user_id, bj.kind, bj.status, bj.priority, bj.attempt_count, bj.max_attempts, bj.scheduled_at, bj.leased_by, bj.lease_expires_at, bj.ownership_version, bj.metadata, bj.last_error_code, bj.last_error_message, bj.created_at, bj.updated_at from background_jobs bj join runs r on r.id=bj.run_id and r.user_id=bj.user_id where bj.user_id=$1 and bj.status='queued' and bj.scheduled_at<=now() and r.stop_requested_at is null and r.status not in ('completed', 'failed', 'stopped') order by bj.priority asc, bj.created_at asc, bj.id asc for update of bj skip locked limit 1`, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, Run{}, false, nil
	}
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, job.RunID, user.ID))
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	if IsRunTerminal(run.Status) || run.StopRequestedAt != nil {
		cancelled, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='cancelled', updated_at=now() where id=$1 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, job.ID))
		if err != nil {
			return BackgroundJob{}, Run{}, false, err
		}
		return cancelled, run, false, tx.Commit(ctx)
	}
	leased, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='leased', leased_by=$1, lease_expires_at=now() + ($2::int * interval '1 second'), attempt_count=attempt_count+1, ownership_version=ownership_version+1, updated_at=now() where id=$3 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, workerID, leaseSeconds, job.ID))
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	run, err = scanRun(tx.QueryRow(ctx, `update runs set status='running', updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	if _, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventJobClaimed, "Job claimed", nil, map[string]any{"job_id": leased.ID, "worker_id": workerID, "attempt": leased.AttemptCount}); err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	return leased, run, true, tx.Commit(ctx)
}

func (r *PostgresRepository) RenewBackgroundJobLease(ctx context.Context, ident identity.LocalIdentity, input RenewBackgroundJobLeaseInput) (BackgroundJob, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, false, err
	}
	leaseSeconds := input.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	job, err := scanBackgroundJob(r.Pool.QueryRow(ctx, `update background_jobs set lease_expires_at=now() + ($1::int * interval '1 second'), updated_at=now() where id=$2 and user_id=$3 and leased_by=$4 and ownership_version=$5 and status='leased' returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, leaseSeconds, input.JobID, user.ID, strings.TrimSpace(input.WorkerID), input.OwnershipVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, false, nil
	}
	if err != nil {
		return BackgroundJob{}, false, err
	}
	run, err := r.GetRun(ctx, ident, job.RunID)
	if err == nil && !IsRunTerminal(run.Status) {
		_, _ = r.AppendRunEvent(ctx, ident, job.RunID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: EventLeaseRenewed, Summary: "Lease renewed", Metadata: map[string]any{"job_id": job.ID, "worker_id": strings.TrimSpace(input.WorkerID), "ownership_version": input.OwnershipVersion}})
	}
	return job, true, nil
}

func (r *PostgresRepository) RecoverBackgroundJobs(ctx context.Context, ident identity.LocalIdentity, input RecoverBackgroundJobsInput) ([]BackgroundJobRecovery, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	code := strings.TrimSpace(input.ErrorCode)
	if code == "" {
		code = "worker_lease_expired"
	}
	message := RedactEventText(strings.TrimSpace(input.ErrorMessage))
	if message == "" {
		message = "Worker lease expired."
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `select id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at from background_jobs where user_id=$1 and status='leased' and lease_expires_at < now() order by lease_expires_at asc, id asc for update skip locked limit $2`, user.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := []BackgroundJob{}
	for rows.Next() {
		job, err := scanBackgroundJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	recoveries := []BackgroundJobRecovery{}
	for _, job := range jobs {
		run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, job.RunID, user.ID))
		if err != nil || IsRunTerminal(run.Status) {
			continue
		}
		previousWorkerID := ""
		if job.LeasedBy != nil {
			previousWorkerID = *job.LeasedBy
		}
		if job.AttemptCount >= job.MaxAttempts {
			dead, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='dead', leased_by=null, lease_expires_at=null, last_error_code=$1, last_error_message=$2, updated_at=now() where id=$3 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, code, message, job.ID))
			if err != nil {
				return nil, err
			}
			failed, err := scanRun(tx.QueryRow(ctx, `update runs set status='failed', completed_at=now(), updated_at=now(), error_code=$1, error_message=$2 where id=$3 and user_id=$4 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, code, message, run.ID, user.ID))
			if err != nil {
				return nil, err
			}
			toolEvents, err := failPostgresExecutingToolCallsForRecovery(ctx, tx, failed, code, message)
			if err != nil {
				return nil, err
			}
			exhausted, err := insertRunEvent(ctx, tx, failed, RunEventCategoryError, EventJobRetryExhausted, message, nil, map[string]any{"job_id": dead.ID, "attempt_count": dead.AttemptCount, "error_code": code})
			if err != nil {
				return nil, err
			}
			final, err := insertRunEvent(ctx, tx, failed, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"job_id": dead.ID, "error_code": code})
			if err != nil {
				return nil, err
			}
			events := append(toolEvents, exhausted, final)
			recoveries = append(recoveries, BackgroundJobRecovery{Job: dead, Run: failed, Events: events, Exhausted: true})
			continue
		}
		backoffSeconds := int(retryBackoffDuration(job.AttemptCount).Seconds())
		queued, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='queued', leased_by=null, lease_expires_at=null, scheduled_at=now() + ($1::int * interval '1 second'), last_error_code=$2, last_error_message=$3, updated_at=now() where id=$4 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, backoffSeconds, code, message, job.ID))
		if err != nil {
			return nil, err
		}
		recoveringRun, err := scanRun(tx.QueryRow(ctx, `update runs set status='recovering', updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
		if err != nil {
			return nil, err
		}
		toolEvents, err := resetPostgresExecutingToolCallsForRecovery(ctx, tx, recoveringRun)
		if err != nil {
			return nil, err
		}
		recovering, err := insertRunEvent(ctx, tx, recoveringRun, RunEventCategoryProgress, EventJobRecovering, "Job recovering", nil, map[string]any{"job_id": queued.ID, "previous_worker_id": previousWorkerID, "attempt": queued.AttemptCount})
		if err != nil {
			return nil, err
		}
		retry, err := insertRunEvent(ctx, tx, recoveringRun, RunEventCategoryProgress, EventJobRetryScheduled, "Job retry scheduled", nil, map[string]any{"job_id": queued.ID, "next_attempt": queued.AttemptCount + 1, "scheduled_at": queued.ScheduledAt})
		if err != nil {
			return nil, err
		}
		events := append(toolEvents, recovering, retry)
		recoveries = append(recoveries, BackgroundJobRecovery{Job: queued, Run: recoveringRun, Events: events})
	}
	return recoveries, tx.Commit(ctx)
}

func resetPostgresExecutingToolCallsForRecovery(ctx context.Context, tx pgx.Tx, run Run) ([]RunEvent, error) {
	calls, err := updatePostgresExecutingToolCalls(ctx, tx, run.ID, ToolCallExecutionNotStarted, "", "")
	if err != nil {
		return nil, err
	}
	state, err := runStepStateProjectionForTx(ctx, tx, run)
	if err != nil {
		return nil, err
	}
	events := make([]RunEvent, 0, len(calls))
	for _, call := range calls {
		metadata := toolCallEventMetadataForState(state, call)
		metadata["recovery_reason"] = "worker_lease_expired"
		event, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallApproved, "Tool call returned to queue after worker recovery", nil, metadata)
		if err != nil {
			return nil, err
		}
		state = AdvanceRunStepState(state, event)
		events = append(events, event)
	}
	return events, nil
}

func failPostgresExecutingToolCallsForRecovery(ctx context.Context, tx pgx.Tx, run Run, code string, message string) ([]RunEvent, error) {
	calls, err := updatePostgresExecutingToolCalls(ctx, tx, run.ID, ToolCallExecutionFailed, code, message)
	if err != nil {
		return nil, err
	}
	state, err := runStepStateProjectionForTx(ctx, tx, run)
	if err != nil {
		return nil, err
	}
	events := make([]RunEvent, 0, len(calls))
	for _, call := range calls {
		metadata := toolCallEventMetadataForState(state, call)
		metadata["recovery_reason"] = "worker_lease_exhausted"
		event, err := insertRunEvent(ctx, tx, run, RunEventCategoryError, EventToolCallFailed, message, nil, metadata)
		if err != nil {
			return nil, err
		}
		state = AdvanceRunStepState(state, event)
		events = append(events, event)
	}
	return events, nil
}

func updatePostgresExecutingToolCalls(ctx context.Context, tx pgx.Tx, runID string, status ToolCallExecutionStatus, code string, message string) ([]ToolCall, error) {
	var rows pgx.Rows
	var err error
	if status == ToolCallExecutionFailed {
		rows, err = tx.Query(ctx, `update tool_calls set execution_status=$1, error_code=$2, error_message=$3, updated_at=now() where run_id=$4 and execution_status='executing' returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, status, code, message, runID)
	} else {
		rows, err = tx.Query(ctx, `update tool_calls set execution_status=$1, updated_at=now() where run_id=$2 and execution_status='executing' returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, status, runID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	calls := []ToolCall{}
	for rows.Next() {
		call, err := scanToolCall(rows)
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	return calls, rows.Err()
}

func (r *PostgresRepository) CompleteBackgroundJob(ctx context.Context, ident identity.LocalIdentity, input CompleteBackgroundJobInput) (BackgroundJob, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, false, err
	}
	job, err := scanBackgroundJob(r.Pool.QueryRow(ctx, `update background_jobs set status='completed', updated_at=now() where id=$1 and user_id=$2 and leased_by=$3 and ownership_version=$4 and status='leased' returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, input.JobID, user.ID, strings.TrimSpace(input.WorkerID), input.OwnershipVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, false, nil
	}
	return job, true, err
}

func (r *PostgresRepository) FailBackgroundJob(ctx context.Context, ident identity.LocalIdentity, input FailBackgroundJobInput) (BackgroundJob, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, false, err
	}
	code := strings.TrimSpace(input.ErrorCode)
	message := RedactEventText(strings.TrimSpace(input.ErrorMessage))
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return BackgroundJob{}, false, err
	}
	defer tx.Rollback(ctx)
	job, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='failed', last_error_code=$1, last_error_message=$2, updated_at=now() where id=$3 and user_id=$4 and leased_by=$5 and ownership_version=$6 and status='leased' returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, code, message, input.JobID, user.ID, strings.TrimSpace(input.WorkerID), input.OwnershipVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, false, nil
	}
	if err != nil {
		return BackgroundJob{}, false, err
	}
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, job.RunID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.Commit(ctx); err != nil {
			return BackgroundJob{}, false, err
		}
		return job, true, nil
	}
	if err != nil {
		return BackgroundJob{}, false, err
	}
	if !IsRunTerminal(run.Status) {
		failed, err := scanRun(tx.QueryRow(ctx, `update runs set status='failed', completed_at=now(), updated_at=now(), error_code=$1, error_message=$2 where id=$3 and user_id=$4 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, code, message, run.ID, user.ID))
		if err != nil {
			return BackgroundJob{}, false, err
		}
		if _, err := insertRunEvent(ctx, tx, failed, RunEventCategoryError, EventJobAttemptFailed, message, nil, map[string]any{"job_id": job.ID, "attempt": job.AttemptCount, "error_code": code}); err != nil {
			return BackgroundJob{}, false, err
		}
		if _, err := insertRunEvent(ctx, tx, failed, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"job_id": job.ID, "error_code": code}); err != nil {
			return BackgroundJob{}, false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return BackgroundJob{}, false, err
	}
	return job, true, nil
}

func (r *PostgresRepository) WorkerQueueDiagnostics(ctx context.Context, ident identity.LocalIdentity) (WorkerQueueDiagnostics, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return WorkerQueueDiagnostics{}, err
	}
	diagnostics := WorkerQueueDiagnostics{QueueStatus: WorkerQueueStatusReady, WorkerStatus: WorkerStatusReady}
	row := r.Pool.QueryRow(ctx, `select
		count(*) filter (where status='queued'),
		count(*) filter (where status='leased'),
		count(*) filter (where status='leased' and lease_expires_at < now()),
		count(*) filter (where status='retrying'),
		count(*) filter (where status='dead'),
		now()
		from background_jobs where user_id=$1`, user.ID)
	if err := row.Scan(&diagnostics.QueuedCount, &diagnostics.LeasedCount, &diagnostics.StaleCount, &diagnostics.RetryingCount, &diagnostics.DeadCount, &diagnostics.UpdatedAt); err != nil {
		return WorkerQueueDiagnostics{}, err
	}
	toolRow := r.Pool.QueryRow(ctx, `select
			count(*) filter (where tc.approval_status='required' and tc.execution_status='blocked'),
			count(*) filter (where tc.approval_status='approved' and tc.execution_status='not_started')
			from tool_calls tc join runs r on r.id=tc.run_id where r.user_id=$1`, user.ID)
	if err := toolRow.Scan(&diagnostics.BlockedToolApprovalCount, &diagnostics.ResumableToolCallCount); err != nil {
		return WorkerQueueDiagnostics{}, err
	}
	if diagnostics.StaleCount > 0 || diagnostics.RetryingCount > 0 || diagnostics.DeadCount > 0 {
		diagnostics.QueueStatus = WorkerQueueStatusDegraded
		diagnostics.WorkerStatus = WorkerStatusDegraded
	}
	return diagnostics, nil
}

func insertRunEvent(ctx context.Context, tx pgx.Tx, run Run, category RunEventCategory, eventType string, summary string, content *string, metadata map[string]any) (RunEvent, error) {
	if err := lockRunEventSequenceTx(ctx, tx, run.ID); err != nil {
		return RunEvent{}, err
	}
	var nextSequence int
	if err := tx.QueryRow(ctx, `select coalesce(max(sequence), 0) + 1 from run_events where run_id=$1`, run.ID).Scan(&nextSequence); err != nil {
		return RunEvent{}, err
	}
	summary = RedactEventText(summary)
	metadata = AnnotateRunStepMetadata(eventType, summary, metadata)
	event, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, run.ThreadID, run.UserID, nextSequence, category, eventType, RedactEventText(summary), content, mustJSON(RedactEventMetadata(metadata))))
	if err != nil {
		return RunEvent{}, err
	}
	return event, updateRunStepStateProjectionTx(ctx, tx, event)
}

func lockRunEventSequenceTx(ctx context.Context, tx pgx.Tx, runID string) error {
	_, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtext($1))`, runID)
	return err
}

func runStepStateProjectionForPool(ctx context.Context, pool *pgxpool.Pool, runID string) (RunStepState, error) {
	var raw []byte
	var lastSequence int
	err := pool.QueryRow(ctx, `select last_sequence, state from run_step_state_projections where run_id=$1`, runID).Scan(&lastSequence, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return runStepStateFromPoolEvents(ctx, pool, runID, 0)
	}
	if err != nil {
		return RunStepState{}, err
	}
	var state RunStepState
	if err := json.Unmarshal(raw, &state); err != nil {
		return runStepStateFromPoolEvents(ctx, pool, runID, 0)
	}
	events, err := runEventsFromPool(ctx, pool, runID, lastSequence)
	if err != nil {
		return RunStepState{}, err
	}
	for _, event := range events {
		state = AdvanceRunStepState(state, event)
	}
	return state, nil
}

func runStepStateProjectionForTx(ctx context.Context, tx pgx.Tx, run Run) (RunStepState, error) {
	var raw []byte
	var lastSequence int
	err := tx.QueryRow(ctx, `select last_sequence, state from run_step_state_projections where run_id=$1 and user_id=$2 for update`, run.ID, run.UserID).Scan(&lastSequence, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		state, err := runStepStateFromTxEvents(ctx, tx, run.ID, 0)
		if err != nil {
			return RunStepState{}, err
		}
		return state, upsertRunStepStateProjectionForRunTx(ctx, tx, run, state)
	}
	if err != nil {
		return RunStepState{}, err
	}
	var state RunStepState
	if err := json.Unmarshal(raw, &state); err != nil {
		state, err := runStepStateFromTxEvents(ctx, tx, run.ID, 0)
		if err != nil {
			return RunStepState{}, err
		}
		return state, upsertRunStepStateProjectionForRunTx(ctx, tx, run, state)
	}
	if !validRunStepStateProjection(run, lastSequence, state) {
		state, err := runStepStateFromTxEvents(ctx, tx, run.ID, 0)
		if err != nil {
			return RunStepState{}, err
		}
		return state, upsertRunStepStateProjectionForRunTx(ctx, tx, run, state)
	}
	events, err := runEventsFromTx(ctx, tx, run.ID, lastSequence)
	if err != nil {
		return RunStepState{}, err
	}
	if len(events) == 0 {
		return state, nil
	}
	for _, event := range events {
		state = AdvanceRunStepState(state, event)
	}
	return state, upsertRunStepStateProjectionForRunTx(ctx, tx, run, state)
}

func activePostgresContinuationJob(ctx context.Context, tx pgx.Tx, run Run, userID string) func(string) bool {
	return func(jobID string) bool {
		var active bool
		err := tx.QueryRow(ctx, `select exists(
			select 1 from background_jobs
			where id=$1 and run_id=$2 and user_id=$3 and status='leased' and lease_expires_at > now()
		)`, strings.TrimSpace(jobID), run.ID, userID).Scan(&active)
		return err == nil && active
	}
}

func updateRunStepStateProjectionTx(ctx context.Context, tx pgx.Tx, event RunEvent) error {
	var raw []byte
	var lastSequence int
	err := tx.QueryRow(ctx, `select last_sequence, state from run_step_state_projections where run_id=$1 for update`, event.RunID).Scan(&lastSequence, &raw)
	state := RunStepState{}
	if err == nil {
		if err := json.Unmarshal(raw, &state); err != nil {
			rebuilt, err := runStepStateFromTxEvents(ctx, tx, event.RunID, 0)
			if err != nil {
				return err
			}
			state = rebuilt
			return upsertRunStepStateProjectionTx(ctx, tx, event, state)
		}
		if !validRunStepStateProjectionForSequence(lastSequence, state) {
			rebuilt, err := runStepStateFromTxEvents(ctx, tx, event.RunID, 0)
			if err != nil {
				return err
			}
			state = rebuilt
			return upsertRunStepStateProjectionTx(ctx, tx, event, state)
		}
		events, err := runEventsFromTx(ctx, tx, event.RunID, lastSequence)
		if err != nil {
			return err
		}
		for _, next := range events {
			state = AdvanceRunStepState(state, next)
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	} else {
		rebuilt, err := runStepStateFromTxEvents(ctx, tx, event.RunID, 0)
		if err != nil {
			return err
		}
		state = rebuilt
	}
	return upsertRunStepStateProjectionTx(ctx, tx, event, state)
}

func upsertRunStepStateProjectionTx(ctx context.Context, tx pgx.Tx, event RunEvent, state RunStepState) error {
	tag, err := tx.Exec(ctx, `insert into run_step_state_projections (run_id, thread_id, user_id, last_sequence, state)
values ($1,$2,$3,$4,$5)
on conflict (run_id) do update set last_sequence=excluded.last_sequence, state=excluded.state, updated_at=now() where run_step_state_projections.last_sequence <= excluded.last_sequence`,
		event.RunID, event.ThreadID, event.UserID, state.LastEventSequence, mustJSON(state))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return NewError(CodeInvalidRequest, "Run step projection was superseded.")
	}
	return nil
}

func upsertRunStepStateProjectionForRunTx(ctx context.Context, tx pgx.Tx, run Run, state RunStepState) error {
	tag, err := tx.Exec(ctx, `insert into run_step_state_projections (run_id, thread_id, user_id, last_sequence, state)
values ($1,$2,$3,$4,$5)
on conflict (run_id) do update set last_sequence=excluded.last_sequence, state=excluded.state, updated_at=now() where run_step_state_projections.last_sequence <= excluded.last_sequence`,
		run.ID, run.ThreadID, run.UserID, state.LastEventSequence, mustJSON(state))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return NewError(CodeInvalidRequest, "Run step projection was superseded.")
	}
	return nil
}

func validRunStepStateProjection(run Run, lastSequence int, state RunStepState) bool {
	if !validRunStepStateProjectionForSequence(lastSequence, state) {
		return false
	}
	if run.Source == RunSourceModelGateway && (strings.TrimSpace(state.TriggerMessageID) == "" || strings.TrimSpace(state.ProviderID) == "") {
		return false
	}
	return true
}

func validRunStepStateProjectionForSequence(lastSequence int, state RunStepState) bool {
	if lastSequence < 0 || state.LastEventSequence != lastSequence {
		return false
	}
	for _, sequence := range []int{state.LastCompletedSequence, state.LastContinuationSequence, state.LastContinuationOutputSequence} {
		if sequence < 0 || sequence > lastSequence {
			return false
		}
	}
	for _, step := range state.Steps {
		if step.Sequence < 0 || step.Sequence > lastSequence {
			return false
		}
	}
	for _, step := range state.PendingToolCalls {
		if step.Sequence < 0 || step.Sequence > lastSequence {
			return false
		}
	}
	for _, step := range state.CompletedToolResults {
		if step.Sequence < 0 || step.Sequence > lastSequence {
			return false
		}
	}
	if state.Terminal != nil && (state.Terminal.Sequence < 0 || state.Terminal.Sequence > lastSequence) {
		return false
	}
	return true
}

func runStepStateFromTxEvents(ctx context.Context, tx pgx.Tx, runID string, afterSequence int) (RunStepState, error) {
	events, err := runEventsFromTx(ctx, tx, runID, afterSequence)
	if err != nil {
		return RunStepState{}, err
	}
	return RebuildRunStepState(events), nil
}

func runStepStateFromPoolEvents(ctx context.Context, pool *pgxpool.Pool, runID string, afterSequence int) (RunStepState, error) {
	events, err := runEventsFromPool(ctx, pool, runID, afterSequence)
	if err != nil {
		return RunStepState{}, err
	}
	return RebuildRunStepState(events), nil
}

func runEventsFromTx(ctx context.Context, tx pgx.Tx, runID string, afterSequence int) ([]RunEvent, error) {
	rows, err := tx.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where run_id=$1 and sequence>$2 order by sequence asc, id asc`, runID, afterSequence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRunEvents(rows)
}

func runEventsFromPool(ctx context.Context, pool *pgxpool.Pool, runID string, afterSequence int) ([]RunEvent, error) {
	rows, err := pool.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where run_id=$1 and sequence>$2 order by sequence asc, id asc`, runID, afterSequence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRunEvents(rows)
}

func (r *PostgresRepository) providerRouteForRunTx(ctx context.Context, tx pgx.Tx, runID string) (ProviderRoute, error) {
	var raw []byte
	err := tx.QueryRow(ctx, `select metadata from run_events where run_id=$1 and type='run_created' order by sequence asc limit 1`, runID).Scan(&raw)
	if err == nil {
		metadata := map[string]any{}
		_ = json.Unmarshal(raw, &metadata)
		providerID := metadataStringValue(metadata, "provider_id")
		return ProviderRoute{ProviderID: providerID, Model: metadataStringValue(metadata, "model"), Available: providerID != ""}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ProviderRoute{}, err
	}
	err = tx.QueryRow(ctx, `select metadata from background_jobs where run_id=$1 order by created_at asc, id asc limit 1`, runID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProviderRoute{}, nil
	}
	if err != nil {
		return ProviderRoute{}, err
	}
	metadata := map[string]any{}
	_ = json.Unmarshal(raw, &metadata)
	providerID := metadataStringValue(metadata, "provider_id")
	return ProviderRoute{ProviderID: providerID, Model: metadataStringValue(metadata, "model"), Available: providerID != ""}, nil
}

func toolCallEventMetadataForPostgresState(ctx context.Context, tx pgx.Tx, run Run, call ToolCall) (map[string]any, error) {
	state, err := runStepStateProjectionForTx(ctx, tx, run)
	if err != nil {
		return nil, err
	}
	return toolCallEventMetadataForState(state, call), nil
}

func scopedPostgresToolCall(ctx context.Context, tx pgx.Tx, userID string, threadID string, runID string, toolCallID string) (Run, ToolCall, error) {
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and thread_id=$2 and user_id=$3 for update`, runID, threadID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return Run{}, ToolCall{}, err
	}
	call, err := scanToolCall(tx.QueryRow(ctx, `select id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at from tool_calls where thread_id=$1 and run_id=$2 and tool_call_id=$3 for update`, threadID, runID, strings.TrimSpace(toolCallID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return Run{}, ToolCall{}, err
	}
	return run, call, nil
}

func (r *PostgresRepository) StopRun(ctx context.Context, ident identity.LocalIdentity, runID string) (StopRunOutput, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return StopRunOutput{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return StopRunOutput{}, err
	}
	defer tx.Rollback(ctx)
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return StopRunOutput{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return StopRunOutput{}, err
	}
	if IsRunTerminal(run.Status) {
		return StopRunOutput{Run: run, Result: StopRunResultAlreadyTerminal}, nil
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'leased', 'retrying')`, run.ID, user.ID); err != nil {
		return StopRunOutput{}, err
	}
	cancelled, err := cancelPostgresUnresolvedToolCallsTx(ctx, tx, run, "")
	if err != nil {
		return StopRunOutput{}, err
	}
	cascadeEvents, err := stopPostgresDelegatedChildRunsTx(ctx, tx, user.ID, run)
	if err != nil {
		return StopRunOutput{}, err
	}
	stopped, err := scanRun(tx.QueryRow(ctx, `update runs set status='stopped', stop_requested_at=coalesce(stop_requested_at, now()), updated_at=now(), completed_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return StopRunOutput{}, err
	}
	lifecycle, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryProgress, EventStopRequested, "Stop requested", nil, map[string]any{})
	if err != nil {
		return StopRunOutput{}, err
	}
	final, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryFinal, EventRunStopped, "Run stopped", nil, map[string]any{})
	if err != nil {
		return StopRunOutput{}, err
	}
	events := append(cancelled, cascadeEvents...)
	events = append(events, lifecycle, final)
	return StopRunOutput{Run: stopped, Result: StopRunResultStopped, Events: events}, tx.Commit(ctx)
}

func stopPostgresDelegatedChildRunsTx(ctx context.Context, tx pgx.Tx, userID string, parentRun Run) ([]RunEvent, error) {
	rows, err := tx.Query(ctx, `select id, thread_id, run_id, role, goal, status, result_summary, coalesce(child_thread_id, ''), coalesce(child_run_id, ''), coalesce(parent_tool_call_id, ''), delegated_at, created_at, updated_at from agent_tasks where user_id=$1 and run_id=$2 and thread_id=$3 and status=$4 and child_run_id is not null for update`, userID, parentRun.ID, parentRun.ThreadID, AgentTaskStatusInProgress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []AgentTask{}
	for rows.Next() {
		task, err := scanAgentTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	events := []RunEvent{}
	for _, task := range tasks {
		childRun, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, task.ChildRunID, userID))
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `update agent_tasks set status=$1, result_summary=$2, updated_at=now() where id=$3 and user_id=$4`, AgentTaskStatusFailed, RedactEventText("Parent run stopped before delegated child run completed."), task.ID, userID); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'leased', 'retrying')`, childRun.ID, userID); err != nil {
			return nil, err
		}
		if IsRunTerminal(childRun.Status) {
			continue
		}
		stoppedChild, err := scanRun(tx.QueryRow(ctx, `update runs set status='stopped', stop_requested_at=coalesce(stop_requested_at, now()), updated_at=now(), completed_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, childRun.ID, userID))
		if err != nil {
			return nil, err
		}
		cancelled, err := cancelPostgresUnresolvedToolCallsTx(ctx, tx, stoppedChild, "")
		if err != nil {
			return nil, err
		}
		metadata := map[string]any{"parent_run_id": parentRun.ID, "parent_thread_id": parentRun.ThreadID, "agent_task_id": task.ID, "reason": "parent_run_stopped"}
		lifecycle, err := insertRunEvent(ctx, tx, stoppedChild, RunEventCategoryProgress, EventStopRequested, "Child run stopped after parent stop", nil, metadata)
		if err != nil {
			return nil, err
		}
		final, err := insertRunEvent(ctx, tx, stoppedChild, RunEventCategoryFinal, EventRunStopped, "Child run stopped", nil, metadata)
		if err != nil {
			return nil, err
		}
		events = append(events, cancelled...)
		events = append(events, lifecycle, final)
	}
	return events, nil
}

func (r *PostgresRepository) SyncBuiltInPersonas(ctx context.Context, ident identity.LocalIdentity, configs []BuiltInPersonaConfig) (PersonaSyncResult, error) {
	if _, err := r.ensureUser(ctx, ident); err != nil {
		return PersonaSyncResult{}, err
	}
	if err := validateBuiltInPersonaConfigs(configs); err != nil {
		return PersonaSyncResult{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return PersonaSyncResult{}, err
	}
	defer tx.Rollback(ctx)
	result := PersonaSyncResult{Synced: len(configs)}
	for _, config := range configs {
		slug := strings.TrimSpace(config.Slug)
		personaID := ""
		err := tx.QueryRow(ctx, `select id from personas where slug=$1 and source='built_in'`, slug).Scan(&personaID)
		if errors.Is(err, pgx.ErrNoRows) {
			personaID = NewPersonaID()
			result.CreatedPersonas++
			if _, err := tx.Exec(ctx, `insert into personas (id, slug, name, description, source, is_default, is_active, active_version) values ($1, $2, $3, $4, 'built_in', $5, true, $6)`, personaID, slug, strings.TrimSpace(config.Name), strings.TrimSpace(config.Description), config.IsDefault, strings.TrimSpace(config.Version)); err != nil {
				return PersonaSyncResult{}, err
			}
		} else if err != nil {
			return PersonaSyncResult{}, err
		} else if _, err := tx.Exec(ctx, `update personas set name=$1, description=$2, is_default=$3, is_active=true, active_version=$4, updated_at=now() where id=$5`, strings.TrimSpace(config.Name), strings.TrimSpace(config.Description), config.IsDefault, strings.TrimSpace(config.Version), personaID); err != nil {
			return PersonaSyncResult{}, err
		}
		if config.IsDefault {
			result.DefaultPersonaSlug = slug
			if _, err := tx.Exec(ctx, `update personas set is_default=false, updated_at=now() where source='built_in' and id<>$1`, personaID); err != nil {
				return PersonaSyncResult{}, err
			}
		}
		var inserted bool
		err = tx.QueryRow(ctx, `insert into persona_versions (persona_id, version, system_prompt, model_route, allowed_tool_names, reasoning_mode, budget_summary) values ($1, $2, $3, $4, $5, $6, $7) on conflict (persona_id, version) do update set system_prompt=excluded.system_prompt, model_route=excluded.model_route, allowed_tool_names=excluded.allowed_tool_names, reasoning_mode=excluded.reasoning_mode, budget_summary=excluded.budget_summary where persona_versions.system_prompt is distinct from excluded.system_prompt or persona_versions.model_route is distinct from excluded.model_route or persona_versions.allowed_tool_names is distinct from excluded.allowed_tool_names or persona_versions.reasoning_mode is distinct from excluded.reasoning_mode or persona_versions.budget_summary is distinct from excluded.budget_summary returning xmax = 0`, personaID, strings.TrimSpace(config.Version), strings.TrimSpace(config.SystemPrompt), mustJSON(config.ModelRoute), config.AllowedToolNames, strings.TrimSpace(config.ReasoningMode), strings.TrimSpace(config.BudgetSummary)).Scan(&inserted)
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		if err != nil {
			return PersonaSyncResult{}, err
		}
		if inserted {
			result.CreatedVersions++
		}
		result.ActivatedVersions++
	}
	if result.DefaultPersonaSlug == "" {
		_ = tx.QueryRow(ctx, `select slug from personas where source='built_in' and is_default=true and is_active=true limit 1`).Scan(&result.DefaultPersonaSlug)
	}
	return result, tx.Commit(ctx)
}

func (r *PostgresRepository) ListPersonas(ctx context.Context, ident identity.LocalIdentity) ([]Persona, error) {
	if _, err := r.ensureUser(ctx, ident); err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select id, slug, name, description, source, is_default, is_active, active_version, created_at, updated_at from personas where is_active=true order by is_default desc, name asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var personas []Persona
	for rows.Next() {
		var persona Persona
		if err := rows.Scan(&persona.ID, &persona.Slug, &persona.Name, &persona.Description, &persona.Source, &persona.IsDefault, &persona.IsActive, &persona.ActiveVersion, &persona.CreatedAt, &persona.UpdatedAt); err != nil {
			return nil, err
		}
		personas = append(personas, persona)
	}
	return personas, rows.Err()
}

func (r *PostgresRepository) resolvePersonaSnapshotTx(ctx context.Context, tx pgx.Tx, threadPersonaID string, runPersonaID string) (PersonaSnapshot, error) {
	if personaID := strings.TrimSpace(runPersonaID); personaID != "" {
		return selectPersonaSnapshot(ctx, tx, personaID, PersonaResolvedFromRun)
	}
	if personaID := strings.TrimSpace(threadPersonaID); personaID != "" {
		return selectPersonaSnapshot(ctx, tx, personaID, PersonaResolvedFromThread)
	}
	var personaID string
	err := tx.QueryRow(ctx, `select id from personas where source='built_in' and is_default=true and is_active=true limit 1`).Scan(&personaID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PersonaSnapshot{}, nil
	}
	if err != nil {
		return PersonaSnapshot{}, err
	}
	return selectPersonaSnapshot(ctx, tx, personaID, PersonaResolvedFromDefault)
}

func selectPersonaSnapshot(ctx context.Context, tx pgx.Tx, personaID string, resolvedFrom PersonaResolvedFrom) (PersonaSnapshot, error) {
	var snapshot PersonaSnapshot
	var rawRoute []byte
	err := tx.QueryRow(ctx, `select p.id, p.slug, p.active_version, p.name, p.description, pv.system_prompt, pv.model_route, pv.allowed_tool_names, pv.reasoning_mode, pv.budget_summary from personas p join persona_versions pv on pv.persona_id=p.id and pv.version=p.active_version where p.id=$1 and p.is_active=true`, personaID).Scan(&snapshot.ID, &snapshot.Slug, &snapshot.Version, &snapshot.Name, &snapshot.Description, &snapshot.SystemPrompt, &rawRoute, &snapshot.AllowedToolNames, &snapshot.ReasoningMode, &snapshot.BudgetSummary)
	if errors.Is(err, pgx.ErrNoRows) {
		return PersonaSnapshot{}, NewError(CodeInvalidRequest, "Persona could not be resolved for this run.")
	}
	if err != nil {
		return PersonaSnapshot{}, err
	}
	if len(rawRoute) > 0 {
		_ = json.Unmarshal(rawRoute, &snapshot.ModelRoute)
	}
	snapshot.ResolvedFrom = resolvedFrom
	return snapshot, nil
}

func validatePersonaReferenceTx(ctx context.Context, tx pgx.Tx, personaID string) error {
	personaID = strings.TrimSpace(personaID)
	if personaID == "" {
		return nil
	}
	var exists int
	err := tx.QueryRow(ctx, `select 1 from personas p join persona_versions pv on pv.persona_id=p.id and pv.version=p.active_version where p.id=$1 and p.is_active=true`, personaID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return NewError(CodeInvalidRequest, "Persona could not be resolved for this thread.")
	}
	return err
}

func insertPersonaSnapshot(ctx context.Context, tx pgx.Tx, runID string, snapshot PersonaSnapshot) error {
	if snapshot.ID == "" {
		return nil
	}
	_, err := tx.Exec(ctx, `insert into run_persona_snapshots (run_id, persona_id, persona_slug, version, name, description, system_prompt, model_route, allowed_tool_names, reasoning_mode, budget_summary, resolved_from) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, runID, snapshot.ID, snapshot.Slug, snapshot.Version, snapshot.Name, snapshot.Description, snapshot.SystemPrompt, mustJSON(snapshot.ModelRoute), snapshot.AllowedToolNames, snapshot.ReasoningMode, snapshot.BudgetSummary, string(snapshot.ResolvedFrom))
	return err
}

func (r *PostgresRepository) getPersonaSnapshot(ctx context.Context, runID string) (PersonaSnapshot, error) {
	var snapshot PersonaSnapshot
	var rawRoute []byte
	err := r.Pool.QueryRow(ctx, `select persona_id, persona_slug, version, name, description, system_prompt, model_route, allowed_tool_names, reasoning_mode, budget_summary, resolved_from from run_persona_snapshots where run_id=$1`, runID).Scan(&snapshot.ID, &snapshot.Slug, &snapshot.Version, &snapshot.Name, &snapshot.Description, &snapshot.SystemPrompt, &rawRoute, &snapshot.AllowedToolNames, &snapshot.ReasoningMode, &snapshot.BudgetSummary, &snapshot.ResolvedFrom)
	if err != nil {
		return PersonaSnapshot{}, err
	}
	if len(rawRoute) > 0 {
		_ = json.Unmarshal(rawRoute, &snapshot.ModelRoute)
	}
	return snapshot, nil
}

func (r *PostgresRepository) ensureUser(ctx context.Context, ident identity.LocalIdentity) (User, error) {
	row := r.Pool.QueryRow(ctx, `insert into users (id, display_name) values ($1, $2) on conflict (id) do update set display_name=excluded.display_name, updated_at=users.updated_at returning id, display_name, created_at, updated_at`, ident.UserID, ident.DisplayName)
	var user User
	if err := row.Scan(&user.ID, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return User{}, err
	}
	return user, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanThread(row scanner) (Thread, error) {
	var thread Thread
	if err := row.Scan(&thread.ID, &thread.UserID, &thread.Title, &thread.Mode, &thread.LifecycleStatus, &thread.PersonaID, &thread.CreatedAt, &thread.UpdatedAt, &thread.ArchivedAt); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func scanRun(row scanner) (Run, error) {
	var run Run
	if err := row.Scan(&run.ID, &run.ThreadID, &run.UserID, &run.Status, &run.Source, &run.Title, &run.CreatedAt, &run.UpdatedAt, &run.CompletedAt, &run.StopRequestedAt, &run.ErrorCode, &run.ErrorMessage); err != nil {
		return Run{}, err
	}
	return run, nil
}

func scanBackgroundJob(row scanner) (BackgroundJob, error) {
	var job BackgroundJob
	var rawMetadata []byte
	if err := row.Scan(&job.ID, &job.RunID, &job.ThreadID, &job.UserID, &job.Kind, &job.Status, &job.Priority, &job.AttemptCount, &job.MaxAttempts, &job.ScheduledAt, &job.LeasedBy, &job.LeaseExpiresAt, &job.OwnershipVersion, &rawMetadata, &job.LastErrorCode, &job.LastError, &job.CreatedAt, &job.UpdatedAt); err != nil {
		return BackgroundJob{}, err
	}
	if len(rawMetadata) > 0 {
		_ = json.Unmarshal(rawMetadata, &job.Metadata)
	}
	if job.Metadata == nil {
		job.Metadata = map[string]any{}
	}
	return job, nil
}

func scanToolCall(row scanner) (ToolCall, error) {
	var call ToolCall
	var rawArguments []byte
	var rawResult []byte
	if err := row.Scan(&call.ID, &call.ThreadID, &call.RunID, &call.ToolCallID, &call.ToolName, &call.CandidateSchemaHash, &rawArguments, &call.ApprovalStatus, &call.ExecutionStatus, &rawResult, &call.ErrorCode, &call.ErrorMessage, &call.RequestedAt, &call.UpdatedAt); err != nil {
		return ToolCall{}, err
	}
	if len(rawArguments) > 0 {
		_ = json.Unmarshal(rawArguments, &call.ArgumentsSummary)
	}
	if len(rawResult) > 0 {
		_ = json.Unmarshal(rawResult, &call.ResultSummary)
	}
	if call.ArgumentsSummary == nil {
		call.ArgumentsSummary = map[string]any{}
	}
	return call, nil
}

func scanToolCallWithHash(row scanner, argumentsHash *string) (ToolCall, error) {
	var call ToolCall
	var rawArguments []byte
	var rawResult []byte
	if err := row.Scan(&call.ID, &call.ThreadID, &call.RunID, &call.ToolCallID, &call.ToolName, &call.CandidateSchemaHash, &rawArguments, argumentsHash, &call.ApprovalStatus, &call.ExecutionStatus, &rawResult, &call.ErrorCode, &call.ErrorMessage, &call.RequestedAt, &call.UpdatedAt); err != nil {
		return ToolCall{}, err
	}
	if len(rawArguments) > 0 {
		_ = json.Unmarshal(rawArguments, &call.ArgumentsSummary)
	}
	if len(rawResult) > 0 {
		_ = json.Unmarshal(rawResult, &call.ResultSummary)
	}
	if call.ArgumentsSummary == nil {
		call.ArgumentsSummary = map[string]any{}
	}
	return call, nil
}

func scanArtifact(row scanner) (Artifact, error) {
	var artifact Artifact
	if err := row.Scan(&artifact.ID, &artifact.ThreadID, &artifact.RunID, &artifact.Title, &artifact.ArtifactType, &artifact.Content, &artifact.ContentBytes, &artifact.TextExcerpt, &artifact.Truncated, &artifact.CreatedAt, &artifact.UpdatedAt); err != nil {
		return Artifact{}, err
	}
	return artifact, nil
}

func scanAgentTask(row scanner) (AgentTask, error) {
	var task AgentTask
	if err := row.Scan(&task.ID, &task.ThreadID, &task.RunID, &task.Role, &task.Goal, &task.Status, &task.ResultSummary, &task.ChildThreadID, &task.ChildRunID, &task.ParentToolCallID, &task.DelegatedAt, &task.CreatedAt, &task.UpdatedAt); err != nil {
		return AgentTask{}, err
	}
	return task, nil
}

func scanContextSource(row scanner) (ContextSource, error) {
	var source ContextSource
	var rawMetadata []byte
	if err := row.Scan(&source.ID, &source.ThreadID, &source.UserID, &source.Kind, &source.Title, &source.Locator, &source.Summary, &source.Status, &rawMetadata, &source.CreatedAt, &source.UpdatedAt); err != nil {
		return ContextSource{}, err
	}
	if len(rawMetadata) > 0 {
		_ = json.Unmarshal(rawMetadata, &source.Metadata)
	}
	if source.Metadata == nil {
		source.Metadata = map[string]any{}
	}
	return source, nil
}

func scanMemoryEntry(row scanner) (MemoryEntry, error) {
	var entry MemoryEntry
	if err := row.Scan(&entry.ID, &entry.UserID, &entry.ScopeType, &entry.ScopeID, &entry.Title, &entry.Summary, &entry.Content, &entry.Status, &entry.SafetyState, &entry.SourceThreadID, &entry.SourceRunID, &entry.SourceEventID, &entry.ContentHash, &entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt, &entry.DeletedBy, &entry.DeleteReason); err != nil {
		return MemoryEntry{}, err
	}
	return entry, nil
}

func scanMemoryProposal(row scanner) (MemoryWriteProposal, error) {
	var proposal MemoryWriteProposal
	if err := row.Scan(&proposal.ID, &proposal.UserID, &proposal.ScopeType, &proposal.ScopeID, &proposal.Title, &proposal.Summary, &proposal.Content, &proposal.Status, &proposal.SafetyState, &proposal.SourceThreadID, &proposal.SourceRunID, &proposal.SourceEventID, &proposal.IdempotencyKey, &proposal.CreatedEntryID, &proposal.CreatedAt, &proposal.DecidedAt, &proposal.DecidedBy, &proposal.DecisionReason); err != nil {
		return MemoryWriteProposal{}, err
	}
	return proposal, nil
}

func intPlaceholder(index int) string {
	return fmt.Sprintf("%d", index)
}

func scanRunEvent(row scanner) (RunEvent, error) {
	var event RunEvent
	var rawMetadata []byte
	if err := row.Scan(&event.ID, &event.RunID, &event.ThreadID, &event.UserID, &event.Sequence, &event.Category, &event.Type, &event.Summary, &event.Content, &rawMetadata, &event.CreatedAt); err != nil {
		return RunEvent{}, err
	}
	if len(rawMetadata) > 0 {
		_ = json.Unmarshal(rawMetadata, &event.Metadata)
	}
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}
	return event, nil
}

func scanRunEvents(rows pgx.Rows) ([]RunEvent, error) {
	events := []RunEvent{}
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func mustJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

func scanMessage(row scanner) (Message, error) {
	var message Message
	var rawMetadata []byte
	if err := row.Scan(&message.ID, &message.ThreadID, &message.UserID, &message.Role, &message.Content, &rawMetadata, &message.ClientMessageID, &message.CreatedAt); err != nil {
		return Message{}, err
	}
	if len(rawMetadata) > 0 {
		_ = json.Unmarshal(rawMetadata, &message.Metadata)
	}
	if message.Metadata == nil {
		message.Metadata = map[string]any{}
	}
	if message.ClientMessageID != nil {
		trimmed := strings.TrimSpace(*message.ClientMessageID)
		message.ClientMessageID = &trimmed
	}
	return message, nil
}
