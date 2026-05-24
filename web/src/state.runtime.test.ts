import { describe, expect, test } from 'bun:test'
import type { Message, Run } from './domain'
import { appendRuntimeEventToRun, applyAssistantDeltaToRun, applyModelGatewayEventToRun, shouldApplyIncomingRunEvent, shouldBlockRuntimeSubmit, shouldIgnoreTerminalRuntimeEvent, shouldUpdateStreamStateForRunEvent } from './state'

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
  test('blocks a second submit while a selected run is pending or running', () => {
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'pending' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'running' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'completed' })).toBe(false)
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
    expect(next.assistantDraft).toMatchObject({ content: '片段', status: 'drafting' })
  })

  test('preserves normalized event identity when applying model gateway events', () => {
    const event = { id: 'evt-delta', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'hel', assistantDelta: 'hel', time: 'Now', status: 'running' } as const

    const next = applyModelGatewayEventToRun(run, event)

    expect(next.events[0]).toEqual(event)
  })

  test('applies model gateway delta and completion events to assistant draft', () => {
    const drafting = applyModelGatewayEventToRun(run, { id: 'evt-delta', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: 'hel', assistantDelta: 'hel', time: 'Now', status: 'running' })
    const completed = applyModelGatewayEventToRun(drafting, { id: 'evt-complete', runId: run.id, threadId: run.threadId, type: 'message.model_output_completed', label: 'message', detail: 'Model output completed', content: 'hello', time: 'Now', status: 'running' })

    expect(drafting.assistantDraft).toMatchObject({ content: 'hel', status: 'drafting' })
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
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'running' })).toBe(false)
  })
})
