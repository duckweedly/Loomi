import { describe, expect, test } from 'bun:test'
import type { Run } from '../domain'
import { deriveChatCanvasState } from './chatCanvasState'

const baseRun: Run = {
  id: 'run-a',
  threadId: 'thread-a',
  status: 'running',
  model: 'Mock',
  context: 'Ready',
  events: [{ id: 'evt-a', type: 'context.loading', label: 'Context', detail: '加载上下文', time: 'Now', status: 'running' }],
}

describe('deriveChatCanvasState', () => {
  test('uses loading, error, backend unavailable, and no-thread priority before content states', () => {
    expect(deriveChatCanvasState({ loading: true, error: 'boom', selectedThreadId: 'thread-a', messageCount: 1, run: baseRun })).toBe('loading')
    expect(deriveChatCanvasState({ loading: false, error: 'boom', backendCapability: 'unavailable', backendUnavailableAttempted: true, selectedThreadId: 'thread-a', messageCount: 1, run: baseRun })).toBe('error')
    expect(deriveChatCanvasState({ loading: false, backendCapability: 'unavailable', backendUnavailableAttempted: true, selectedThreadId: 'thread-a', messageCount: 1, run: baseRun })).toBe('backend-unavailable')
    expect(deriveChatCanvasState({ loading: false, backendCapability: 'available', selectedThreadId: null, messageCount: 0 })).toBe('no-thread')
  })

  test('derives every non-error visible state from selected thread, messages, and run', () => {
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 0 })).toBe('empty-thread')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1 })).toBe('history')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'pending', events: [] } })).toBe('waiting-run')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: baseRun })).toBe('running')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 2, run: { ...baseRun, status: 'completed' } })).toBe('completed')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'failed' } })).toBe('failed')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'stopped' } })).toBe('stopped')
  })

  test('derives assistant draft states from the selected run draft', () => {
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'pending', events: [], assistantDraft: { content: '', status: 'pending' } } })).toBe('waiting-run')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, assistantDraft: { content: 'Partial', status: 'streaming' } } })).toBe('running')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 2, run: { ...baseRun, status: 'completed', assistantDraft: { content: 'Final', status: 'completed', messageId: 'msg-final' } } })).toBe('completed')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'failed', assistantDraft: { content: 'Partial', status: 'failed' } } })).toBe('failed')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'stopped', assistantDraft: { content: 'Partial', status: 'stopped' } } })).toBe('stopped')
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'recovering', assistantDraft: { content: 'Restored', status: 'recovering' } } })).toBe('recovering')
  })

  test('keeps idle completed runs as history instead of a completed runtime state', () => {
    expect(deriveChatCanvasState({ loading: false, selectedThreadId: 'thread-a', messageCount: 1, run: { ...baseRun, status: 'completed', events: [] } })).toBe('history')
  })

  test('shows backend unavailable only after a real-mode runtime attempt', () => {
    expect(deriveChatCanvasState({ loading: false, backendCapability: 'unavailable', selectedThreadId: 'thread-a', messageCount: 1 })).toBe('history')
    expect(deriveChatCanvasState({ loading: false, backendCapability: 'unavailable', backendUnavailableAttempted: true, selectedThreadId: 'thread-a', messageCount: 1 })).toBe('backend-unavailable')
  })
})
