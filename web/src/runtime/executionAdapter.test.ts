import { describe, expect, test } from 'bun:test'
import type { RuntimeEvent } from '../domain'
import { applyRealRunEvent } from './realExecutionAdapter'

describe('M6 execution adapter worker events', () => {
  const baseRun = { id: 'run-a', threadId: 'thread-a', status: 'queued', model: 'Model gateway', context: 'model_gateway', source: 'model_gateway', events: [] } as const

  test('maps queued claimed and completed events into run state', () => {
    const queued: RuntimeEvent = { id: 'evt-queued', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'run.queued', label: 'lifecycle', detail: 'Run queued', time: 'Now', status: 'queued' }
    const claimed: RuntimeEvent = { id: 'evt-claimed', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'job.claimed', label: 'progress', detail: 'Job claimed', time: 'Now', status: 'running' }
    const completed: RuntimeEvent = { id: 'evt-completed', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'run.completed', label: 'final', detail: 'Run completed', time: 'Now', status: 'completed' }

    const queuedRun = applyRealRunEvent(baseRun, queued)
    const runningRun = applyRealRunEvent(queuedRun, claimed)
    const completedRun = applyRealRunEvent(runningRun, completed)

    expect(queuedRun.status).toBe('queued')
    expect(queuedRun.assistantDraft).toMatchObject({ status: 'queued' })
    expect(runningRun.status).toBe('running')
    expect(completedRun.status).toBe('completed')
    expect(completedRun.completedAt).toBe('Now')
    expect(completedRun.events.map((event) => event.type)).toEqual(['run.queued', 'job.claimed', 'run.completed'])
  })

  test('maps recovery and retry exhaustion into visible run states', () => {
    const recovering: RuntimeEvent = { id: 'evt-recovering', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job.recovering', label: 'progress', detail: 'Job recovering', time: 'Now', status: 'recovering' }
    const retry: RuntimeEvent = { id: 'evt-retry', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'job.retry_scheduled', label: 'progress', detail: 'Job retry scheduled', time: 'Now', status: 'recovering' }
    const exhausted: RuntimeEvent = { id: 'evt-exhausted', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'job.retry_exhausted', label: 'error', detail: 'Retries exhausted', time: 'Later', status: 'failed' }

    const recoveringRun = applyRealRunEvent(baseRun, recovering)
    const retryRun = applyRealRunEvent(recoveringRun, retry)
    const failedRun = applyRealRunEvent(retryRun, exhausted)

    expect(recoveringRun.status).toBe('recovering')
    expect(recoveringRun.assistantDraft).toMatchObject({ status: 'recovering' })
    expect(retryRun.status).toBe('recovering')
    expect(failedRun.status).toBe('failed')
    expect(failedRun.completedAt).toBe('Later')
    expect(failedRun.assistantDraft).toMatchObject({ status: 'failed' })
    expect(failedRun.events.map((event) => event.type)).toEqual(['job.recovering', 'job.retry_scheduled', 'job.retry_exhausted'])
  })

  test('keeps pipeline events observable while the worker run is active', () => {
    const started: RuntimeEvent = { id: 'evt-step-started', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'pipeline.step.started', label: 'progress', detail: 'Invoke runtime started', time: 'Now', status: 'running', group: 'worker-job' }
    const completed: RuntimeEvent = { id: 'evt-step-completed', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'pipeline.step.completed', label: 'progress', detail: 'Invoke runtime completed', time: 'Now', status: 'running', group: 'worker-job' }

    const next = applyRealRunEvent(applyRealRunEvent(baseRun, started), completed)

    expect(next.status).toBe('running')
    expect(next.events.map((event) => event.type)).toEqual(['pipeline.step.started', 'pipeline.step.completed'])
  })
})
