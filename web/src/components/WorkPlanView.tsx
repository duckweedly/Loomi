import { FileText, ListChecks, LoaderCircle, PackageCheck } from 'lucide-react'
import type { WorkPlanProjection, WorkStepStatus } from '../domain'
import type { Locale } from '../i18n'

type Props = {
  projection: WorkPlanProjection
  loading: boolean
  error?: string | null
  locale?: Locale
}

const workPlanCopy: Record<Locale, {
  aria: string
  loading: string
  unavailable: string
  workPlan: string
  waitingMetadata: string
  steps: string
  todos: string
  artifacts: string
  recent: string
  planQueue: string
  active: string
  done: string
  blocked: string
  noSteps: string
  noTodos: string
  noArtifacts: string
  noEvents: string
  updatedBy: (name: string) => string
  redacted: string
  status: Record<WorkStepStatus | WorkPlanProjection['status'], string>
}> = {
  zh: {
    aria: '工作计划',
    loading: '正在读取工作计划',
    unavailable: '工作计划不可用',
    workPlan: '工作计划',
    waitingMetadata: '等待计划信息',
    steps: '步骤',
    todos: '待办',
    artifacts: '产物',
    recent: '最近进度',
    planQueue: '计划队列',
    active: '进行中',
    done: '已完成',
    blocked: '需处理',
    noSteps: '还没有步骤。',
    noTodos: '还没有待办。',
    noArtifacts: '还没有产物引用。',
    noEvents: '还没有进度。',
    updatedBy: (name) => `更新来源 ${name}`,
    redacted: '已隐藏敏感信息',
    status: {
      empty: '空',
      pending: '待处理',
      queued: '排队中',
      running: '运行中',
      retrying: '重试中',
      recovering: '恢复中',
      blocked: '受阻',
      blocked_on_tool_approval: '等待确认',
      stopping: '停止中',
      completed: '完成',
      failed: '失败',
      stopped: '已停止',
      cancelled: '已取消',
    },
  },
  en: {
    aria: 'Work plan',
    loading: 'Loading work plan',
    unavailable: 'Work plan unavailable',
    workPlan: 'Work plan',
    waitingMetadata: 'Waiting for plan metadata',
    steps: 'Steps',
    todos: 'Todos',
    artifacts: 'Artifacts',
    recent: 'Recent progress',
    planQueue: 'Plan queue',
    active: 'Active',
    done: 'Done',
    blocked: 'Needs attention',
    noSteps: 'No steps projected yet.',
    noTodos: 'No todo snapshot yet.',
    noArtifacts: 'No artifact references yet.',
    noEvents: 'No recent events yet.',
    updatedBy: (name) => `Updated by ${name}`,
    redacted: 'Redacted unsafe metadata',
    status: {
      empty: 'Empty',
      pending: 'Pending',
      queued: 'Queued',
      running: 'Running',
      retrying: 'Retrying',
      recovering: 'Recovering',
      blocked: 'Blocked',
      blocked_on_tool_approval: 'Confirm',
      stopping: 'Stopping',
      completed: 'Done',
      failed: 'Failed',
      stopped: 'Stopped',
      cancelled: 'Cancelled',
    },
  },
}

function displayTime(value: string, locale: Locale) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleTimeString(locale === 'zh' ? 'zh-CN' : 'en-US', { hour: '2-digit', minute: '2-digit' })
}

function localizedProjectionText(value: string, locale: Locale) {
  if (locale !== 'zh') return value
  if (value === 'No plan metadata yet.') return '还没有计划信息。'
  if (value === 'No active run yet.') return '还没有运行记录。'
  const runStatus = value.match(/^Run (.+)$/)
  if (runStatus) return `运行状态 ${workPlanCopy.zh.status[runStatus[1] as WorkPlanProjection['status']] ?? runStatus[1]}`
  return value
}

function countStatuses(projection: WorkPlanProjection) {
  const items = [...projection.steps, ...(projection.todoSnapshot?.items ?? [])]
  return {
    active: items.filter((item) => item.status === 'running' || item.status === 'pending').length,
    done: items.filter((item) => item.status === 'completed').length,
    blocked: items.filter((item) => item.status === 'blocked' || item.status === 'failed').length,
  }
}

export function WorkPlanView({ projection, loading, error, locale = 'en' }: Props) {
  const copy = workPlanCopy[locale]
  const counts = countStatuses(projection)
  if (loading) {
    return (
      <section className="work-plan-view loading" aria-label={copy.aria}>
        <LoaderCircle size={16} strokeWidth={1.7} />
        <strong>{copy.loading}</strong>
      </section>
    )
  }

  if (error) {
    return (
      <section className="work-plan-view error" aria-label={copy.aria}>
        <strong>{copy.unavailable}</strong>
        <span>{error}</span>
      </section>
    )
  }

  return (
    <section className="work-plan-view" aria-label={copy.aria}>
      <div className="work-plan-header">
        <div>
          <span className="rail-card-kicker">{copy.workPlan}</span>
          <h2>{projection.goal}</h2>
        </div>
        <span className={`work-plan-status ${projection.status}`}>{copy.status[projection.status] ?? projection.status}</span>
      </div>
      <div className="work-plan-queue-strip" aria-label={copy.planQueue}>
        <span><strong>{projection.steps.length}</strong>{copy.steps}</span>
        <span><strong>{projection.todoSnapshot?.items.length ?? 0}</strong>{copy.todos}</span>
        <span><strong>{counts.active}</strong>{copy.active}</span>
        <span><strong>{counts.done}</strong>{copy.done}</span>
        <span><strong>{counts.blocked}</strong>{copy.blocked}</span>
        <span><strong>{projection.artifacts.length}</strong>{copy.artifacts}</span>
      </div>

      {projection.emptyReason ? (
        <div className="work-plan-empty">
          <strong>{copy.waitingMetadata}</strong>
          <span>{localizedProjectionText(projection.emptyReason, locale)}</span>
        </div>
      ) : (
        <div className="work-plan-grid">
          <section className="work-plan-section">
            <div className="work-plan-section-title">
              <ListChecks size={15} strokeWidth={1.8} />
              <strong>{copy.steps}</strong>
            </div>
            {projection.steps.length ? projection.steps.map((step, index) => (
              <div className="work-step-row" key={`${step.id}-${index}`}>
                <span className={`work-step-index ${step.status}`}>{index + 1}</span>
                <div>
                  <strong>{step.title}</strong>
                  {step.summary && <span>{step.summary}</span>}
                </div>
                <em>{copy.status[step.status]}</em>
              </div>
            )) : <p>{copy.noSteps}</p>}
          </section>

          <section className="work-plan-section">
            <div className="work-plan-section-title">
              <ListChecks size={15} strokeWidth={1.8} />
              <strong>{copy.todos}</strong>
            </div>
            {projection.todoSnapshot?.items.length ? (
              <>
                {projection.todoSnapshot.items.map((todo, index) => (
                  <div className="work-step-row" key={`${todo.id}-${index}`}>
                    <span className={`work-step-index ${todo.status}`}>{index + 1}</span>
                    <div>
                      <strong>{todo.title}</strong>
                      {todo.summary && <span>{todo.summary}</span>}
                    </div>
                    <em>{copy.status[todo.status]}</em>
                  </div>
                ))}
                <p>
                  {[
                    projection.todoSnapshot.updatedBy ? copy.updatedBy(projection.todoSnapshot.updatedBy) : '',
                    projection.todoSnapshot.updatedAtEventId,
                    projection.todoSnapshot.redactionApplied ? copy.redacted : '',
                  ].filter(Boolean).join(' · ')}
                </p>
              </>
            ) : <p>{copy.noTodos}</p>}
          </section>

          <section className="work-plan-section">
            <div className="work-plan-section-title">
              <PackageCheck size={15} strokeWidth={1.8} />
              <strong>{copy.artifacts}</strong>
            </div>
            {projection.artifacts.length ? projection.artifacts.map((artifact, index) => (
              <article className="work-artifact-card" key={`${artifact.id}-${index}`}>
                <div>
                  <strong>{artifact.title}</strong>
                  <span>{artifact.type}</span>
                </div>
                <p>{artifact.summary}</p>
                <small>{[artifact.type, artifact.updatedAt ?? artifact.createdAt].filter(Boolean).join(' · ')}</small>
                {artifact.redactionApplied && <span className="work-artifact-redaction">{copy.redacted}</span>}
              </article>
            )) : <p>{copy.noArtifacts}</p>}
          </section>
        </div>
      )}

      <section className="work-plan-section recent">
        <div className="work-plan-section-title">
          <FileText size={15} strokeWidth={1.8} />
          <strong>{copy.recent}</strong>
        </div>
        <p>{localizedProjectionText(projection.statusDetail, locale)}</p>
        {projection.recentEvents.length ? (
          <div className="work-progress-list">
            {projection.recentEvents.map((event, index) => (
              <div className="work-progress-row" key={`${event.id}-${index}`}>
                <span>{displayTime(event.time, locale)}</span>
                <strong>{event.title ?? event.type}</strong>
                {event.detail && event.detail !== event.title && <em>{event.detail}</em>}
              </div>
            ))}
          </div>
        ) : <span className="work-plan-muted">{copy.noEvents}</span>}
      </section>
    </section>
  )
}
