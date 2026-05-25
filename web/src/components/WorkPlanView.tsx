import { FileText, ListChecks, LoaderCircle, PackageCheck } from 'lucide-react'
import type { WorkPlanProjection, WorkStepStatus } from '../domain'

type Props = {
  projection: WorkPlanProjection
  loading: boolean
  error?: string | null
}

const stepLabels: Record<WorkStepStatus, string> = {
  pending: 'Pending',
  running: 'Running',
  completed: 'Done',
  blocked: 'Blocked',
  failed: 'Failed',
}

export function WorkPlanView({ projection, loading, error }: Props) {
  if (loading) {
    return (
      <section className="work-plan-view loading" aria-label="Work plan">
        <LoaderCircle size={16} strokeWidth={1.7} />
        <strong>Loading work plan</strong>
      </section>
    )
  }

  if (error) {
    return (
      <section className="work-plan-view error" aria-label="Work plan">
        <strong>Work plan unavailable</strong>
        <span>{error}</span>
      </section>
    )
  }

  return (
    <section className="work-plan-view" aria-label="Work plan">
      <div className="work-plan-header">
        <div>
          <span className="rail-card-kicker">Work plan</span>
          <h2>{projection.goal}</h2>
        </div>
        <span className={`work-plan-status ${projection.status}`}>{projection.status}</span>
      </div>

      {projection.emptyReason ? (
        <div className="work-plan-empty">
          <strong>No plan yet</strong>
          <span>{projection.emptyReason}</span>
        </div>
      ) : (
        <div className="work-plan-grid">
          <section className="work-plan-section">
            <div className="work-plan-section-title">
              <ListChecks size={15} strokeWidth={1.8} />
              <strong>Steps</strong>
            </div>
            {projection.steps.length ? projection.steps.map((step, index) => (
              <div className="work-step-row" key={step.id}>
                <span className={`work-step-index ${step.status}`}>{index + 1}</span>
                <div>
                  <strong>{step.title}</strong>
                  {step.summary && <span>{step.summary}</span>}
                </div>
                <em>{stepLabels[step.status]}</em>
              </div>
            )) : <p>No steps projected yet.</p>}
          </section>

          <section className="work-plan-section">
            <div className="work-plan-section-title">
              <PackageCheck size={15} strokeWidth={1.8} />
              <strong>Artifacts</strong>
            </div>
            {projection.artifacts.length ? projection.artifacts.map((artifact) => (
              <article className="work-artifact-card" key={artifact.id}>
                <div>
                  <strong>{artifact.title}</strong>
                  <span>{artifact.type}</span>
                </div>
                <p>{artifact.summary}</p>
                <small>{[artifact.sourceThreadId, artifact.sourceRunId, artifact.updatedAt ?? artifact.createdAt].filter(Boolean).join(' · ')}</small>
              </article>
            )) : <p>No artifact references yet.</p>}
          </section>
        </div>
      )}

      <section className="work-plan-section recent">
        <div className="work-plan-section-title">
          <FileText size={15} strokeWidth={1.8} />
          <strong>Recent progress</strong>
        </div>
        <p>{projection.statusDetail}</p>
        {projection.recentEvents.length ? (
          <div className="work-progress-list">
            {projection.recentEvents.map((event) => (
              <div className="work-progress-row" key={event.id}>
                <span>{event.time}</span>
                <strong>{event.type}</strong>
                <em>{event.detail}</em>
              </div>
            ))}
          </div>
        ) : <span className="work-plan-muted">No recent events yet.</span>}
      </section>
    </section>
  )
}
