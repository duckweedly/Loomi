import { describe, expect, test } from 'bun:test'
import type { Message, Run } from './domain'
import { createRuntimeEvent, getRuntimeScriptSteps } from './runtime/runtimeScripts'
import { appendRuntimeEventToRun, applyAssistantDeltaToRun, applyModelGatewayEventToRun, applyRunStreamEventToRun, createRegenerateAttemptRun, createRetryAttemptRun, createWorkspaceSettingsState, shouldApplyIncomingRunEvent, shouldBlockRuntimeSubmit, shouldIgnoreTerminalRuntimeEvent, shouldUpdateStreamStateForRunEvent } from './state'

const run: Run = {
  id: 'run-a',
  threadId: 'thread-a',
  status: 'pending',
  model: 'Mock',
  context: 'Ready',
  events: [],
  assistantDraft: { content: '', status: 'empty' },
}

const message: Message = {
  id: 'msg-a',
  threadId: 'thread-a',
  role: 'user',
  content: 'hello',
  createdAt: 'Now',
}

describe('runtime state orchestration helpers', () => {
  test('creates session-local settings defaults', () => {
    expect(createWorkspaceSettingsState()).toEqual({ defaultWorkspaceMode: 'chat', selectedRuntimeScript: 'success' })
    expect(createWorkspaceSettingsState({ defaultWorkspaceMode: 'work', selectedRuntimeScript: 'failure' })).toEqual({ defaultWorkspaceMode: 'work', selectedRuntimeScript: 'failure' })
  })

  test('blocks a second submit while a selected run is pending, running, retrying, or recovering', () => {
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'pending' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'queued' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'running' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'retrying' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'recovering' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'stopping' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'completed' })).toBe(false)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'cancelled' })).toBe(false)
    expect(shouldBlockRuntimeSubmit(null)).toBe(false)
  })

  test('appends events in order and updates run status from event status', () => {
    const next = appendRuntimeEventToRun(run, { id: 'evt-a', runId: run.id, threadId: run.threadId, type: 'run.created', label: 'Run', detail: '已创建', time: 'Now', status: 'running' })

    expect(next.status).toBe('running')
    expect(next.events.map((event) => event.type)).toEqual(['run.created'])
  })

  test('accumulates assistant draft without changing the user message', () => {
    const next = applyAssistantDeltaToRun(run, '片段')

    expect(message.content).toBe('hello')
    expect(next.assistantDraft).toMatchObject({ content: '片段', status: 'streaming' })
  })

  test('ignores out-of-order model deltas that arrive after a later sequence', () => {
    const runningRun: Run = {
      ...run,
      status: 'running',
      events: [{ id: 'evt-2', runId: run.id, threadId: run.threadId, sequence: 2, type: 'model.delta', label: 'Model', detail: 'later', time: 'Now', status: 'running', assistantDelta: 'later' }],
      assistantDraft: { content: 'later', status: 'streaming', lastEventId: 'evt-2' },
    }

    const next = applyRunStreamEventToRun(runningRun, { id: 'evt-1', runId: run.id, threadId: run.threadId, sequence: 1, type: 'model.delta', label: 'Model', detail: 'earlier', time: 'Earlier', status: 'running', assistantDelta: 'earlier' })

    expect(next.assistantDraft?.content).toBe('later')
    expect(next.events.map((event) => event.id)).toEqual(['evt-2', 'evt-1'])
  })

  test('keeps stale-delta guard based on highest known sequence after lower sequence arrivals', () => {
    const runningRun: Run = {
      ...run,
      status: 'running',
      events: [{ id: 'evt-3', runId: run.id, threadId: run.threadId, sequence: 3, type: 'model.delta', label: 'Model', detail: 'latest', time: 'Now', status: 'running', assistantDelta: 'latest' }],
      assistantDraft: { content: 'latest', status: 'streaming', lastEventId: 'evt-3' },
    }

    const withLateEvent = applyRunStreamEventToRun(runningRun, { id: 'evt-1', runId: run.id, threadId: run.threadId, sequence: 1, type: 'model.delta', label: 'Model', detail: 'earlier', time: 'Earlier', status: 'running', assistantDelta: 'earlier' })
    const next = applyRunStreamEventToRun(withLateEvent, { id: 'evt-2', runId: run.id, threadId: run.threadId, sequence: 2, type: 'model.delta', label: 'Model', detail: 'middle', time: 'Middle', status: 'running', assistantDelta: 'middle' })

    expect(next.assistantDraft?.content).toBe('latest')
    expect(next.events.map((event) => event.id)).toEqual(['evt-3', 'evt-1', 'evt-2'])
  })

  test('preserves normalized event identity when applying model gateway events', () => {
    const event = { id: 'evt-delta', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'hel', assistantDelta: 'hel', time: 'Now', status: 'running' } as const

    const next = applyModelGatewayEventToRun(run, event)

    expect(next.events[0]).toEqual(event)
  })

  test('applies model gateway delta and completion events to assistant draft', () => {
    const drafting = applyModelGatewayEventToRun(run, { id: 'evt-delta', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'hel', assistantDelta: 'hel', time: 'Now', status: 'running' })
    const completed = applyModelGatewayEventToRun(drafting, { id: 'evt-complete', runId: run.id, threadId: run.threadId, type: 'message.model_output_completed', label: 'message', detail: 'Model output completed', content: 'hello', time: 'Now', status: 'running' })

    expect(drafting.assistantDraft).toMatchObject({ content: 'hel', status: 'streaming', lastEventId: 'evt-delta' })
    expect(completed.assistantDraft).toMatchObject({ content: 'hello', status: 'completed' })
  })

  test('ignores duplicate stream events before applying assistant deltas', () => {
    const current = { ...run, events: [{ id: 'evt-a', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'hel', assistantDelta: 'hel', time: 'Now', status: 'running' }] }

    expect(shouldApplyIncomingRunEvent(current, { id: 'evt-a', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'hel', assistantDelta: 'hel', time: 'Now', status: 'running' })).toBe(false)
  })

  test('does not update stream state for ignored terminal-run events', () => {
    const current = { ...run, status: 'stopped' as const }
    const lateEvent = { id: 'evt-late', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'Late delta', content: 'late', assistantDelta: 'late', time: 'Now', status: 'running' } as const

    expect(shouldUpdateStreamStateForRunEvent(current, lateEvent)).toBe(false)
  })

  test('ignores later script events after a terminal run event', () => {
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'failed' })).toBe(true)
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'stopped' })).toBe(true)
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'cancelled' })).toBe(true)
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'retrying' })).toBe(false)
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'recovering' })).toBe(false)
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'running' })).toBe(false)
  })

  test('does not append deltas or stale completion events to terminal runs', () => {
    const stoppedRun: Run = { ...run, status: 'stopped', assistantDraft: { content: 'partial', status: 'stopped' } }
    const withDelta = applyAssistantDeltaToRun(stoppedRun, ' stale')
    const withFinal = appendRuntimeEventToRun(stoppedRun, {
      id: 'evt-final',
      runId: stoppedRun.id,
      threadId: stoppedRun.threadId,
      type: 'model.final',
      label: 'Final',
      detail: 'stale final',
      time: 'Later',
      status: 'completed',
    })

    expect(withDelta.assistantDraft).toEqual(stoppedRun.assistantDraft)
    expect(withFinal.status).toBe('stopped')
    expect(withFinal.events).toEqual(stoppedRun.events)
  })

  test('keeps terminal runs unchanged when applying live stream events', () => {
    const stoppedRun: Run = { ...run, status: 'stopped' }
    const staleEvent = {
      id: 'evt-final',
      runId: stoppedRun.id,
      threadId: stoppedRun.threadId,
      type: 'model.final' as const,
      label: 'Final',
      detail: 'stale final',
      time: 'Later',
      status: 'completed' as const,
    }

    expect(applyRunStreamEventToRun(stoppedRun, staleEvent)).toBe(stoppedRun)
  })

  test('replays stopping and stopped worker events', () => {
    const runningRun: Run = { ...run, status: 'running', events: [] }
    const events: Run['events'] = [
      { id: 'evt-stopping', runId: run.id, threadId: run.threadId, sequence: 1, type: 'run.stopping', label: 'Run', detail: 'stopping', time: 'Now', status: 'stopping' },
      { id: 'evt-stopped', runId: run.id, threadId: run.threadId, sequence: 2, type: 'run.stopped', label: 'Run', detail: 'stopped', time: 'Later', status: 'stopped' },
    ]

    const next = events.reduce(applyRunStreamEventToRun, runningRun)

    expect(next.status).toBe('stopped')
    expect(next.assistantDraft).toMatchObject({ status: 'stopped' })
    expect(next.events.map((event) => event.type)).toEqual(['run.stopping', 'run.stopped'])
  })

  test('replays recovery history before retry exhaustion failure', () => {
    const recoveringRun: Run = { ...run, status: 'recovering', events: [] }
    const events: Run['events'] = [
      { id: 'evt-recovering', runId: run.id, threadId: run.threadId, sequence: 1, type: 'job.recovering', label: 'Worker', detail: 'recovering', time: 'Now', status: 'recovering' },
      { id: 'evt-retry', runId: run.id, threadId: run.threadId, sequence: 2, type: 'job.retry_scheduled', label: 'Worker', detail: 'retry scheduled', time: 'Now', status: 'recovering' },
      { id: 'evt-exhausted', runId: run.id, threadId: run.threadId, sequence: 3, type: 'job.retry_exhausted', label: 'Worker', detail: 'exhausted', time: 'Later', status: 'failed' },
    ]

    const next = events.reduce(applyRunStreamEventToRun, recoveringRun)

    expect(next.status).toBe('failed')
    expect(next.assistantDraft).toMatchObject({ status: 'failed' })
    expect(next.events.map((event) => event.type)).toEqual(['job.recovering', 'job.retry_scheduled', 'job.retry_exhausted'])
  })

  test('replays queued worker history before terminal completion', () => {
    const queuedRun: Run = { ...run, status: 'queued', events: [] }
    const events: Run['events'] = [
      { id: 'evt-queued', runId: run.id, threadId: run.threadId, sequence: 1, type: 'run.queued', label: 'Run', detail: 'queued', time: 'Now', status: 'queued' },
      { id: 'evt-claimed', runId: run.id, threadId: run.threadId, sequence: 2, type: 'job.claimed', label: 'Worker', detail: 'claimed', time: 'Now', status: 'running' },
      { id: 'evt-completed', runId: run.id, threadId: run.threadId, sequence: 3, type: 'run.completed', label: 'Run', detail: 'completed', time: 'Later', status: 'completed' },
    ]

    const next = events.reduce(applyRunStreamEventToRun, queuedRun)

    expect(next.status).toBe('completed')
    expect(next.events.map((event) => event.type)).toEqual(['run.queued', 'job.claimed', 'run.completed'])
  })

  test('merges live stream events into non-terminal runs', () => {
    const runningRun: Run = { ...run, status: 'running' }
    const event = {
      id: 'evt-completed',
      runId: runningRun.id,
      threadId: runningRun.threadId,
      type: 'run.completed' as const,
      label: 'Run',
      detail: 'completed',
      time: 'Later',
      status: 'completed' as const,
    }

    const next = applyRunStreamEventToRun(runningRun, event)

    expect(next).not.toBe(runningRun)
    expect(next.status).toBe('completed')
    expect(next.events).toEqual([event])
  })

  test('keeps model final and run completed when applying model stream events', () => {
    const applied = getRuntimeScriptSteps('model-stream').reduce((current, step, index) => {
      return applyRunStreamEventToRun(current, createRuntimeEvent({ threadId: current.threadId, runId: current.id, sequence: index, step }))
    }, { ...run, status: 'running' } as Run)

    expect(applied.status).toBe('completed')
    expect(applied.events.map((event) => event.type)).toEqual(['run.created', 'job.queued', 'worker.claimed', 'job.retrying', 'model.delta', 'model.delta', 'model.final', 'run.completed'])
  })

  test('keeps model error and run failed when applying model error events', () => {
    const applied = getRuntimeScriptSteps('model-error').reduce((current, step, index) => {
      return applyRunStreamEventToRun(current, createRuntimeEvent({ threadId: current.threadId, runId: current.id, sequence: index, step }))
    }, { ...run, status: 'running' } as Run)

    expect(applied.status).toBe('failed')
    expect(applied.events.map((event) => event.type)).toEqual(['run.created', 'model.delta', 'provider.error', 'model.error', 'run.failed'])
  })

  test('creates retry and regenerate attempts without clearing prior context', () => {
    const failedRun: Run = { ...run, status: 'failed', assistantDraft: { content: 'partial', status: 'failed' } }
    const retryRun = createRetryAttemptRun(failedRun)
    const regenerateRun = createRegenerateAttemptRun(run, 'msg-a')

    expect(retryRun.status).toBe('pending')
    expect(retryRun.assistantDraft).toEqual({ content: '', status: 'pending' })
    expect(failedRun.assistantDraft).toEqual({ content: 'partial', status: 'failed' })
    expect(regenerateRun.attemptOfMessageId).toBe('msg-a')
    expect(regenerateRun.status).toBe('pending')
  })
})
