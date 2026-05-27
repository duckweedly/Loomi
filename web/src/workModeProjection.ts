import type { Message, Run, RunEvent, Thread, WorkArtifactReference, WorkPlanProjection, WorkProgressEvent, WorkStep, WorkStepStatus, WorkTodoItem, WorkTodoSnapshot } from './domain'
import { humanToolName } from './runtime/toolPreview'

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

function redactionFlag(value: unknown): boolean {
  if (typeof value === 'boolean') return value
  if (typeof value === 'string') return value.trim().toLowerCase() === 'true'
  return false
}

function valueWouldRedact(value: unknown): boolean {
  if (typeof value === 'string' || typeof value === 'number') {
    const text = String(value).trim()
    return !!text && (SECRET_PATTERN.test(text) || EXECUTION_VALUE_PATTERN.test(text))
  }
  if (Array.isArray(value)) return value.some(valueWouldRedact)
  if (isRecord(value)) return Object.entries(value).some(([key, item]) => EXECUTION_KEY_PATTERN.test(key) || valueWouldRedact(item))
  return false
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

function deriveSteps(metadata: Record<string, unknown>[]): WorkStep[] {
  const metadataSteps = metadata.flatMap((item) => metadataList(item.work_steps ?? item.steps))
  if (metadataSteps.length) return metadataSteps.map(stepFromMetadata)
  return []
}

function safeTodoUpdatedBy(value: unknown): WorkTodoSnapshot['updatedBy'] {
  return value === 'provider' || value === 'runtime' || value === 'user' ? value : undefined
}

function todoItemFromMetadata(item: Record<string, unknown>, index: number): WorkTodoItem {
  const redactionApplied = redactionFlag(item.redaction_applied ?? item.redactionApplied) || valueWouldRedact(item)
  return {
    id: safeString(item.id, `todo-${index + 1}`),
    title: safeString(item.title, `Todo ${index + 1}`),
    status: safeStatus(item.status, index === 0 ? 'running' : 'pending'),
    summary: safeString(item.summary) || undefined,
    redactionApplied: redactionApplied || undefined,
  }
}

function todoSnapshotFromEvent(event: RunEvent): WorkTodoSnapshot | null {
  if (!isRecord(event.metadata)) return null
  const items = metadataList(event.metadata.todo_items ?? event.metadata.todoItems)
  if (!items.length) return null
  const redactionApplied = redactionFlag(event.metadata.redaction_applied ?? event.metadata.redactionApplied)
    || valueWouldRedact(event.metadata)
    || items.some(valueWouldRedact)

  return {
    items: items.map(todoItemFromMetadata),
    updatedBy: safeTodoUpdatedBy(event.metadata.updated_by ?? event.metadata.updatedBy),
    updatedAtEventId: safeString(event.metadata.updated_at_event_id ?? event.metadata.updatedAtEventId, event.id),
    redactionApplied,
  }
}

function deriveTodoSnapshot(run: Run | null): WorkTodoSnapshot | undefined {
  const snapshots = (run?.events ?? []).map(todoSnapshotFromEvent).filter((snapshot): snapshot is WorkTodoSnapshot => snapshot !== null)
  return snapshots.at(-1)
}

function artifactFromMetadata(item: Record<string, unknown>, index: number, thread: Thread, run: Run | null, redactionApplied: boolean): WorkArtifactReference {
  return {
    id: safeString(item.id, `artifact-${index + 1}`),
    title: safeString(item.title, `Artifact ${index + 1}`),
    type: safeString(item.type, 'artifact'),
    sourceThreadId: safeString(item.source_thread_id ?? item.sourceThreadId, thread.id),
    sourceRunId: safeString(item.source_run_id ?? item.sourceRunId, run?.id ?? ''),
    summary: safeString(item.summary, 'Safe metadata preview only.'),
    createdAt: safeString(item.created_at ?? item.createdAt) || undefined,
    updatedAt: safeString(item.updated_at ?? item.updatedAt) || undefined,
    redactionApplied,
  }
}

function deriveArtifacts(thread: Thread, run: Run | null, metadata: Record<string, unknown>[]) {
  const metadataArtifacts = metadata.flatMap((item) => metadataList(item.work_artifacts ?? item.artifacts))
  return metadataArtifacts.map((item, index) => {
    const safeItem = Object.fromEntries(Object.entries(item).filter(([key]) => !EXECUTION_KEY_PATTERN.test(key)))
    const removedExecutableFields = Object.keys(safeItem).length !== Object.keys(item).length
    const redactionApplied = removedExecutableFields || redactionFlag(item.redaction_applied ?? item.redactionApplied) || valueWouldRedact(item)
    return artifactFromMetadata(safeItem, index, thread, run, redactionApplied)
  })
}

function eventDetail(event: RunEvent) {
  if (event.type.startsWith('tool.call.')) return eventTitle(event)
  return safeString(event.detail || event.content || event.type, event.type)
}

function eventTitle(event: RunEvent) {
  if (event.type.startsWith('tool.call.')) {
    const toolName = typeof event.metadata?.tool_name === 'string' ? event.metadata.tool_name : undefined
    const state = event.type.endsWith('.approval_required')
      ? '等待确认'
      : event.type.endsWith('.succeeded')
        ? '完成'
        : event.type.endsWith('.failed')
          ? '失败'
          : '运行中'
    return `${humanToolName(toolName, 'zh')} ${state}`
  }
  if (event.type === 'run.completed') return '运行完成'
  if (event.type === 'run.failed') return '运行失败'
  if (event.type.includes('artifact')) return '产物已更新'
  if (event.type.includes('plan') || event.type.includes('todo')) return '计划已更新'
  if (event.type.includes('assistant') || event.type.includes('message')) return '回复完成'
  if (event.group === 'worker-job') return '任务进度已更新'
  return '进度已更新'
}

function deriveRecentEvents(run: Run | null): WorkProgressEvent[] {
  return (run?.events ?? [])
    .filter((event) => event.type.includes('work') || event.type.includes('plan') || event.type.includes('artifact') || event.group === 'worker-job' || event.group === 'tool-call' || event.status !== 'completed')
    .slice(-6)
    .map((event) => ({
      id: event.id,
      title: eventTitle(event),
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
  const steps = deriveSteps(metadata)
  const todoSnapshot = deriveTodoSnapshot(run)
  const artifacts = deriveArtifacts(thread, run, metadata)
  if (!steps.length && !todoSnapshot && !artifacts.length) return null

  const goal = deriveGoal(thread, messages, metadata)
  const recentEvents = deriveRecentEvents(run)

  return {
    goal,
    steps,
    todoSnapshot,
    status: run?.status ?? 'empty',
    statusDetail: statusDetail(run, recentEvents),
    artifacts,
    recentEvents,
  }
}
