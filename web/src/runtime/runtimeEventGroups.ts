import type { RunEvent, RuntimeEventGroup } from '../domain'

export type RuntimeEventGroupView = {
  id: RuntimeEventGroup
  title: string
  events: RunEvent[]
}

const eventGroups: Array<{ id: RuntimeEventGroup; title: string }> = [
  { id: 'run-lifecycle', title: 'Run lifecycle' },
  { id: 'model-stream', title: 'Model stream' },
  { id: 'worker-job', title: 'Worker/job' },
  { id: 'error', title: 'Error' },
]

function isErrorEvent(event: RunEvent) {
  return event.status === 'failed' || event.severity === 'error' || /(^|\.)(error|failed|unavailable|timeout)$/.test(event.type)
}

export function mapRuntimeEventGroup(event: RunEvent): RuntimeEventGroup {
  if (isErrorEvent(event)) return 'error'
  if (event.group) return event.group
  if (event.type.startsWith('model.') || event.type.startsWith('assistant.')) return 'model-stream'
  if (event.type.startsWith('worker.') || event.type.startsWith('job.')) return 'worker-job'
  return 'run-lifecycle'
}

export function groupRuntimeEvents(events: RunEvent[]): RuntimeEventGroupView[] {
  return eventGroups.map((group) => ({
    ...group,
    events: events
      .filter((event) => mapRuntimeEventGroup(event) === group.id)
      .sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0)),
  }))
}
