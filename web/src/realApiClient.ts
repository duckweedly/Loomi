import type { ApiClient } from './apiClient'
import type { LocalProviderDetection, MemoryAuditItem, MemoryEntry, MemoryFilters, Message, Persona, ProviderCapability, ProviderFamily, Run, RunEvent, RunSource, RunStatus, Thread, ToolCall, ToolCatalogItem, WorkerQueueDiagnostics, WorkerQueueStatus, WorkerStatus } from './domain'
import { isRuntimeTerminal } from './runtime/executionAdapter'
import { applyRealRunEvent } from './runtime/realExecutionAdapter'

const apiBaseUrl = (import.meta.env.VITE_LOOMI_API_BASE_URL ?? '').replace(/\/$/, '')

export function hasRealApiBase() {
  return apiBaseUrl.length > 0
}

type ApiThread = {
  id: string
  title: string
  mode: 'chat' | 'work'
  lifecycle_status: 'active' | 'archived'
  created_at: string
  updated_at: string
  archived_at?: string | null
}

type ApiPersona = {
  id: string
  slug: string
  name: string
  description: string
  active_version: string
  is_default: boolean
}

type ApiMessage = {
  id: string
  thread_id: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
  client_message_id?: string | null
  run_id?: string | null
  attempt_of_message_id?: string | null
}

export type ApiRun = {
  id: string
  thread_id: string
  status: RunStatus
  source: RunSource
  title: string
  created_at: string
  updated_at: string
  completed_at?: string | null
  error_code?: string | null
  error_message?: string | null
}

export type ApiProviderCapability = {
  id: string
  family: ProviderFamily
  base_url?: string | null
  model: string
  status: ProviderCapability['status']
  message?: string | null
}

export type ApiLocalProviderDetection = {
  provider_id: string
  display_name: string
  provider_kind: string
  auth_mode: LocalProviderDetection['authMode']
  status: LocalProviderDetection['status']
  model_candidates: string[]
  source: LocalProviderDetection['source']
  redaction_applied: boolean
  message?: string | null
}

export type ApiWorkerQueueDiagnostics = {
  queue_status: WorkerQueueStatus
  worker_status: WorkerStatus
  queued_count: number
  leased_count: number
  stale_count: number
  retrying_count: number
  dead_count: number
  blocked_tool_approval_count?: number
  resumable_tool_call_count?: number
  updated_at: string
}

export type ApiToolCall = {
  id: string
  thread_id: string
  run_id: string
  tool_call_id: string
  tool_name: string
  arguments_summary: Record<string, unknown>
  approval_status: ToolCall['approvalStatus']
  execution_status: ToolCall['executionStatus']
  result_summary?: Record<string, unknown> | null
  error_code?: string | null
  error_message?: string | null
}

export type ApiToolCatalogItem = {
  name: string
  display_name: string
  description: string
  source: ToolCatalogItem['source']
  group: ToolCatalogItem['group']
  input_schema_hash?: string | null
  risk_level: ToolCatalogItem['riskLevel']
  approval_policy: ToolCatalogItem['approvalPolicy']
  enabled: boolean
  execution_state: ToolCatalogItem['executionState']
}

export type ApiMemoryEntry = {
  id: string
  title: string
  summary: string
  scope_type: 'user' | 'thread'
  scope_id?: string | null
  status?: 'approved' | 'tombstoned' | 'disabled'
  safety_state?: 'safe' | 'redacted' | 'blocked'
  source_thread_id?: string | null
  source_run_id?: string | null
  source_event_id?: string | null
  source_type?: 'manual' | 'thread' | 'run' | null
  created_at: string
  updated_at: string
  deleted_at?: string | null
  redaction_applied?: boolean
}

export type ApiMemoryAuditItem = {
  id: string
  event_type: MemoryAuditItem['eventType']
  summary: string
  thread_id?: string | null
  run_id?: string | null
  memory_entry_id?: string | null
  memory_proposal_id?: string | null
  status?: string | null
  scope_type?: string | null
  source_type?: string | null
  redaction_applied?: boolean
  occurred_at: string
}

type ApiRunEvent = {
  id: string
  run_id: string
  thread_id: string
  sequence: number
  category: 'lifecycle' | 'progress' | 'message' | 'error' | 'final'
  type: string
  summary: string
  content?: string | null
  metadata: Record<string, unknown>
  created_at: string
}

export class ApiRequestError extends Error {
  constructor(message: string, public code: string, public status: number) {
    super(message)
  }
}

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })
  const body = await response.json().catch(() => null)
  if (!response.ok) {
    const message = body?.error?.message ?? `Request failed with ${response.status}`
    const code = body?.error?.code ?? 'request_failed'
    throw new ApiRequestError(message, code, response.status)
  }
  return body as T
}

function mapThread(thread: ApiThread): Thread {
  return {
    id: thread.id,
    title: thread.title,
    project: 'Loomi',
    mode: thread.mode,
    updatedAt: thread.updated_at,
    lifecycleStatus: thread.lifecycle_status,
    runStatus: 'completed',
  }
}

function mapMessage(message: ApiMessage): Message {
  return {
    id: message.id,
    threadId: message.thread_id,
    role: message.role,
    content: message.content,
    createdAt: message.created_at,
    runId: message.run_id ?? undefined,
    attemptOfMessageId: message.attempt_of_message_id ?? undefined,
  }
}

function mapPersona(persona: ApiPersona): Persona {
  return {
    id: persona.id,
    slug: persona.slug,
    name: persona.name,
    description: persona.description,
    activeVersion: persona.active_version,
    isDefault: persona.is_default,
  }
}

function mapApiToolCatalogItem(tool: ApiToolCatalogItem): ToolCatalogItem {
  return {
    name: tool.name,
    displayName: tool.display_name,
    description: tool.description,
    source: tool.source,
    group: tool.group,
    inputSchemaHash: tool.input_schema_hash ?? undefined,
    riskLevel: tool.risk_level,
    approvalPolicy: tool.approval_policy,
    enabled: tool.enabled,
    executionState: tool.execution_state,
  }
}

function restoreAssistantDraftFromEvents(run: Run, events: RunEvent[]): Run {
  return events.reduce((current, event) => {
    if (current.events.some((existing) => existing.id === event.id)) return current

    const lastSequence = current.events.at(-1)?.sequence ?? -1
    const shouldApplyAssistantDelta = !event.assistantDelta || event.sequence === undefined || lastSequence <= event.sequence
    const events = [...current.events, event].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0))
    if (isRuntimeTerminal(current.status)) return { ...current, events }

    const content = event.assistantDelta && shouldApplyAssistantDelta ? `${current.assistantDraft?.content ?? ''}${event.assistantDelta}` : current.assistantDraft?.content ?? ''
    const assistantContent = event.type === 'assistant.message.completed' && event.content ? event.content : content

    if (event.status === 'completed') {
      return { ...current, status: 'completed', events, completedAt: event.time, assistantDraft: { content: event.content ?? assistantContent, status: 'completed', lastEventId: event.id } }
    }
    if (event.status === 'failed' || event.status === 'stopped') {
      return { ...current, status: event.status, events, completedAt: event.time, assistantDraft: { content, status: event.status, lastEventId: event.id } }
    }
    if (event.status === 'recovering' || event.status === 'queued' || event.status === 'stopping') {
      return { ...current, status: event.status, events, assistantDraft: { content, status: event.status, lastEventId: event.id } }
    }
    if (event.type === 'assistant.message.completed' && event.content) {
      return { ...current, status: event.status, events, assistantDraft: { content: event.content, status: 'completed', lastEventId: event.id } }
    }
    return { ...current, status: event.status, events, assistantDraft: event.assistantDelta && shouldApplyAssistantDelta ? { content, status: 'streaming', lastEventId: event.id } : current.assistantDraft }
  }, run)
}

export function mapApiRun(run: ApiRun, events: RunEvent[] = []): Run {
  const assistantDraft = run.status === 'pending' || run.status === 'queued' || run.status === 'running' || run.status === 'recovering' || run.status === 'stopping'
    ? { content: '', status: run.status === 'recovering' ? 'recovering' as const : run.status === 'stopping' ? 'stopping' as const : run.status === 'queued' ? 'queued' as const : 'pending' as const }
    : undefined
  const mappedRun: Run = {
    id: run.id,
    threadId: run.thread_id,
    status: run.status,
    model: run.source === 'model_gateway' ? 'Model gateway' : 'Local simulated',
    context: run.source,
    source: run.source,
    events: [],
    assistantDraft,
  }
  const restored = restoreAssistantDraftFromEvents(mappedRun, events)
  return events
    .filter((event) => event.type.startsWith('tool.call.'))
    .reduce((current, event) => applyRealRunEvent({ ...current, events: current.events.filter((existing) => existing.id !== event.id) }, { ...event, runId: event.runId ?? run.id, threadId: event.threadId ?? run.thread_id }), restored)
}

function metadataString(metadata: Record<string, unknown>) {
  const usageKeys = new Set(['input_tokens', 'output_tokens', 'total_tokens'])
  return Object.entries(metadata)
    .filter(([key, value]) => !usageKeys.has(key) && (typeof value === 'string' || typeof value === 'number'))
    .map(([key, value]) => `${key}: ${value}`)
    .join(' · ')
}

function tokenUsage(metadata: Record<string, unknown>) {
  const inputTokens = typeof metadata.input_tokens === 'number' ? metadata.input_tokens : undefined
  const outputTokens = typeof metadata.output_tokens === 'number' ? metadata.output_tokens : undefined
  const totalTokens = typeof metadata.total_tokens === 'number' ? metadata.total_tokens : undefined
  return inputTokens !== undefined || outputTokens !== undefined || totalTokens !== undefined
    ? { inputTokens, outputTokens, totalTokens }
    : undefined
}

function canonicalRunEventType(event: ApiRunEvent) {
  if (event.type.includes('.')) return event.type
  const m6Types: Record<string, string> = {
    run_queued: 'run.queued',
    job_claimed: 'job.claimed',
    lease_renewed: 'worker.lease_renewed',
    pipeline_step_started: 'pipeline.step.started',
    pipeline_step_completed: 'pipeline.step.completed',
    pipeline_step_failed: 'pipeline.step.failed',
    mcp_discovery_succeeded: 'mcp.discovery.succeeded',
    mcp_discovery_failed: 'mcp.discovery.failed',
    mcp_discovery_rejected: 'mcp.discovery.rejected',
    mcp_tools_available: 'mcp.tools.available',
    mcp_tools_non_executable: 'mcp.tools.non_executable',
    job_recovering: 'job.recovering',
    job_retry_scheduled: 'job.retry_scheduled',
    stop_requested: 'run.stopping',
    job_attempt_failed: 'job.attempt_failed',
    job_retry_exhausted: 'job.retry_exhausted',
    tool_call_requested: 'tool.call.requested',
    tool_call_approval_required: 'tool.call.approval_required',
    tool_call_approved: 'tool.call.approved',
    tool_call_denied: 'tool.call.denied',
    tool_call_executing: 'tool.call.executing',
    tool_call_succeeded: 'tool.call.succeeded',
    tool_call_failed: 'tool.call.failed',
    tool_call_cancelled: 'tool.call.cancelled',
    run_completed: 'run.completed',
    run_failed: 'run.failed',
    run_stopped: 'run.stopped',
  }
  if (m6Types[event.type]) return m6Types[event.type]
  if (event.category === 'progress' && event.type === 'drafting') return 'assistant.drafting'
  if (event.category === 'message' && event.type === 'assistant_message') return 'assistant.message.completed'
  return `${event.category}.${event.type}`
}

export function mapApiWorkerQueueDiagnostics(diagnostics: ApiWorkerQueueDiagnostics): WorkerQueueDiagnostics {
  return {
    queueStatus: diagnostics.queue_status,
    workerStatus: diagnostics.worker_status,
    queuedCount: diagnostics.queued_count,
    leasedCount: diagnostics.leased_count,
    staleCount: diagnostics.stale_count,
    retryingCount: diagnostics.retrying_count,
    deadCount: diagnostics.dead_count,
    blockedToolApprovalCount: diagnostics.blocked_tool_approval_count,
    resumableToolCallCount: diagnostics.resumable_tool_call_count,
    updatedAt: diagnostics.updated_at,
  }
}

export function mapApiProviderCapability(provider: ApiProviderCapability): ProviderCapability {
  return {
    id: provider.id,
    family: provider.family,
    baseUrl: provider.base_url ?? null,
    model: provider.model,
    status: provider.status,
    message: provider.message ?? null,
  }
}

export function mapApiLocalProviderDetection(provider: ApiLocalProviderDetection): LocalProviderDetection {
  return {
    providerId: provider.provider_id,
    displayName: provider.display_name,
    providerKind: provider.provider_kind,
    authMode: provider.auth_mode,
    status: provider.status,
    modelCandidates: provider.model_candidates,
    source: provider.source,
    redactionApplied: provider.redaction_applied,
    message: provider.message ?? null,
  }
}

export function mapApiToolCall(call: ApiToolCall): ToolCall {
  const status = call.approval_status === 'required' && call.execution_status === 'blocked'
    ? 'approval_required'
    : call.approval_status === 'approved' && call.execution_status === 'not_started'
      ? 'approved'
      : call.approval_status === 'denied'
        ? 'denied'
        : call.execution_status === 'executing'
          ? 'executing'
          : call.execution_status === 'succeeded'
            ? 'succeeded'
            : call.execution_status === 'failed'
              ? 'failed'
              : call.execution_status === 'cancelled'
                ? 'cancelled'
                : 'requested'
  return {
    id: call.id,
    toolCallId: call.tool_call_id,
    name: call.tool_name,
    status,
    approvalStatus: call.approval_status,
    executionStatus: call.execution_status,
    summary: status === 'approval_required' ? 'Approval required' : status === 'denied' ? 'Denied' : status === 'approved' ? 'Approved' : status === 'executing' ? 'Executing' : call.tool_name,
    input: JSON.stringify(call.arguments_summary),
    output: call.result_summary ? JSON.stringify(call.result_summary) : '',
    argumentsSummary: call.arguments_summary,
    resultSummary: call.result_summary ?? null,
    errorCode: call.error_code ?? null,
    errorMessage: call.error_message ?? null,
  }
}

export function mapApiMemoryEntry(entry: ApiMemoryEntry): MemoryEntry {
  return {
    id: entry.id,
    title: entry.title,
    summary: entry.summary,
    scopeType: entry.scope_type,
    scopeId: entry.scope_id ?? undefined,
    status: entry.status ?? 'approved',
    safetyState: entry.safety_state ?? undefined,
    sourceThreadId: entry.source_thread_id ?? undefined,
    sourceRunId: entry.source_run_id ?? undefined,
    sourceEventId: entry.source_event_id ?? undefined,
    sourceType: entry.source_type ?? undefined,
    createdAt: entry.created_at,
    updatedAt: entry.updated_at,
    deletedAt: entry.deleted_at ?? undefined,
    redactionApplied: Boolean(entry.redaction_applied),
  }
}

export function mapApiMemoryAuditItem(item: ApiMemoryAuditItem): MemoryAuditItem {
  return {
    id: item.id,
    eventType: item.event_type,
    summary: item.summary,
    threadId: item.thread_id ?? undefined,
    runId: item.run_id ?? undefined,
    memoryEntryId: item.memory_entry_id ?? undefined,
    memoryProposalId: item.memory_proposal_id ?? undefined,
    status: item.status ?? undefined,
    scopeType: item.scope_type ?? undefined,
    sourceType: item.source_type ?? undefined,
    redactionApplied: Boolean(item.redaction_applied),
    occurredAt: item.occurred_at,
  }
}

function memoryQueryString(filters: MemoryFilters = {}, query = '') {
  const params = new URLSearchParams()
  if (query.trim()) params.set('q', query.trim())
  if (filters.scopeType) params.set('scope_type', filters.scopeType)
  if (filters.scopeId?.trim()) params.set('scope_id', filters.scopeId.trim())
  if (filters.sourceThreadId?.trim()) params.set('source_thread_id', filters.sourceThreadId.trim())
  if (filters.sourceRunId?.trim()) params.set('source_run_id', filters.sourceRunId.trim())
  if (filters.sourceType && filters.sourceType !== 'any') params.set('source_type', filters.sourceType)
  if (filters.includeTombstoned) params.set('include_tombstoned', 'true')
  if (filters.limit) params.set('limit', String(filters.limit))
  const encoded = params.toString()
  return encoded ? `?${encoded}` : ''
}

function memoryFilterRequestFields(filters: MemoryFilters = {}) {
  return {
    scope_type: filters.scopeType || undefined,
    scope_id: filters.scopeId?.trim() || undefined,
    source_thread_id: filters.sourceThreadId?.trim() || undefined,
    source_run_id: filters.sourceRunId?.trim() || undefined,
    source_type: filters.sourceType && filters.sourceType !== 'any' ? filters.sourceType : undefined,
    include_tombstoned: filters.includeTombstoned || undefined,
    limit: filters.limit || undefined,
  }
}

function memorySearchRequestBody(query: string, filters: MemoryFilters = {}) {
  return { query, ...memoryFilterRequestFields({ limit: 20, ...filters }) }
}

function memoryDeleteRequestBody(filters: MemoryFilters = {}) {
  return {
    scope_type: filters.scopeType || undefined,
    scope_id: filters.scopeId?.trim() || undefined,
    source_thread_id: filters.sourceThreadId?.trim() || undefined,
    source_run_id: filters.sourceRunId?.trim() || undefined,
  }
}

export function mapApiRunEvent(event: ApiRunEvent): RunEvent {
  const type = canonicalRunEventType(event)
  const metadataDetail = metadataString(event.metadata)
  const isError = /(^|\.)(error|failed|unavailable|timeout)$/.test(type) || event.category === 'error'
  const isModelStream = type.startsWith('model.') || type.startsWith('assistant.') || type === 'message.model_output_delta' || type === 'message.model_output_completed'
  const isToolCall = type.startsWith('tool.call.')
  const isMCP = type.startsWith('mcp.discovery.') || type.startsWith('mcp.tools.')
  return {
    id: event.id,
    runId: event.run_id,
    threadId: event.thread_id,
    sequence: event.sequence,
    category: event.category,
    type,
    label: event.category,
    detail: metadataDetail ? `${event.summary} · ${metadataDetail}` : event.summary,
    content: event.content ?? null,
    time: event.created_at,
    status: statusFromApiEvent(event),
    metadata: event.metadata,
    group: isError
      ? 'error'
      : isModelStream
        ? 'model-stream'
        : isToolCall
          ? 'tool-call'
          : isMCP
            ? 'worker-job'
          : type.startsWith('worker.') || type.startsWith('job.') || type.startsWith('pipeline.')
            ? 'worker-job'
            : 'run-lifecycle',
    severity: isError ? 'error' : type === 'model.delta' || type === 'assistant.drafting' || type === 'message.model_output_delta' ? 'progress' : 'info',
    usage: tokenUsage(event.metadata),
    assistantDelta: type === 'model.delta' || type === 'message.model_output_delta' ? event.content ?? undefined : undefined,
  }
}

function statusFromApiEvent(event: ApiRunEvent): RunStatus {
  const type = canonicalRunEventType(event)
  if (type === 'run.queued') return 'queued'
  if (type === 'tool.call.approval_required') return 'blocked_on_tool_approval'
  if (type === 'job.recovering' || type === 'run.recovering') return 'recovering'
  if (type === 'run.stopping') return 'stopping'
  if (type === 'run.stopped') return 'stopped'
  if (type === 'run.failed' || type === 'job.retry_exhausted' || type === 'pipeline.step.failed') return 'failed'
  if (type === 'run.cancelled') return 'cancelled'
  if (type === 'run.completed') return 'completed'
  if (event.category !== 'final') return event.category === 'error' ? 'failed' : 'running'
  return 'completed'
}

export function createClientMessageID() {
  const random = globalThis.crypto?.randomUUID?.() ?? Math.random().toString(36).slice(2)
  return `web-${Date.now()}-${random}`
}

function deferredRun(threadId: string): Run {
  return {
    id: `deferred-${threadId}`,
    threadId,
    status: 'completed',
    model: 'Deferred',
    context: 'M3 thread/message only',
    events: [],
  }
}

async function loadRunWithEvents(run: ApiRun) {
  const events = await realApiClient.getRunEvents(run.id)
  return mapApiRun(run, events)
}

export const realApiClient: ApiClient = {
  mode: 'real_api',

  async listThreads() {
    const body = await requestJSON<{ threads: ApiThread[] }>('/v1/threads')
    return body.threads.map(mapThread)
  },

  async getThreadMessages(threadId: string) {
    const body = await requestJSON<{ messages: ApiMessage[] }>(`/v1/threads/${threadId}/messages`)
    return body.messages.map(mapMessage)
  },

  async getThreadRun(threadId: string) {
    try {
      const body = await requestJSON<{ run: ApiRun }>(`/v1/threads/${threadId}/runs/current`)
      return loadRunWithEvents(body.run)
    } catch (err) {
      if (err instanceof ApiRequestError && err.code === 'run_not_found') return deferredRun(threadId)
      throw err
    }
  },

  async getRunEvents(runId: string) {
    const body = await requestJSON<{ events: ApiRunEvent[] }>(`/v1/runs/${runId}/events`)
    return body.events.map(mapApiRunEvent)
  },

  async listPersonas() {
    const body = await requestJSON<{ personas: ApiPersona[] }>('/v1/personas')
    return body.personas.map(mapPersona)
  },

  async listModelProviders() {
    const body = await requestJSON<{ providers: ApiProviderCapability[] }>('/v1/model-providers')
    return body.providers.map(mapApiProviderCapability)
  },

  async listToolCatalog() {
    const body = await requestJSON<{ tools: ApiToolCatalogItem[] }>('/v1/tools/catalog')
    return body.tools.map(mapApiToolCatalogItem)
  },

  async listLocalProviderDetections() {
    const body = await requestJSON<{ providers: ApiLocalProviderDetection[] }>('/v1/local-provider-detections')
    return body.providers.map(mapApiLocalProviderDetection)
  },

  async checkModelProvider(providerId: string) {
    const body = await requestJSON<{ provider: ApiProviderCapability }>('/v1/model-providers/check', {
      method: 'POST',
      body: JSON.stringify({ provider_id: providerId }),
    })
    return mapApiProviderCapability(body.provider)
  },

  async saveModelProvider(input: { baseUrl: string; model: string; apiKey: string }) {
    const body = await requestJSON<{ provider: ApiProviderCapability }>('/v1/model-providers', {
      method: 'POST',
      body: JSON.stringify({ base_url: input.baseUrl, model: input.model, api_key: input.apiKey }),
    })
    return mapApiProviderCapability(body.provider)
  },

  async getWorkerQueueDiagnostics() {
    const body = await requestJSON<{ diagnostics: ApiWorkerQueueDiagnostics }>('/v1/diagnostics/worker-queue')
    return mapApiWorkerQueueDiagnostics(body.diagnostics)
  },

  async getToolCall(threadId: string, runId: string, toolCallId: string) {
    const body = await requestJSON<{ tool_call: ApiToolCall }>(`/v1/threads/${threadId}/runs/${runId}/tool-calls/${toolCallId}`)
    return mapApiToolCall(body.tool_call)
  },

  async approveToolCall(threadId: string, runId: string, toolCallId: string) {
    const body = await requestJSON<{ tool_call: ApiToolCall }>(`/v1/threads/${threadId}/runs/${runId}/tool-calls/${toolCallId}/approve`, { method: 'POST' })
    return mapApiToolCall(body.tool_call)
  },

  async denyToolCall(threadId: string, runId: string, toolCallId: string) {
    const body = await requestJSON<{ tool_call: ApiToolCall }>(`/v1/threads/${threadId}/runs/${runId}/tool-calls/${toolCallId}/deny`, { method: 'POST' })
    return mapApiToolCall(body.tool_call)
  },

  async listMemoryEntries(filters = {}) {
    const body = await requestJSON<{ items: ApiMemoryEntry[] }>(`/v1/memory${memoryQueryString(filters)}`)
    return body.items.map(mapApiMemoryEntry)
  },

  async searchMemory(query: string, filters = {}) {
    const body = await requestJSON<{ items: ApiMemoryEntry[] }>('/v1/memory/search', {
      method: 'POST',
      body: JSON.stringify(memorySearchRequestBody(query, filters)),
    })
    return body.items.map(mapApiMemoryEntry)
  },

  async getMemoryEntry(entryId: string, filters = {}) {
    const body = await requestJSON<{ entry: ApiMemoryEntry }>(`/v1/memory/entries/${entryId}${memoryQueryString(filters)}`)
    return mapApiMemoryEntry(body.entry)
  },

  async deleteMemoryEntry(entryId: string, filters = {}) {
    await requestJSON<{ status: string }>(`/v1/memory/entries/${entryId}`, { method: 'DELETE', body: JSON.stringify(memoryDeleteRequestBody(filters)) })
  },

  async listMemoryAudit(filters = {}) {
    const body = await requestJSON<{ items: ApiMemoryAuditItem[] }>(`/v1/memory/audit${memoryQueryString(filters)}`)
    return body.items.map(mapApiMemoryAuditItem)
  },

  async startRun(threadId: string, input: { messageId?: string; source?: RunSource; providerId?: string; model?: string; personaId?: string } = {}) {
    const body = await requestJSON<{ run: ApiRun }>(`/v1/threads/${threadId}/runs`, {
      method: 'POST',
      body: JSON.stringify(input.source === 'model_gateway'
        ? { message_id: input.messageId, source: 'model_gateway', provider_id: input.providerId, model: input.model, persona_id: input.personaId }
        : { script_name: 'm4_smoke', persona_id: input.personaId }),
    })
    return loadRunWithEvents(body.run)
  },

  subscribeRunEvents(runId: string, afterSequence: number, onEvent: (event: RunEvent) => void, onError: () => void) {
    const url = `${apiBaseUrl}/v1/runs/${runId}/events/stream?after_sequence=${afterSequence}`
    const source = new EventSource(url)
    source.addEventListener('run_event', (raw) => {
      try {
        const data = JSON.parse((raw as MessageEvent).data) as { event: ApiRunEvent }
        onEvent(mapApiRunEvent(data.event))
      } catch {
        onError()
        source.close()
      }
    })
    source.onerror = () => {
      onError()
      source.close()
    }
    return () => source.close()
  },

  async createThread(title: string, mode: Thread['mode']) {
    const body = await requestJSON<{ thread: ApiThread }>('/v1/threads', {
      method: 'POST',
      body: JSON.stringify({ title, mode }),
    })
    return mapThread(body.thread)
  },

  async updateThread(threadId: string, input: Partial<Pick<Thread, 'title' | 'mode'>>) {
    const body = await requestJSON<{ thread: ApiThread }>(`/v1/threads/${threadId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    })
    return mapThread(body.thread)
  },

  async archiveThread(threadId: string) {
    const body = await requestJSON<{ thread: ApiThread }>(`/v1/threads/${threadId}/archive`, { method: 'POST' })
    return mapThread(body.thread)
  },

  async sendMessage(threadId: string, content: string, personaId?: string) {
    const created = await requestJSON<{ message: ApiMessage }>(`/v1/threads/${threadId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content, client_message_id: createClientMessageID() }),
    })
    let run: Run | undefined
    try {
      const providers = await this.listModelProviders?.()
      const provider = providers?.find((candidate) => candidate.status === 'available')
      if (!provider) throw new ApiRequestError('Model provider is unavailable.', 'provider_unavailable', 503)
      run = await this.startRun?.(threadId, { messageId: created.message.id, source: 'model_gateway', providerId: provider.id, model: provider.model, personaId })
    } catch (err) {
      if (!(err instanceof ApiRequestError) || err.code !== 'active_run_exists') throw err
      run = await this.getThreadRun(threadId)
    }
    return {
      messages: await this.getThreadMessages(threadId),
      run: run ?? deferredRun(threadId),
    }
  },

  async stopRun(runId: string) {
    const body = await requestJSON<{ run: ApiRun }>(`/v1/runs/${runId}/stop`, { method: 'POST' })
    return loadRunWithEvents(body.run)
  },
}
