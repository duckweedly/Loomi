---
title: M98 Agent Runtime Parity Hardening
description: Parallel tool calls, provider retry, provider context compaction, and event replay reduction.
---

## Completed

- Fixed OpenAI-compatible streaming tool-call flushing so one model turn can emit multiple indexed tool calls without dropping later calls.
- Enabled parallel tool calls for the Local Codex OAuth Responses bridge and kept parsing streamed function calls until `response.completed`, so Local Codex can emit more than one tool call in a model turn.
- Completed Local Codex Responses turns can now use final `response.completed.response.output` text when no text deltas were streamed, and `[DONE]` terminates text-delta-only streams as a normal completion.
- Treated OpenAI-compatible `finish_reason=length` as an incomplete provider response instead of a completed assistant answer.
- Sent enabled tool schemas to Anthropic and Gemini provider requests, not just OpenAI-compatible and Local Codex requests.
- Treated Anthropic `max_tokens`, Gemini `MAX_TOKENS`, and Local Codex Responses `response.incomplete` as `provider_incomplete` instead of generic provider errors or completed answers.
- Stopped OpenAI-compatible SSE parsing immediately after terminal tool-call or completion frames, so trailing transport errors after a complete tool-call turn do not fail an otherwise valid run.
- Preserved Anthropic and Gemini tool-call ids/arguments across multiple tool calls in one model turn, with stable generated ids where the provider does not return one.
- Accumulated Anthropic streamed `input_json_delta` chunks before emitting the tool call so real provider streams keep complete argument summaries.
- Serialized OpenAI-compatible continuation history for multiple completed tool calls as one assistant `tool_calls` message followed by matching tool results.
- Allowed multiple auto-approved read-only tool calls to be recorded in one run while keeping approval-gated and write-capable tools serialized.
- Updated the queued runner to execute a ready auto-approved batch concurrently before provider continuation, including batches requested by a later continuation turn, so continuation sees all matching tool results together.
- Ordered continuation tool-call/result messages by the original provider request order so concurrent execution timing does not scramble model context.
- Rejected duplicate same-argument workspace read/list/grep requests inside one provider turn before the buffered continuation batch is recorded.
- Allowed multiple approval-required tool calls to remain blocked together in one run, so a model turn that asks for several gated actions is fully observable instead of failing on the second request.
- Kept mixed provider batches coherent: if one auto-approved tool succeeds while another tool call in the same model turn is still blocked for approval, the queued runner leaves the run blocked instead of continuing the model with a partial batch.
- Added bounded provider retry for retryable failures before visible output or tool state exists, with contract coverage for rate limits, timeouts, empty attempts, stream errors, and retry exhaustion.
- Added real HTTPProvider retry smoke coverage for 429, 408, 504, and pre-output temporary network errors, proving the gateway retries the actual transport boundary and not only fake provider events.
- Added safe registered context sources to the provider system prompt while excluding workspace-path sources.
- Added deterministic provider-context compaction with `context_compacted` telemetry, recent-message windows, per-message caps, tool-call/tool-result pair preservation, and latest-user preservation after tool history.
- Reused prepared enabled-tool snapshots for initial provider tool schemas and scoped tool-call validation. Continuation now uses the run-step projection for provider messages and continuation tool schemas instead of replaying the full event stream on the hot path.
- Added a rebuildable `run_step_state_projections` checkpoint so queued-runner next-action decisions no longer need to replay all run events on every loop.
- Extended the checkpoint with run route metadata and workspace guardrail facts so continuation hydration and repeated workspace-tool checks can avoid full event replay.
- Made projection catch-up and rebuild writes fail visibly instead of returning an unmaterialized state after a database write error.
- Guarded projection catch-up/rebuild writes so they only advance `last_sequence`; stale readers re-read a newer checkpoint instead of writing an older state over it.
- Switched Work-mode runtime todo snapshots to derive from the run-step projection instead of scanning the whole event stream after each tool state change.
- Rebuilt approved-tool resume `RunContext` from the projection for built-in and MCP tool chains by storing MCP candidate schema hashes in `RunStepState`, including safe MCP availability for resumed contexts.
- Rebuilt initial model-gateway `RunContext` from the same projection when route metadata is present, so normal prepare-context jobs no longer need to read the full run event stream just to recover `run_created` metadata.
- Stored MCP discovery availability summaries in `RunStepState`, preserving safe MCP visibility/diagnostic metadata when prepare-context hydrates from the projection instead of events.
- Skipped PostgreSQL run-step projection upserts inside transactions when no newer run events exist, avoiding unnecessary projection row writes and locks on hot lifecycle paths.
- Removed prepare-context, gateway, and queued-runner continuation fallbacks that replayed run events from sequence 0 when the run-step projection could not provide context; those paths now fail or skip optional guardrails explicitly instead of silently reintroducing O(n²) replay.
- Added worker-side PostgreSQL projection ensure for the claimed run before runtime execution, so missing or semantically corrupt projection rows are repaired for the run being processed without broad historical backfill blocking current jobs.
- Split continuation request start from continuation output in the projection so a worker restart can resume a start-only continuation without suppressing the pending model resume.
- Added an atomic queued-runner continuation claim keyed by completed tool result and `job_id`, with a short claim lease so another job cannot invoke provider continuation for the same frontier until the prior claim expires.
- Moved the PostgreSQL continuation claim decision onto the run-step projection catch-up path instead of rebuilding the full run event stream inside the claim transaction.
- Moved tool lifecycle loop metadata to the run-step projection, removing the remaining full event-stream scan from request/approve/execute/success/failure/recovery metadata writes.
- Made direct gateway `RunAsync` stop-aware while keeping it independent from the caller request context.
- Removed one redundant worker-side full run-event replay after history publication.
- Removed the remaining worker claim-time full replay used for live publication; the worker now uses the claimed run-step projection cursor to publish only the just-claimed frontier.
- Removed stale gateway helper fallbacks that could still replay a full run event stream for scoped tool validation when a prepared context or run-step projection is already available.
- Serialized Postgres run-event sequence allocation per run and moved memory-audit timeline mirrors into the same transaction so parallel tool/runtime writes cannot race into duplicate `(run_id, sequence)` values or silently lose mirrored events.
- Moved Postgres memory proposal/create/approve/deny/delete audit writes into the same transaction as the memory mutation, so audit insert failures roll back the proposal, entry, decision, or tombstone instead of leaving a half-committed memory state.
- Replaced external post-run memory commit idempotency replay with a scoped event-type existence query, so long completed runs do not reread their entire event stream before committing memory.
- Added worker lease heartbeat renewal while a runner is active, cancelling the runner if renewal fails or ownership is lost.
- Recovered expired leases by resetting in-flight `executing` tool calls back to approved/not-started for retry, and marked them failed when retry exhaustion fails the run.
- Reused one run-step projection snapshot while recording multiple executing-tool recovery events for the same run, avoiding repeated projection lookups during stale lease recovery.
- Made Postgres worker failure persistence atomic across the owned job, terminal run state, `job_attempt_failed`, and final `run_failed`, and made the worker return a persistence error instead of swallowing it when failure recording cannot be saved.
- Rejected deny-after-approve tool decisions, including the approved/not-started window, so an already approved queued resume cannot be flipped into a stopped run by a later conflicting decision.
- Added a worker stop watcher so `StopRun` cancels an active runner promptly instead of waiting for the next lease heartbeat.
- Added SSE persisted-event backfill from the last sent sequence so dropped live-buffer events and terminal close markers are recovered without waiting for frontend full replay.
- Added SSE comment heartbeats while an active stream is idle so clients and proxies keep long-running run streams open.
- Switched active-run frontend reconciliation to cursor-based event refresh and treated SSE EOF without `stream_closed` as a recoverable disconnect.
- Published tool execution start/success/failure events from queued tool execution so live timelines show long-running tools moving without refresh.
- Preserved continuation system prompts after tool results so persona, mode, memory/notebook, workspace, and safety policies remain present in later model calls.
- Allowed enabled `runtime.get_current_time` requests during bounded continuation, preserving approval and loop-limit boundaries instead of failing the run as `unsupported_tool_loop`.
- Brought memory/notebook tools into the same enabled-tool snapshot boundary as other gateway tools: unenabled memory requests are rejected before recording a tool call, while enabled memory/notebook tools can continue through the bounded tool loop.
- Made continuation tool turns atomic against the remaining loop budget: if one provider response asks for more tools than the run can still accept, Loomi fails the turn without leaving a partially recorded approved tool call behind.
- Raised the bounded continuation budget from 6 to 24 accepted tool calls and widened Work-mode survey intent detection for read/check/summarize/explain page-source tasks, so project walkthroughs do not stop halfway after a few reads.
- Persisted recoverable tool execution errors as failed tool results that resume provider continuation, so a missing guessed path can be corrected by the model instead of failing the whole run; permission and unbound-workspace failures remain terminal.
- Added retry exhaustion telemetry coverage so final provider failures keep provider route/model/attempt metadata while redacting keys, paths, and provider traces.
- Added repository contracts for recording multiple auto-approved tool calls in one model turn, including an optional Postgres contract when `LOOMI_TEST_DATABASE_URL` is available.
- Tightened product-data duplicate `tool_call_id` idempotency: exact retries return the existing projection, but conflicting tool names, schemas, states, or argument identities are rejected before writing events.
- Cancelled every unresolved sibling tool call when a run is stopped or a pending approval is denied, and mapped `tool_call_cancelled` into the run-step projection so terminal runs do not keep hidden pending/executing tools.
- Extended the same cancellation rule to delegated child runs stopped by a parent stop, so child-side pending approvals or executing tools cannot survive the parent terminal state.
- Preserved the parent run workspace-root snapshot when `agent.delegate` child-run reconciliation queues the parent continuation job.
- Added optional PostgreSQL delegate reconciliation smoke coverage for `agent.delegate -> child run completed -> parent tool succeeded -> parent resume job queued`, gated by `LOOMI_TEST_DATABASE_URL`.
- Added optional PostgreSQL worker smoke coverage for the full approved `agent.delegate` path through `QueuedRunRouter`, proving the parent waits while the child run is active and resumes only after reconciliation.
- Added optional PostgreSQL worker/provider smoke coverage for `agent.delegate -> child worker terminal -> reconciliation -> parent provider continuation`, proving the handoff path works without hand-written child completion events.
- Stabilized workspace repeat guard keys as `tool_name + arguments_hash` and extended same-turn repeat detection to `workspace.glob` and `workspace.tree_summary`.
- Serialized Anthropic and Gemini continuation histories with native tool-use/tool-result shapes instead of dropping provider-neutral `assistant_tool_call` / `tool_result` messages as empty text.
- Made Gemini generated tool-call ids unique across the whole SSE stream, avoiding duplicate ids when separate frames each contain a function call.
- Added a queued-runner execution-state gate after `StartToolCallExecution`, so a stop/deny race that returns `cancelled` does not enter `ToolBroker` or write a spurious failure.
- Added a queued-runner owner fence before persisting tool success or failure, so stale workers drop completed tool results after losing their claimed job lease.
- Made Postgres queued-job claim skip terminal or stop-requested runs in the claim query, matching the in-memory queue behavior that keeps looking for a runnable job in the same poll.
- Aligned the web-search provider plan with the implemented read-only auto-approved policy.
- Added an HTTP-level code-agent smoke that proves one run can combine same-turn parallel workspace reads, continuation, approval-gated `agent.spawn`, approval-gated `agent.delegate`, a real queued child model run, parent reconciliation, and final parent completion.
- Extended Work-mode intent detection so English delegation requests such as `delegate` / `subagent` / `child agent` keep agent tools in the prepared enabled-tool snapshot alongside workspace tools.
- Bounded aggregate tool-result payloads after JSON serialization, so long arrays of small result objects compact into `truncated` metadata instead of bloating provider continuation context.
- Bounded provider assistant tool-call argument summaries during context compaction, preserving the latest tool-call/result pair instead of dropping it when large arguments such as patch bodies exceed the message budget.

## Validation

```bash
go test ./internal/runtime ./internal/productdata
go test ./internal/httpapi -run 'TestRunEventStreamBackfillsDroppedLiveEvents|TestRunEventStreamSubscribesBeforeHistoryRead|TestRunEventStreamDeliversHistoryBeforeCloseMarker' -count=1
go test ./internal/httpapi -run 'TestRunEventStreamSendsHeartbeatWhileIdle|TestRunEventStreamBackfillsDroppedLiveEvents|TestRunEventStreamFlushesHistoryAndCloseMarker' -count=1
go test ./internal/runtime -run 'TestWorkerRenewsLeaseWhileRunnerIsStillRunning|TestWorkerCancelsRunnerSoonAfterStopRun|TestQueuedRunRouterExecutesParallelAutoApprovedToolsBeforeContinuation|TestQueuedRunRouterRunsReadyAutoApprovedToolsConcurrently|TestQueuedRunRouterDoesNotContinueUntilAllParallelToolCallsResolved|TestQueuedRunRouterDrainsParallelAutoApprovedToolsAfterContinuation|TestGatewayRecordsMultipleApprovalRequiredToolCalls|TestGatewayRedactsRetryTelemetryAndExhaustionFailure' -count=1
go test ./internal/runtime -run TestGatewayRetriesHTTPProviderTransientFailuresBeforeOutput -count=1
go test ./internal/runtime -run 'TestWorkerDoesNotReplayFullRunHistoryOnClaim|TestWorkerEnsuresClaimedRunStepStateProjectionBeforeRenew|TestWorkerPublishesServiceCreatedJobEvents' -count=1
go test ./internal/runtime -run 'TestOpenAI|TestLocalCodexResponsesParser' -count=1
go test ./internal/productdata -run 'TestRepositoryContractCoversM7ToolCallRequestProjection|TestRepositoryContractRecoversExecutingToolCallAfterExpiredLease' -count=1
go test ./internal/productdata ./internal/runtime -run 'TestRepositoryContractCancelsUnresolvedToolCallsWhenRunStops|TestAgentDelegateReconcilePreservesRunScopedWorkspaceRootSnapshot|TestRepeatedWorkspaceToolRequestThisTurnKeysByToolAndCoversGlob' -count=1
go test ./internal/runtime -run 'TestHTTPProviderSerializesAnthropicToolResultContinuation|TestHTTPProviderSerializesGeminiToolResultContinuation|TestHTTPProviderStreamsGeminiFunctionCallsAcrossFramesWithUniqueIDs|TestQueuedRunRouterDoesNotExecuteToolWhenStartReturnsCancelled' -count=1
go test ./internal/runtime -run TestCompactToolResult -count=1
go test ./internal/runtime -run 'TestProviderContextCompactionBoundsLargeToolArguments|TestProviderContextCompactionPreservesToolCallResultPairs|TestProviderContextCompactionKeepsLatestUserAfterToolHistory' -count=1
go test ./internal/httpapi -run TestM98CodeAgentParallelReadDelegateChildFinalSmoke -count=1
go test ./internal/productdata ./internal/runtime -count=1
bun test web/src/realApiClient.test.ts web/src/state.test.ts
LOOMI_TEST_DATABASE_URL=... go test ./internal/runtime -run TestPostgresWorkerWaitsForDelegatedChildRunBeforeParentContinuation -count=1 -v
LOOMI_TEST_DATABASE_URL=... go test ./internal/runtime -run TestPostgresAgentDelegateChildWorkerTerminalResumesParent -count=1 -v
LOOMI_TEST_DATABASE_URL=... go test ./internal/productdata -run 'TestPostgresConcurrentRunEventInsertsSerializeSequence|TestPostgresRunEventsUseUniqueSequenceOrdering|TestPostgresMemoryEntryScopeAndTerminalAudit|TestPostgresReconcilesDelegatedAgentTaskAfterChildRunCompletes' -count=1 -v
bun run build
```

## Notes

Durable run events remain the audit source of truth. The run-step projection is a materialized checkpoint only; missing, stale, or unreadable projection rows are rebuilt from `run_events`.
