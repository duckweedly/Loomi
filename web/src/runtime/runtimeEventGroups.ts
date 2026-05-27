import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import type { RunEvent, RuntimeEventGroup } from '../domain'

export type RuntimeEventGroupView = {
  id: RuntimeEventGroup
  title: string
  events: RunEvent[]
}

const eventGroups: RuntimeEventGroup[] = ['run-lifecycle', 'model-stream', 'worker-job', 'tool-call', 'error']

function isErrorEvent(event: RunEvent) {
  return event.status === 'failed' || event.severity === 'error' || /(^|[._])(error|failed|unavailable|timeout|attempt_failed|retry_exhausted)$/.test(event.type)
}

export function mapRuntimeEventGroup(event: RunEvent): RuntimeEventGroup {
  if (isErrorEvent(event)) return 'error'
  if (event.group) return event.group
  if (event.type.startsWith('model.') || event.type.startsWith('assistant.')) return 'model-stream'
  if (event.type.startsWith('tool.call.')) return 'tool-call'
  if (event.type.startsWith('worker.') || event.type.startsWith('job.') || event.type.startsWith('pipeline.')) return 'worker-job'
  if (event.type.startsWith('mcp.discovery.') || event.type.startsWith('mcp.tools.')) return 'worker-job'
  if (event.type.startsWith('worker_') || event.type.startsWith('job_') || event.type === 'lease_renewed' || event.type.includes('worker')) return 'worker-job'
  return 'run-lifecycle'
}

export function groupRuntimeEvents(events: RunEvent[], locale: Locale = 'en'): RuntimeEventGroupView[] {
  const copy = getDictionary(locale).runtime.eventGroups
  return eventGroups.map((group) => ({
    id: group,
    title: copy[group],
    events: events
      .filter((event) => mapRuntimeEventGroup(event) === group),
  }))
}
