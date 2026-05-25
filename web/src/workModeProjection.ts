import type { Message, Run, RunEvent, Thread, WorkArtifactReference, WorkPlanProjection, WorkProgressEvent, WorkStep, WorkStepStatus } from './domain'

const SECRET_PATTERN = /(sk-[a-z0-9_-]{8,}|api[_-]?key|token|secret|authorization|bearer\s+[a-z0-9._-]+)/i
const EXECUTION_KEY_PATTERN = /(path|file|command|shell|browser|filesystem|execute|url)/i
const EXECUTION_VALUE_PATTERN = /(https?:\/\/|file:\/\/|(?:^|\s)(?:\/Users\/|\/tmp\/|\.\/|\.\.\/|~\/)|[a-z]:\\|(?:^|\s)(?:open|cat|rm|curl|wget|node|python|bash|zsh)\s+(?:\/|\.|~|https?:\/\/)|\b(?:shell|browser|filesystem)\s+(?:tool|automation|command|control|exec))/i

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function safeString(value: unknown, fallback = '') {
  if (typeof value !== 'string' && typeof value !== 'number') return fallback
  const text = String(value).trim()
  if (!text) return fallback
  return SECRET_PATTERN.test(text) || EXECUTION_VALUE_PATTERN.test(text) ? '[redacted]' : text
}

function safeStatus(value: unknown, fallback: WorkStepStatus): WorkStepStatus {
  return value === 'pending' || value === 'running' || value === 'completed' || value === 'blocked' || value === 'failed' ? value : fallback
}

function metadataList(value: unknown): Record<string, unknown>[] {
  if (!Array.isArray(value)) return []
  return value.filter(isRecord)
}

function metadataFromEvents(run: Run | null): Record<string, unknown>[] {
  return run?.events.map((event) => event.metadata).filter(isRecord) ?? []
}

function firstMetadataString(metadata: Record<string, unknown>[], keys: string[]) {
  for (const item of metadata) {
    for (const key of keys) {
      const value = safeString(item[key])
      if (value) return value
    }
  }
  return ''
}

function deriveGoal(thread: Thread, messages: Message[], metadata: Record<string, unknown>[]) {
  return firstMetadataString(metadata, ['work_goal', 'goal'])
    || safeString(messages.find((message) => message.role === 'user')?.content)
    || thread.title
}

function stepFromMetadata(item: Record<string, unknown>, index: number): WorkStep {
  return {
    id: safeString(item.id, `step-${index + 1}`),
    title: safeString(item.title, `Step ${index + 1}`),
    status: safeStatus(item.status, index === 0 ? 'running' : 'pending'),
    summary: safeString(item.summary) || undefined,
  }
}

function deriveSteps(messages: Message[], metadata: Record<string, unknown>[], run: Run | null): WorkStep[] {
  const metadataSteps = metadata.flatMap((item) => metadataList(item.work_steps ?? item.steps))
  if (metadataSteps.length) return metadataSteps.map(stepFromMetadata)

  const userMessages = messages.filter((message) => message.role === 'user')
  if (userMessages.length) {
    return userMessages.slice(0, 4).map((message, index) => ({
      id: `message-step-${message.id}`,
      title: safeString(message.content, `Step ${index + 1}`),
      status: index === userMessages.length - 1 && run && run.status !== 'completed' ? 'running' : 'completed',
    }))
  }
  return []
}

function artifactFromMetadata(item: Record<string, unknown>, index: number, thread: Thread, run: Run | null): WorkArtifactReference {
  return {
    id: safeString(item.id, `artifact-${index + 1}`),
    title: safeString(item.title, `Artifact ${index + 1}`),
    type: safeString(item.type, 'artifact'),
    sourceThreadId: safeString(item.source_thread_id ?? item.sourceThreadId, thread.id),
    sourceRunId: safeString(item.source_run_id ?? item.sourceRunId, run?.id ?? ''),
    summary: safeString(item.summary, 'Safe metadata preview only.'),
    createdAt: safeString(item.created_at ?? item.createdAt) || undefined,
    updatedAt: safeString(item.updated_at ?? item.updatedAt) || undefined,
  }
}

function deriveArtifacts(thread: Thread, run: Run | null, metadata: Record<string, unknown>[]) {
  const metadataArtifacts = metadata.flatMap((item) => metadataList(item.work_artifacts ?? item.artifacts))
  return metadataArtifacts.map((item, index) => {
    const safeItem = Object.fromEntries(Object.entries(item).filter(([key]) => !EXECUTION_KEY_PATTERN.test(key)))
    return artifactFromMetadata(safeItem, index, thread, run)
  })
}

function eventDetail(event: RunEvent) {
  return safeString(event.detail || event.content || event.type, event.type)
}

function deriveRecentEvents(run: Run | null): WorkProgressEvent[] {
  return (run?.events ?? [])
    .filter((event) => event.type.includes('work') || event.type.includes('plan') || event.type.includes('artifact') || event.group === 'worker-job' || event.group === 'tool-call' || event.status !== 'completed')
    .slice(-6)
    .map((event) => ({
      id: event.id,
      type: event.type,
      detail: eventDetail(event),
      time: event.time,
      status: event.status,
    }))
}

function statusDetail(run: Run | null, recentEvents: WorkProgressEvent[]) {
  if (!run) return 'No active run yet.'
  return recentEvents.at(-1)?.detail || `Run ${run.status}`
}

export function safeWorkMetadataPreview(value: unknown): string {
  if (!isRecord(value)) return safeString(value)
  return Object.entries(value)
    .filter(([key]) => !EXECUTION_KEY_PATTERN.test(key))
    .map(([key, item]) => `${key}: ${safeString(item, '[safe metadata]')}`)
    .join(' · ')
}

export function deriveWorkPlanProjection(thread: Thread | null, messages: Message[], run: Run | null): WorkPlanProjection | null {
  if (!thread || thread.mode !== 'work') return null

  const metadata = metadataFromEvents(run)
  const goal = deriveGoal(thread, messages, metadata)
  const steps = deriveSteps(messages, metadata, run)
  const artifacts = deriveArtifacts(thread, run, metadata)
  const recentEvents = deriveRecentEvents(run)
  const emptyReason = !messages.length && !run ? 'No work plan metadata yet.' : undefined

  return {
    goal,
    steps,
    status: run?.status ?? 'empty',
    statusDetail: statusDetail(run, recentEvents),
    artifacts,
    recentEvents,
    emptyReason,
  }
}
