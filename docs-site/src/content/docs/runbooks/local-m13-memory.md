---
title: M13 Memory 本地 Smoke
description: M13/M13.5 memory migration、real PG/httpapi smoke、Settings > Memory 浏览器 smoke 和验证命令。
---

## 环境变量

```bash
APP_ENV=local
HTTP_ADDR=127.0.0.1:8080
DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
LOOMI_TEST_DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080
```

## 启动 Postgres 并应用 migration

```bash
docker compose up -d postgres
export DATABASE_URL=postgres://loomi:loomi@127.0.0.1:55433/loomi_m2?sslmode=disable
migrate -path migrations -database "$DATABASE_URL" up
migrate -path migrations -database "$DATABASE_URL" version
```

M14 blocker foundation 需要 version `10` 且 clean；M42 memory provider foundation 需要 version `14` 且 clean；M47 provider config details 需要 version `15` 且 clean。API 不会在启动时自动执行 migration。

## Real PG/httpapi smoke

```bash
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/httpapi -run TestM13MemoryRealPGHTTPAPISmoke -count=1 -v
```

这个 smoke 使用真实 Postgres repository 和 HTTP handlers，覆盖：

- create/propose memory write
- approve proposal
- approved memory list/search
- RunContext safe memory snapshot
- delete 后 tombstone 并立即从 list/search/RunContext 排除
- duplicate approve/deny/delete 不重复 entry/audit
- out-of-scope 不泄露存在性
- terminal run 后 proposal/approve/deny/delete audit 仍可从 `/v1/memory/audit` 查询
- sensitive content 不进入 API response、RunContext safe summary、memory audit metadata

## M14 blocker foundation checks

```bash
LOOMI_TEST_DATABASE_URL="$DATABASE_URL" go test ./internal/productdata -run TestPostgresMemoryEntryScopeAndTerminalAudit -count=1 -v
```

这个 smoke 使用真实 Postgres repository，覆盖同一用户下 thread A 不能 read/delete thread B memory、thread A scope 可 read/delete、terminal run 后 approve/deny/delete audit 仍存在，以及重复 deny/delete 不重复 audit。

## M42 provider foundation checks

```bash
go test ./internal/productdata -run 'TestMemoryProviderStatusDefaultsFallbackAndRedaction|TestPrepareRunContextIncludesMemoryProviderReadiness' -count=1 -v
go test ./internal/httpapi -run TestMemoryProviderHandlersGetAndUpdateSafeStatus -count=1 -v
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/memory.test.ts
bun run --cwd web build
```

这些检查覆盖默认 local provider、semantic 未配置、未知 provider 降级、HTTP GET/PUT、安全 run readiness summary、Settings > Memory provider 面板和前端 API 映射。

## M47 provider config detail checks

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestMemoryProvider|TestPrepareRunContextIncludesMemoryProviderReadiness' -count=1
bun test --cwd web src/memory.test.ts src/components/SettingsView.runtime.test.tsx
bun run --cwd web build
```

这些检查覆盖 OpenViking/Nowledge 配置状态、key 只写不回显、Settings > Memory 的 provider 选择和模型字段，以及 legacy semantic provider 的兼容状态。该 slice 不执行外部 provider read/write adapter。

## M43 memory tool checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestMemoryTool|TestWorkerExecutesApprovedMemory|TestGatewayExposesCodeAgentToolsToProvider' -count=1
bun test --cwd web src/components/SettingsView.tools.test.tsx
```

这些检查覆盖 `memory.search`、`memory.read`、`memory.write`、`memory.forget`、`memory.status` 的目录、参数校验、provider schema、ToolBroker/worker continuation、safe result summary、Settings > Tools 展示。该 slice 不包含自动 distill、embedding/vector search 或外部 semantic provider 执行。

## M48 memory tool parity checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestMemoryTool|TestGatewayExposesCodeAgentToolsToProvider|TestMemoryToolDefinitions|TestWorkerExecutesApprovedMemory' -count=1
bun test --cwd web src/components/SettingsView.tools.test.tsx src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts
bun run --cwd web build
```

这些检查覆盖扩展后的 `memory.list`、`memory.edit`、`memory.context`、`memory.timeline`、`memory.connections`、`memory.thread_search`、`memory.thread_fetch`，以及原有 memory tools 的目录、校验、provider schema、执行器和 Settings > Tools 展示。该 slice 仍然只使用本地 productdata 安全摘要，不执行外部 OpenViking/Nowledge adapter。

## M49 snapshot/impression checks

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestMemory.*Snapshot|TestMemorySnapshot|TestMemoryOverviewAndImpression|TestMemoryProvider|TestPrepareRunContextIncludesMemoryProviderReadiness|TestPrepareRunContextIncludesSafeMemorySnapshot' -count=1
bun test --cwd web src/memory.test.ts src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts
bun run --cwd web build
```

这些检查覆盖 `/v1/memory/snapshot`、`/v1/memory/snapshot/rebuild`、`/v1/memory/impression`、`/v1/memory/impression/rebuild`，以及 Settings > Memory 的记忆画像、记忆快照和重建动作。该 slice 仍然只使用本地 approved memory 安全摘要，不执行外部 OpenViking/Nowledge adapter。

## M50 memory content view checks

```bash
go test ./internal/httpapi -run TestMemorySnapshotAndImpressionHandlers -count=1
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts
bun run --cwd web build
```

这些检查覆盖 `GET /v1/memory/content?uri=memory://...&layer=read`、Settings > Memory 快照命中点击打开安全内容弹窗、以及不暴露 raw content、content hash、provider trace 或 secret-like 文本。

## M51 manual memory add checks

```bash
go test ./internal/httpapi -run 'TestMemoryHandlersCreateManualEntry|TestMemorySnapshotAndImpressionHandlers' -count=1
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts
bun run --cwd web build
```

这些检查覆盖 `POST /v1/memory/entries`、Settings > Memory 手动添加记忆、添加后刷新列表和快照、以及响应不暴露 raw content/content hash。

## M52 recent memory errors checks

```bash
go test ./internal/httpapi -run 'TestMemoryErrorsReportsProviderDiagnostic|TestMemoryHandlersCreateManualEntry' -count=1
bun test --cwd web src/memory.test.ts src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts
bun run --cwd web build
```

这些检查覆盖 `/v1/memory/errors` 和 Settings > Memory provider 面板中的近期异常区。错误项只来自安全 provider diagnostic，不包含 key、Authorization header 或上游原始 trace。

## M53 Nowledge local detect checks

```bash
go test ./internal/httpapi -run 'TestMemoryNowledgeDetectSafeMiss|TestMemoryErrorsReportsProviderDiagnostic' -count=1
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts
bun run --cwd web build
```

这些检查覆盖 `/v1/memory/provider/nowledge/detect` 和 Settings > Memory 的“检测本地实例”按钮。检测只访问 `127.0.0.1:14242/health`，不会扫描远端地址或读取 API key。

## M54 provider-aware memory tool checks

```bash
go test ./internal/productdata -run 'TestToolCatalogHidesNowledgeUnsupportedMemoryEdit|TestNowledgeRunContextFiltersUnsupportedMemoryEdit|TestToolCatalogIncludesMemoryRuntimeTools' -count=1
bun run --cwd web build
```

这些检查覆盖 Nowledge 下 `memory.edit` 在 Settings > Tools 中禁用、在 prepared RunContext 中被移除，以及本地/OpenViking 仍保留完整 memory tool set。禁用原因只使用安全 reason code。

## M55 memory provider config modal checks

```bash
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts
bun run --cwd web build
```

浏览器 smoke 需要打开 Settings > Memory，选择 Nowledge，点击“配置”，确认弹层出现，再点击“检测本地实例”并看到安全提示。provider key 字段仍为写入控件，不从状态回显原始 secret。

## M56 provider card selection checks

```bash
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts
bun run --cwd web build
```

浏览器 smoke 复用 M55 路径，但 Nowledge 选择必须通过 provider 卡片完成。小屏布局应退为单列，避免按钮文字挤压。

## M57 notebook tool checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestValidateMemory|TestToolCatalogIncludesMemory|TestMemoryToolsAreAvailable|TestMemoryToolDefinitions|TestMemoryToolExecutorNotebookLifecycle|TestGatewayExposesCodeAgentToolsToProvider' -count=1
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/mockApiClient.test.ts src/memory.test.ts
bun run --cwd web build
```

这些检查覆盖 `notebook.read`、`notebook.write`、`notebook.edit`、`notebook.forget` 的参数校验、目录展示、RunContext 可用性、provider schema 映射和 runtime 生命周期。Notebook 条目复用 memory store，并以 `source_type=notebook` 区分。

## M58 prompt notebook snapshot checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestPrepareRunContextIncludesNotebookSnapshot|TestPrepareRunContextIncludesSafeMemorySnapshot|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
```

这些检查覆盖 `RunContext.NotebookSnapshot` 以及 provider system prompt 中的 `<memory>` / `<notebook>` 安全摘要块。Prompt 注入只使用 safe summary，不包含 raw content、content hash、provider trace、tool output 或 secret-like 文本。

## M59 semantic/notebook snapshot separation checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestMemoryOverviewAndImpressionSnapshotsAreSafe|TestPrepareRunContextIncludesNotebookSnapshot|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
```

这些检查覆盖 Notebook 条目不会出现在 `/v1/memory/snapshot` 和 `/v1/memory/impression` 的语义投影中，同时仍通过 `RunContext.NotebookSnapshot` 注入 `<notebook>`。

## M60 external provider read adapter checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestMemoryProviderStatusDefaultsFallbackAndRedaction|TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider' -count=1
```

这些检查使用本地 `httptest` 服务模拟 OpenViking 和 Nowledge，覆盖 `memory.search` / `memory.read` 的 provider 路由、认证 header、provider URI 和 safe-summary-only 结果。验证不会调用真实外部服务，也不会执行 provider write。

## M61 external provider write adapter checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider' -count=1
```

这些检查继续使用本地 `httptest` 服务模拟 provider，覆盖 OpenViking `memory.write` / `memory.edit` / `memory.forget` 和 Nowledge `memory.write` / `memory.forget`。真实外部 provider 不会在测试中被调用。

## M62 Nowledge rich adapter checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesNowledgeProvider' -count=1
```

这个检查使用本地 `httptest` 服务模拟 Nowledge，覆盖 `memory.connections`、`memory.timeline`、`memory.thread_search`、`memory.thread_fetch` 的 provider 路由和 safe excerpt 结果。

## M63 external post-run commit checks

```bash
go test ./internal/productdata ./internal/runtime -run 'TestPostRunMemory|TestWorkerProposesPostRunMemoryWhenCommitAfterRunEnabled' -count=1
```

这些检查覆盖 local provider 仍创建 pending proposal，外部 provider 通过 write adapter 提交 post-run outcome，并使用 `memory_provider_commit_completed` 做终端 run 后的幂等记录。

## M64 OpenViking local detect checks

```bash
go test ./internal/httpapi -run 'TestMemoryOpenVikingDetectSafeMiss|TestMemoryNowledgeDetectSafeMiss' -count=1
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/App.settings.test.tsx src/realApiClient.test.ts
bun run --cwd web build
```

这些检查覆盖 `/v1/memory/provider/openviking/detect` 和 Settings > Memory 的 OpenViking “检测本地实例”按钮。检测只访问 `127.0.0.1:8282` 的 OpenViking API，不读取、不发送、不返回任何 provider key。

## M65 OpenViking connections checks

```bash
go test ./internal/runtime -run TestMemoryToolExecutorSearchesOpenVikingProvider -count=1
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestMemoryToolExecutorSearchReadStatusWriteAndForget|TestMemoryToolExecutorNotebookLifecycle' -count=1
```

这些检查覆盖 OpenViking 配置下 `memory.connections` 对 `viking://...` URI 调用 `/api/v1/fs/ls`，并只返回 bounded safe child resource summaries。

## M66 external prompt snapshot checks

```bash
go test ./internal/runtime -run 'TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesOpenVikingProvider|TestMemoryToolExecutorSearchesNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
```

这些检查覆盖 Gateway 初始模型请求前用最新 user message 查询外部 provider，并把 safe hits 注入 `<memory>` prompt block。provider failure 或无结果不阻断 run。

## M67 external prompt snapshot event checks

```bash
go test ./internal/runtime -run TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider -count=1
```

这个检查覆盖 `memory_external_snapshot_loaded` progress event，确保外部 provider prompt recall 对时间线可见，且 metadata 不包含 query、raw content、provider trace 或 secret。

## M68 Nowledge external prompt snapshot checks

```bash
go test ./internal/runtime -run 'TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestGatewayEnrichesPromptMemorySnapshotFromNowledgeProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
go test ./internal/productdata ./internal/runtime -run 'TestMemoryToolExecutorSearchesNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromNowledgeProvider|TestGatewayEnrichesPromptMemorySnapshotFromExternalProvider|TestRunSystemPromptIncludesSafeMemoryAndNotebookSnapshots' -count=1
```

这些检查覆盖 Nowledge 配置下 Gateway 初始模型请求前调用 `/memories/search`，把 safe hits 注入 `<memory>`，并写入 `provider=nowledge` 的 `memory_external_snapshot_loaded` progress event。事件 metadata 不包含 query、raw content、provider trace 或 secret。

## M69 memory provider runtime error checks

```bash
go test ./internal/runtime ./internal/httpapi -run 'TestGatewayRecordsExternalMemorySnapshotFailureForRecentErrors|TestMemoryErrorsReportsRuntimeProviderFailures|TestMemoryErrorsReportsProviderDiagnostic' -count=1
bun test --cwd web src/memory.test.ts
```

这些检查覆盖外部 prompt recall 失败时写入 `memory_external_snapshot_failed`，run 继续使用原始 context，并且 `/v1/memory/errors` 同时返回配置诊断和 runtime provider failure。错误项只包含 safe code/message/provider/state/run/event 信息，不包含 query、raw content、provider trace 或 secret。

## M70 memory provider error UI checks

```bash
bun test --cwd web src/components/SettingsView.runtime.test.tsx src/memory.test.ts
```

这些检查覆盖 Settings > Memory 近期异常区会渲染 runtime error 的 `eventType` 和 `runId`，并保持现有 provider diagnostic 展示路径。

## M44 post-run proposal checks

```bash
go test ./internal/runtime -run 'TestWorkerProposesPostRunMemory|TestPostRunMemory' -count=1
bun test --cwd web src/components/SettingsView.runtime.test.tsx
```

这些检查覆盖 Settings > Memory 的每轮后整理文案、`commit_after_run` 开关打开后 completed run 生成 pending write proposal、默认关闭不生成、同一 run 重试不重复生成。该 slice 不自动 approve，不让 pending proposal 进入 memory search，也不执行 embedding/vector search。

## M45 proposal review checks

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestListMemoryWriteProposals|TestMemoryHandlersListPendingWriteProposals' -count=1
bun test --cwd web src/memory.test.ts src/components/MemoryPanel.test.tsx
```

这些检查覆盖 pending proposal 安全列表、`GET /v1/memory/write-proposals`、Settings > Memory 的待审批区、保存/拒绝动作入口、以及不暴露 raw content/idempotency key。

## M46 proposal edit checks

```bash
go test ./internal/productdata ./internal/httpapi -run 'TestUpdateMemoryWriteProposal|TestMemoryHandlersUpdateWriteProposal' -count=1
bun test --cwd web src/memory.test.ts src/components/MemoryPanel.test.tsx
```

这些检查覆盖 pending proposal 的标题/摘要编辑、`PATCH /v1/memory/write-proposals/{proposal_id}`、审批后使用编辑后的记忆文本、已审批 proposal 不能继续编辑、以及 Settings > Memory 的编辑入口。

## Settings > Memory 浏览器 smoke

M14 full done gate uses seeded memory entries and real API mode. It must cover list/search/filter, detail drawer or modal, delete confirmation, error/empty/loading/tombstoned states, and real audit history. Do not mock audit history in the UI.

启动 API：

```bash
APP_ENV=local HTTP_ADDR=127.0.0.1:8080 DATABASE_URL="$DATABASE_URL" go run ./cmd/loomi-api
```

启动真实 API 模式 web：

```bash
VITE_LOOMI_API_BASE_URL=http://127.0.0.1:8080 bun run --cwd web dev --host 127.0.0.1
```

浏览器检查：

- 打开 web shell。
- 进入 Settings > Memory。
- 验证顶部 Memory Service 面板来自 `/v1/memory/provider`，显示 enablement、provider、state、configured、diagnostic，并且刷新按钮可触发状态读取。
- 验证 Memory Snapshot 和 Memory Impression 卡片来自 `/v1/memory/snapshot` 与 `/v1/memory/impression`，重建按钮可触发 rebuild，并且页面不显示 raw content、provider trace、secret-like 文本。
- 点击 Memory Snapshot 的命中标签，验证弹窗显示安全标题/摘要，并且关闭后页面状态正常。
- 在新增记忆表单输入标题和内容，点击添加，验证条目进入已保存记忆并刷新快照命中。
- 切换到未配置 provider 时，验证近期异常区出现安全诊断；恢复已配置 provider 后异常清空。
- 选择 Nowledge，点击检测本地实例；如果本地服务存在，应填入 `http://127.0.0.1:14242`，否则显示安全未检测到信息。
- 选择 OpenViking，点击检测本地实例；如果本地服务存在，应填入 `http://127.0.0.1:8282`，否则显示安全未检测到信息。
- 验证 Memory category 可打开。
- 验证 list/search/delete 状态连接真实 API，删除后条目消失。
- 验证默认界面只展示搜索、记忆列表、简洁历史和空状态；高级筛选折叠，系统快照事件不刷屏。
- 验证待审批记忆可进入编辑状态，修改标题/摘要后返回待审批卡片，并且仍可保存或拒绝。
- M14 full smoke additionally verifies detail, filters, delete confirmation, and audit history from `/v1/memory/audit`.
- DevTools console 无红色错误。

2026-05-25 M14 candidate smoke evidence:

- API ran on `127.0.0.1:18080` because `8080` was already occupied.
- Web ran on `127.0.0.1:5173` with `VITE_LOOMI_API_BASE_URL=http://127.0.0.1:18080`.
- Seed used thread `thr_1779701971994970000_b2ab6b56d4c5`, run `run_1779701972024596000_bad6f6217636`, search `m14-smoke-173931`, kept memory `mem_1779701972093540000_89e109af7e29`, and deleted memory `mem_1779702112854813000_f094a9187f91`.
- Browser verified scoped list/search/filter, detail, delete confirmation, post-delete list refresh, and real audit history showing `memory_deleted`, `memory_write_proposed`, `memory_write_approved`, `memory_write_denied`, and `memory_snapshot_loaded`.
- Console errors: none after the frontend delete request body stopped sending UI-only `limit`.

如果本地端口被占用导致无法启动其中一个服务，记录端口/进程原因，并用 real PG/httpapi smoke、`bun test --cwd web`、`bun run --cwd web build` 作为等价后端与前端边界证据。

## Validation commands

```bash
go test ./...
bun test --cwd web
bun run --cwd web build
bun run --cwd docs-site build
git diff --check
```

## Deferred

M13/M13.5/M14/M42/M43/M44/M45/M46 不包含 vector DB、embedding、RAG、外部语义 memory adapter、LLM distill worker、activity recorder、sandbox、MCP rewrite、worker/job queue rewrite 或多 agent 自动记忆。
