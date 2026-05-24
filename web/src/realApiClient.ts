import type { ApiClient } from './apiClient'
import type { Message, Run, RunEvent, RunStatus, Thread } from './domain'
import { isRuntimeTerminal } from './runtime/executionAdapter'

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
  source: 'local_simulated' | 'real_model'
  title: string
  created_at: string
  updated_at: string
  completed_at?: string | null
  error_code?: string | null
  error_message?: string | null
}

export type ApiRunEvent = {
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

function restoreAssistantDraftFromEvents(run: Run, events: RunEvent[]): Run {
  return events.reduce((current, event) => {
    if (isRuntimeTerminal(current.status)) return current
    if (current.events.some((existing) => existing.id === event.id)) return current

    const lastSequence = current.events.at(-1)?.sequence ?? -1
    const shouldApplyAssistantDelta = !event.assistantDelta || event.sequence === undefined || lastSequence <= event.sequence
    const events = [...current.events, event].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0))
    const content = event.assistantDelta && shouldApplyAssistantDelta ? `${current.assistantDraft?.content ?? ''}${event.assistantDelta}` : current.assistantDraft?.content ?? ''
    const assistantContent = event.type === 'assistant.message.completed' && event.content ? event.content : content

    if (event.status === 'completed') {
      return { ...current, status: 'completed', events, completedAt: event.time, assistantDraft: { content: event.content ?? assistantContent, status: 'completed', lastEventId: event.id } }
    }
    if (event.status === 'failed' || event.status === 'stopped') {
      return { ...current, status: event.status, events, completedAt: event.time, assistantDraft: { content, status: event.status, lastEventId: event.id } }
    }
    if (event.status === 'recovering') {
      return { ...current, status: 'recovering', events, assistantDraft: { content, status: 'recovering', lastEventId: event.id } }
    }
    if (event.type === 'assistant.message.completed' && event.content) {
      return { ...current, status: event.status, events, assistantDraft: { content: event.content, status: 'completed', lastEventId: event.id } }
    }
    return { ...current, status: event.status, events, assistantDraft: event.assistantDelta && shouldApplyAssistantDelta ? { content, status: 'streaming', lastEventId: event.id } : current.assistantDraft }
  }, run)
}

export function mapApiRun(run: ApiRun, events: RunEvent[] = []): Run {
  const assistantDraft = run.status === 'pending' || run.status === 'running' || run.status === 'recovering'
    ? { content: '', status: run.status === 'recovering' ? 'recovering' as const : 'pending' as const }
    : undefined
  const mappedRun = {
    id: run.id,
    threadId: run.thread_id,
    status: run.status,
    model: run.source === 'real_model' ? 'Real model' : 'Local simulated',
    context: run.source,
    events: [],
    assistantDraft,
  }
  return restoreAssistantDraftFromEvents(mappedRun, events)
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
  if (event.category === 'progress' && event.type === 'drafting') return 'assistant.drafting'
  if (event.category === 'message' && event.type === 'assistant_message') return 'assistant.message.completed'
  if (event.category === 'final' && event.type === 'run_completed') return 'run.completed'
  if (event.category === 'final' && event.type === 'run_stopped') return 'run.stopped'
  if (event.category === 'final' && event.type === 'run_failed') return 'run.failed'
  return `${event.category}.${event.type}`
}

export function mapApiRunEvent(event: ApiRunEvent): RunEvent {
  const type = canonicalRunEventType(event)
  const metadataDetail = metadataString(event.metadata)
  const isError = /(^|\.)(error|failed|unavailable|timeout)$/.test(type) || event.category === 'error'
  const isModelStream = type.startsWith('model.') || type.startsWith('assistant.')
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
    group: isError
      ? 'error'
      : isModelStream
        ? 'model-stream'
        : type.startsWith('worker.') || type.startsWith('job.')
          ? 'worker-job'
          : 'run-lifecycle',
    severity: isError ? 'error' : type === 'model.delta' || type === 'assistant.drafting' ? 'progress' : 'info',
    usage: tokenUsage(event.metadata),
    assistantDelta: type === 'model.delta' ? event.content ?? undefined : undefined,
  }
}

function statusFromApiEvent(event: ApiRunEvent): RunStatus {
  if (event.type === 'run.recovering' || event.type === 'run_recovering') return 'recovering'
  if (event.type === 'run.stopped' || event.type === 'run_stopped' || event.type === 'run_stopped') return 'stopped'
  if (event.type === 'run.failed' || event.type === 'run_failed') return 'failed'
  if (event.type === 'run.cancelled' || event.type === 'run_cancelled') return 'cancelled'
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

  async startRun(threadId: string) {
    const body = await requestJSON<{ run: ApiRun }>(`/v1/threads/${threadId}/runs`, {
      method: 'POST',
      body: JSON.stringify({ script_name: 'm4_smoke' }),
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

  async sendMessage(threadId: string, content: string) {
    await requestJSON<{ message: ApiMessage }>(`/v1/threads/${threadId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content, client_message_id: createClientMessageID() }),
    })
    let run: Run | undefined
    try {
      run = await this.startRun?.(threadId)
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
