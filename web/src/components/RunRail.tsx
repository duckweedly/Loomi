import { useState } from 'react'
import { Check, ChevronDown, Minus } from 'lucide-react'
import type { Run, RuntimeScriptId } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import type { BackendCapabilityStatus } from '../runtime/backendCapabilityStatus'
import { getBackendCapabilityCopy } from '../runtime/backendCapabilityStatus'
import { buildToolEventPreview, redactPreviewText } from '../runtime/toolPreview'
import { AgentStateMotion } from './AgentStateMotion'

type Props = {
  run: Run | null
  open: boolean
  onStopRun?: () => void
  selectedRuntimeScript?: RuntimeScriptId
  capabilityStatus?: BackendCapabilityStatus
  locale?: Locale
  onSelectRuntimeScript?: (scriptId: RuntimeScriptId) => void
}

function getEventClassName(event: Run['events'][number]) {
  if (event.severity === 'error' || event.group === 'error' || event.status === 'failed') return 'progress-row failed'
  if (event.status === 'stopped' || event.status === 'cancelled' || event.severity === 'warning') return 'progress-row warning'
  if (event.status === 'queued' || event.status === 'running' || event.status === 'retrying' || event.status === 'recovering' || event.status === 'stopping') return 'progress-row active'
  return 'progress-row done'
}

function getEventMark(event: Run['events'][number], index: number) {
  if (event.severity === 'error' || event.group === 'error' || event.status === 'failed') return <Minus size={10} />
  if (event.status === 'stopped' || event.status === 'cancelled' || event.severity === 'warning') return <Minus size={10} />
  if (event.status === 'queued' || event.status === 'running' || event.status === 'retrying' || event.status === 'recovering' || event.status === 'stopping') return index + 1
  return <Check size={11} />
}

const railCopy = {
  zh: {
    title: '进度',
    recent: '最近活动',
    noEvents: '暂无活动',
    scenario: '场景',
    success: '成功',
    fail: '失败',
    stopRun: '停止运行',
    runCreated: '已创建运行',
    runCompleted: '运行完成',
    runFailed: '运行失败',
    runStopped: '运行已停止',
    runCancelled: '运行已取消',
    modelUsage: '模型用量',
    modelStarted: '模型开始响应',
    modelCompleted: '模型响应完成',
    messageCompleted: '回复完成',
    initialModelPhase: '初始模型响应',
    continuationModelPhase: '继续模型响应',
    providerFailure: 'Provider 异常',
    toolBlocked: '等待工具确认',
    detailsRecorded: '详情已记录',
    tokens: (parts: string[]) => parts.join(' / '),
  },
  en: {
    title: 'Progress',
    recent: 'Recent activity',
    noEvents: 'No activity yet',
    scenario: 'Scenario',
    success: 'Success',
    fail: 'Fail',
    stopRun: 'Stop run',
    runCreated: 'Run created',
    runCompleted: 'Run completed',
    runFailed: 'Run failed',
    runStopped: 'Run stopped',
    runCancelled: 'Run cancelled',
    modelUsage: 'Model usage',
    modelStarted: 'Model started',
    modelCompleted: 'Model completed',
    messageCompleted: 'Message',
    initialModelPhase: 'Initial model phase',
    continuationModelPhase: 'Continuation model phase',
    providerFailure: 'Provider failure',
    toolBlocked: 'Waiting for tool confirmation',
    detailsRecorded: 'Details recorded',
    tokens: (parts: string[]) => parts.join(' / '),
  },
}

function compactDetail(detail: string) {
  const redacted = redactPreviewText(detail).replace(/\s+/g, ' ').trim()
  if (!redacted) return ''
  if (/(^|[\s,])[\w.-]*_id\s*:/i.test(redacted) || /\b(runId|threadId|messageId|personaId|providerId)\b/i.test(redacted)) return ''
  if (/[a-z]+_[0-9a-f]{8,}/i.test(redacted)) return ''
  return redacted.length > 96 ? `${redacted.slice(0, 93)}...` : redacted
}

function statusTitle(event: Run['events'][number], locale: Locale) {
  const copy = getDictionary(locale).runtime.workerJob
  const labels: Partial<Record<Run['status'], string>> = {
    pending: copy.statusQueued,
    queued: copy.statusQueued,
    running: copy.running,
    retrying: copy.statusRetrying,
    recovering: copy.statusRecovering,
    completed: copy.statusCompleted,
    failed: copy.statusFailed,
    stopped: copy.statusCancelled,
    cancelled: copy.statusCancelled,
  }
  return labels[event.status] ?? event.status
}

function getToolEventDetail(event: Run['events'][number], loopCopy: string, locale: Locale) {
  return buildToolEventPreview(event, loopCopy, locale).primary
}

function getToolEventMetadataDetail(event: Run['events'][number], loopCopy: string, locale: Locale) {
  return buildToolEventPreview(event, loopCopy, locale).details
}

function getEventDetail(event: Run['events'][number], locale: Locale) {
  const workerCopy = getDictionary(locale).runtime.workerJob
  const usage = event.usage
  const usageParts = usage ? [
    usage.inputTokens !== undefined ? `${usage.inputTokens} in` : null,
    usage.outputTokens !== undefined ? `${usage.outputTokens} out` : null,
    usage.totalTokens !== undefined ? `${usage.totalTokens} total` : null,
  ].filter(Boolean) : []
  const eventLabels: Record<string, string> = {
    job_claimed: workerCopy.jobClaimed,
    'job.claimed': workerCopy.jobClaimed,
    lease_renewed: workerCopy.leaseRenewed,
    'worker.lease_renewed': workerCopy.leaseRenewed,
    job_recovering: workerCopy.jobRecovering,
    'job.recovering': workerCopy.jobRecovering,
    job_retry_scheduled: workerCopy.retryScheduled,
    'job.retry_scheduled': workerCopy.retryScheduled,
    job_attempt_failed: workerCopy.attemptFailed,
    'job.attempt_failed': workerCopy.attemptFailed,
    job_retry_exhausted: workerCopy.retryExhausted,
    'job.retry_exhausted': workerCopy.retryExhausted,
    cancellation: workerCopy.cancellationRequested,
    'run.cancelled': workerCopy.cancellationRequested,
    worker_diagnostics: workerCopy.diagnostics,
    'worker.claimed': workerCopy.jobClaimed,
  }
  const modelPhase = event.metadata?.model_phase === 'continuation'
    ? 'Continuation model phase'
    : event.metadata?.model_phase === 'initial'
      ? 'Initial model phase'
      : ''
  const loopIndex = typeof event.metadata?.loop_index === 'number' ? event.metadata.loop_index : undefined
  const loopMax = typeof event.metadata?.loop_max === 'number' ? event.metadata.loop_max : undefined
  const loopCopy = loopIndex !== undefined
    ? loopMax !== undefined
      ? `Loop ${loopIndex}/${loopMax}`
      : `Loop ${loopIndex}`
    : ''
  const humanToolDetail = event.type.startsWith('tool.call.') ? getToolEventDetail(event, loopCopy, locale) : ''
  const detail = modelPhase
    ? `${modelPhase} · ${redactPreviewText(event.detail)}`
    : humanToolDetail
      ? humanToolDetail
    : event.type === 'error.provider_error' || event.type === 'error.provider_timeout' || event.type === 'error.provider_rate_limited'
    ? `Provider failure · ${redactPreviewText(event.detail)}`
    : event.type === 'progress.tool_call_blocked'
      ? `Tool request blocked · ${redactPreviewText(event.detail)}`
      : eventLabels[event.type]
        ? `${eventLabels[event.type]} · ${redactPreviewText(event.detail)}`
        : event.type.includes('worker') || event.type.includes('job')
          ? `${workerCopy.unknownWorkerEvent} · ${event.type} · ${redactPreviewText(event.detail)}`
          : redactPreviewText(event.detail)
  return usageParts.length > 0 ? `${detail} · ${usageParts.join(' / ')}` : detail
}

function getEventSecondaryDetail(event: Run['events'][number], locale: Locale) {
  if (!event.type.startsWith('tool.call.')) return ''
  const loopIndex = typeof event.metadata?.loop_index === 'number' ? event.metadata.loop_index : undefined
  const loopMax = typeof event.metadata?.loop_max === 'number' ? event.metadata.loop_max : undefined
  const loopCopy = loopIndex !== undefined
    ? loopMax !== undefined
      ? `Loop ${loopIndex}/${loopMax}`
      : `Loop ${loopIndex}`
    : ''
  return getToolEventMetadataDetail(event, loopCopy, locale)
}

function getEventView(event: Run['events'][number], locale: Locale) {
  const copy = railCopy[locale]
  const workerCopy = getDictionary(locale).runtime.workerJob
  const usage = event.usage
  const usageParts = usage ? [
    usage.inputTokens !== undefined ? `${usage.inputTokens} in` : null,
    usage.outputTokens !== undefined ? `${usage.outputTokens} out` : null,
    usage.totalTokens !== undefined ? `${usage.totalTokens} total` : null,
  ].filter((part): part is string => Boolean(part)) : []
  const workerLabels: Record<string, string> = {
    job_claimed: workerCopy.jobClaimed,
    'job.claimed': workerCopy.jobClaimed,
    lease_renewed: workerCopy.leaseRenewed,
    'worker.lease_renewed': workerCopy.leaseRenewed,
    job_recovering: workerCopy.jobRecovering,
    'job.recovering': workerCopy.jobRecovering,
    job_retry_scheduled: workerCopy.retryScheduled,
    'job.retry_scheduled': workerCopy.retryScheduled,
    job_attempt_failed: workerCopy.attemptFailed,
    'job.attempt_failed': workerCopy.attemptFailed,
    job_retry_exhausted: workerCopy.retryExhausted,
    'job.retry_exhausted': workerCopy.retryExhausted,
    cancellation: workerCopy.cancellationRequested,
    'run.cancelled': copy.runCancelled,
    worker_diagnostics: workerCopy.diagnostics,
    'worker.claimed': workerCopy.jobClaimed,
  }

  if (event.type.startsWith('tool.call.')) {
    return {
      title: getEventDetail(event, locale),
      detail: getEventSecondaryDetail(event, locale),
      debug: event.type,
    }
  }
  if (event.type === 'run.created') return { title: copy.runCreated, detail: statusTitle(event, locale), debug: event.type }
  if (event.type === 'run.completed') return { title: copy.runCompleted, detail: compactDetail(event.detail), debug: event.type }
  if (event.type === 'run.stopped') return { title: copy.runStopped, detail: compactDetail(event.detail), debug: event.type }
  if (event.type === 'run.cancelled') return { title: copy.runCancelled, detail: compactDetail(event.detail), debug: event.type }
  if (event.status === 'failed' || event.severity === 'error') return { title: copy.runFailed, detail: compactDetail(event.detail), debug: event.type }
  if (event.type === 'model.usage') return { title: copy.modelUsage, detail: usageParts.length > 0 ? copy.tokens(usageParts) : compactDetail(event.detail), debug: event.type }
  if (event.type.startsWith('model.') || event.type.startsWith('message.') || event.type.startsWith('assistant.')) {
    const title = event.metadata?.model_phase === 'initial'
      ? copy.initialModelPhase
      : event.metadata?.model_phase === 'continuation'
        ? copy.continuationModelPhase
        : event.type.includes('completed') || event.type.includes('message.completed')
          ? copy.messageCompleted
          : copy.modelStarted
    return { title, detail: compactDetail(event.detail), debug: event.type }
  }
  if (event.type === 'error.provider_error' || event.type === 'error.provider_timeout' || event.type === 'error.provider_rate_limited' || event.type.includes('provider')) {
    return { title: copy.providerFailure, detail: compactDetail(event.detail), debug: event.type }
  }
  if (event.type === 'progress.tool_call_blocked') return { title: copy.toolBlocked, detail: compactDetail(event.detail), debug: event.type }
  if (workerLabels[event.type]) return { title: workerLabels[event.type], detail: compactDetail(event.detail), debug: event.type }
  if (event.type.includes('worker') || event.type.includes('job') || event.type.includes('pipeline')) {
    return { title: workerCopy.unknownWorkerEvent, detail: compactDetail(event.detail), debug: event.type }
  }

  return {
    title: event.label && event.label !== event.type ? event.label : copy.detailsRecorded,
    detail: compactDetail(event.detail),
    debug: event.type,
  }
}

function isRecentActivityEvent(event: Run['events'][number]) {
  if (event.type.startsWith('tool.call.')) return true
  if (event.type === 'run.completed' || event.type === 'run.failed' || event.type === 'run.stopped' || event.type === 'run.cancelled') return true
  if (event.type === 'model.output.completed' || event.type === 'message.completed') return true
  if (event.status === 'failed' || event.severity === 'error') return true
  return false
}

function displayEventTime(value: string, locale: Locale) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString(locale === 'zh' ? 'zh-CN' : 'en-US', { hour: '2-digit', minute: '2-digit' })
}

export function RunRail({ run, open, onStopRun, selectedRuntimeScript = 'success', capabilityStatus, locale = 'en', onSelectRuntimeScript }: Props) {
  const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set())
  const toggleSection = (section: string) => {
    setCollapsedSections((current) => {
      const next = new Set(current)
      if (next.has(section)) next.delete(section)
      else next.add(section)
      return next
    })
  }

  const capabilityCopy = capabilityStatus ? getBackendCapabilityCopy(capabilityStatus, locale) : null
  const copy = railCopy[locale]
  const recentEvents = (run?.events ?? []).filter(isRecentActivityEvent).slice(-5).reverse()

  return (
    <aside className={open ? 'floating-rail open' : 'floating-rail'}>
      <section className={collapsedSections.has('progress') ? 'rail-card progress-card collapsed' : 'rail-card progress-card'}>
        <button className="rail-card-head" onClick={() => toggleSection('progress')}>
          <h2>{copy.title}</h2>
          <ChevronDown size={18} />
        </button>
        <div className="rail-card-body progress-list">
          <AgentStateMotion run={run} compact locale={locale} />
          {onSelectRuntimeScript && (
            <div className="runtime-script-switch compact" aria-label="Mock runtime script">
              <span>{copy.scenario}</span>
              <button className={selectedRuntimeScript === 'success' ? 'selected' : undefined} onClick={() => onSelectRuntimeScript('success')}>{copy.success}</button>
              <button className={selectedRuntimeScript === 'failure' ? 'selected' : undefined} onClick={() => onSelectRuntimeScript('failure')}>{copy.fail}</button>
            </div>
          )}
          {capabilityCopy && <div className={`capability-rail ${capabilityStatus}`}><strong>{capabilityCopy.title}</strong><span>{capabilityCopy.detail}</span></div>}
          {(run?.status === 'queued' || run?.status === 'running' || run?.status === 'retrying' || run?.status === 'recovering' || run?.status === 'blocked_on_tool_approval') && onStopRun && <button className="runtime-stop-button ghost" onClick={onStopRun}>{copy.stopRun}</button>}
          <section className="runtime-event-group recent">
            <h3>{copy.recent}</h3>
            {recentEvents.length === 0 ? <p className="runtime-event-empty">{copy.noEvents}</p> : recentEvents.map((event, index) => {
              const eventView = getEventView(event, locale)
              return (
                <div key={event.id} className={getEventClassName(event)}>
                  <span className="progress-mark">{getEventMark(event, recentEvents.length - index - 1)}</span>
                  <span className="progress-copy">
                    <strong>{eventView.title}</strong>
                    {eventView.detail && <small>{eventView.detail}</small>}
                  </span>
                  <small>{displayEventTime(event.time, locale)}</small>
                </div>
              )
            })}
          </section>
        </div>
      </section>

    </aside>
  )
}
