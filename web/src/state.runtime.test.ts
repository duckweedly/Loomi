import { describe, expect, test } from 'bun:test'
import type { Message, Run } from './domain'
import { appendRuntimeEventToRun, applyAssistantDeltaToRun, shouldBlockRuntimeSubmit, shouldIgnoreTerminalRuntimeEvent } from './state'

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

  test('ignores later script events after a terminal run event', () => {
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'failed' })).toBe(true)
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'stopped' })).toBe(true)
    expect(shouldIgnoreTerminalRuntimeEvent({ ...run, status: 'running' })).toBe(false)
  })
})
