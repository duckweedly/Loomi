import { describe, expect, test } from 'bun:test'
import { applyRunStreamEventToRun, shouldApplyRuntimeEvent } from './state'
import type { Run, RunEvent } from './domain'

describe('shouldApplyRuntimeEvent', () => {
  test('rejects events from old selected threads and superseded runs', () => {
    expect(shouldApplyRuntimeEvent({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-b', runId: 'run-a', activeRunId: 'run-a' })).toBe(false)
    expect(shouldApplyRuntimeEvent({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-a', runId: 'run-a', activeRunId: 'run-b' })).toBe(false)
  })

  test('deduplicates replayed draft events before appending assistant deltas', () => {
    const run: Run = { id: 'run-a', threadId: 'thread-a', status: 'running', model: 'Mock', context: 'Ready', events: [] }
    const event: RunEvent = { id: 'evt-delta-1', type: 'model.delta', label: 'model', detail: 'Hello', content: 'Hello', assistantDelta: 'Hello', time: 'Now', status: 'running' }

    const once = applyRunStreamEventToRun(run, event)
    const replayed = applyRunStreamEventToRun(once, event)

    expect(replayed.assistantDraft?.content).toBe('Hello')
    expect(replayed.events).toHaveLength(1)
    expect(replayed.assistantDraft?.lastEventId).toBe('evt-delta-1')
  })

  test('preserves stopped drafts when stale finals replay for the same run', () => {
    const stoppedRun: Run = {
      id: 'run-a',
      threadId: 'thread-a',
      status: 'stopped',
      model: 'Mock',
      context: 'Ready',
      events: [{ id: 'evt-stop', type: 'run.stopped', label: 'stopped', detail: 'Stopped', time: 'Now', status: 'stopped' }],
      assistantDraft: { content: 'Partial', status: 'stopped', lastEventId: 'evt-stop' },
    }
    const finalEvent: RunEvent = { id: 'evt-final', type: 'model.final', label: 'final', detail: 'Final', content: 'Final', time: 'Later', status: 'completed' }

    const result = applyRunStreamEventToRun(stoppedRun, finalEvent)

    expect(result.status).toBe('stopped')
    expect(result.assistantDraft?.content).toBe('Partial')
    expect(result.assistantDraft?.status).toBe('stopped')
  })

  test('accepts events for the current selected thread and active run', () => {
    expect(shouldApplyRuntimeEvent({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-a', runId: 'run-a', activeRunId: 'run-a' })).toBe(true)
  })
})
