import { Terminal } from 'lucide-react'
import type { Run } from '../domain'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import { getRightPanelItemCopy, rightPanelItems, type RightPanelItemId } from '../rightPanelItems'

type Props = {
  open: boolean
  selectedPanelId: RightPanelItemId
  run?: Run | null
  locale?: Locale
}

function PreviewPanel() {
  return (
    <div className="right-panel-preview">
      <pre className="artifact-code">{`$ loomi preview artifact

Panel: workspace shell
Mode: mock
Status: ready

Runtime after M1.`}</pre>
      <div className="artifact-note">
        <Terminal size={15} /> Terminal / browser placeholder
      </div>
    </div>
  )
}

function backgroundEventLabel(type: string, locale: Locale) {
  const copy = getDictionary(locale).runtime.workerJob
  const labels: Record<string, string> = {
    job_claimed: copy.jobClaimed,
    'job.claimed': copy.jobClaimed,
    lease_renewed: copy.leaseRenewed,
    'worker.lease_renewed': copy.leaseRenewed,
    job_recovering: copy.jobRecovering,
    'job.recovering': copy.jobRecovering,
    job_retry_scheduled: copy.retryScheduled,
    'job.retry_scheduled': copy.retryScheduled,
    job_attempt_failed: copy.attemptFailed,
    'job.attempt_failed': copy.attemptFailed,
    job_retry_exhausted: copy.retryExhausted,
    'job.retry_exhausted': copy.retryExhausted,
    pipeline_step_started: 'Pipeline stage started',
    'pipeline.step.started': 'Pipeline stage started',
    pipeline_step_completed: 'Pipeline stage completed',
    'pipeline.step.completed': 'Pipeline stage completed',
    pipeline_step_failed: 'Pipeline stage failed',
    'pipeline.step.failed': 'Pipeline stage failed',
    worker_diagnostics: copy.diagnostics,
  }
  return labels[type] ?? (type.includes('worker') || type.includes('job') || type.includes('pipeline') ? `${copy.unknownWorkerEvent} · ${type}` : type)
}

function jobStatusLabel(status: Run['status'], locale: Locale) {
  const copy = getDictionary(locale).runtime.workerJob
  const labels: Partial<Record<Run['status'], string>> = {
    pending: copy.statusQueued,
    queued: copy.statusQueued,
    running: copy.statusLeased,
    retrying: copy.statusRetrying,
    recovering: copy.statusRecovering,
    completed: copy.statusCompleted,
    failed: copy.statusFailed,
    stopped: copy.statusCancelled,
    cancelled: copy.statusCancelled,
  }
  return labels[status] ?? copy.statusDead
}

function BackgroundTasksPanel({ run, Icon, locale }: { run?: Run | null; Icon: React.ComponentType<{ size?: number; strokeWidth?: number }>; locale: Locale }) {
  const copy = getDictionary(locale).runtime.workerJob
  const workerEvents = run?.events.filter((event) => event.type.includes('worker') || event.type.includes('job') || event.type.includes('lease') || event.type.includes('pipeline')) ?? []
  const diagnostics = workerEvents.find((event) => event.type === 'worker_diagnostics' || event.type.includes('diagnostic'))

  if (!run || workerEvents.length === 0) {
    return (
      <div className="right-panel-empty background-tasks-empty">
        <div className="right-panel-empty-icon">
          <Icon size={24} strokeWidth={1.7} />
        </div>
        <strong>{copy.noTaskRunning}</strong>
        <p>{copy.runRealMessage}</p>
        <span>{copy.readOnlyNoControls}</span>
      </div>
    )
  }

  return (
    <div className="background-task-panel">
      <section className="background-task-section">
        <span className="rail-card-kicker">{copy.currentRunJob}</span>
        <strong>{run.id}</strong>
        <p>{jobStatusLabel(run.status, locale)}</p>
      </section>
      {diagnostics && (
        <section className="background-task-section">
          <span className="rail-card-kicker">{copy.diagnostics}</span>
          <p>{diagnostics.detail}</p>
        </section>
      )}
      <section className="background-task-section">
        <span className="rail-card-kicker">{copy.latestEvents}</span>
        {workerEvents.length === 0 ? (
          <p>{copy.noEventsYet}</p>
        ) : workerEvents.slice(-5).map((event) => (
          <div className="background-task-event" key={event.id}>
            <strong>{backgroundEventLabel(event.type, locale)}</strong>
            <span>{event.detail}</span>
          </div>
        ))}
      </section>
    </div>
  )
}

export function RightToolDrawer({ open, selectedPanelId, run, locale = 'en' }: Props) {
  const selectedPanel = rightPanelItems.find((item) => item.id === selectedPanelId) ?? rightPanelItems[0]
  const selectedPanelCopy = getRightPanelItemCopy(selectedPanel, locale)
  const SelectedIcon = selectedPanel.Icon
  const isPreview = selectedPanel.id === 'preview'
  const isBackgroundTasks = selectedPanel.id === 'background-tasks'

  return (
    <aside className={open ? 'right-tool-drawer open' : 'right-tool-drawer'}>
      <div className="right-panel-head">
        <div>
          <strong>{selectedPanelCopy.title}</strong>
          <span>{isPreview ? (locale === 'zh' ? '产物' : 'Artifact') : isBackgroundTasks ? getDictionary(locale).runtime.workerJob.readOnlyObserver : (locale === 'zh' ? '预留面板' : 'Placeholder')}</span>
        </div>
      </div>
      {isPreview ? (
        <PreviewPanel />
      ) : isBackgroundTasks ? (
        <BackgroundTasksPanel run={run} Icon={SelectedIcon} locale={locale} />
      ) : (
        <div className="right-panel-empty">
          <div className="right-panel-empty-icon">
            <SelectedIcon size={24} strokeWidth={1.7} />
          </div>
          <strong>{selectedPanelCopy.title}</strong>
          <p>{selectedPanelCopy.description}</p>
          <span>{locale === 'zh' ? '即将接入' : 'Coming soon'}</span>
        </div>
      )}
    </aside>
  )
}
