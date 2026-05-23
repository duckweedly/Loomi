import { describe, expect, test } from 'bun:test'
import type { Message, Run, Thread } from './domain'
import { mergeRunEvents, shouldApplyRunStreamEvent, getWorkspaceRefreshThreadId, shouldApplySendMessageResult, shouldApplyWorkspaceRefresh, shouldSelectWorkspaceRefreshThread } from './state'

const threadA: Thread = {
  id: 'thread-a',
  title: 'Thread A',
  project: 'Loomi',
  mode: 'chat',
  updatedAt: '2026-05-23T00:00:00Z',
  lifecycleStatus: 'active',
  runStatus: 'completed',
}

const threadB: Thread = {
  ...threadA,
  id: 'thread-b',
  title: 'Thread B',
}

const runA: Run = {
  id: 'run-a',
  threadId: 'thread-a',
  status: 'completed',
  model: 'Deferred',
  context: 'test',
  events: [],
}

const messageA: Message = {
  id: 'msg-a',
  threadId: 'thread-a',
  role: 'user',
  content: 'A',
  createdAt: '2026-05-23T00:00:00Z',
}

describe('getWorkspaceRefreshThreadId', () => {
  test('falls back to the first returned thread when the requested thread is missing', () => {
    expect(getWorkspaceRefreshThreadId('thread-brief', [threadA, threadB])).toBe('thread-a')
  })

  test('keeps the requested thread when it exists in returned threads', () => {
    expect(getWorkspaceRefreshThreadId('thread-b', [threadA, threadB])).toBe('thread-b')
  })
})

describe('shouldSelectWorkspaceRefreshThread', () => {
  test('selects the resolved thread when the requested id is missing but still current', () => {
    expect(shouldSelectWorkspaceRefreshThread({ requestedThreadId: 'thread-brief', resolvedThreadId: 'thread-a', currentSelectedThreadId: 'thread-brief' })).toBe(true)
  })

  test('does not replace selection after the user has switched threads', () => {
    expect(shouldSelectWorkspaceRefreshThread({ requestedThreadId: 'thread-brief', resolvedThreadId: 'thread-a', currentSelectedThreadId: 'thread-b' })).toBe(false)
  })
})

describe('shouldApplyWorkspaceRefresh', () => {
  test('rejects stale refresh results for an older selected thread', () => {
    expect(
      shouldApplyWorkspaceRefresh({
        requestedThreadId: 'thread-a',
        currentSelectedThreadId: 'thread-b',
        threads: [threadA, threadB],
        messages: [messageA],
        run: runA,
      }),
    ).toBe(false)
  })

  test('allows initial refresh to choose the first returned thread', () => {
    expect(
      shouldApplyWorkspaceRefresh({
        requestedThreadId: '',
        currentSelectedThreadId: '',
        threads: [threadA],
        messages: [messageA],
        run: runA,
      }),
    ).toBe(true)
  })
})

describe('shouldApplySendMessageResult', () => {
  test('rejects a send result when the user has switched threads', () => {
    expect(shouldApplySendMessageResult({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-b' })).toBe(false)
  })

  test('allows a send result for the current selected thread', () => {
    expect(shouldApplySendMessageResult({ requestedThreadId: 'thread-a', currentSelectedThreadId: 'thread-a' })).toBe(true)
  })
})

describe('run stream state helpers', () => {
  test('rejects stale stream events from an older run or thread', () => {
    expect(shouldApplyRunStreamEvent({ eventThreadId: 'thread-a', eventRunId: 'run-a', selectedThreadId: 'thread-b', currentRunId: 'run-a' })).toBe(false)
    expect(shouldApplyRunStreamEvent({ eventThreadId: 'thread-a', eventRunId: 'run-a', selectedThreadId: 'thread-a', currentRunId: 'run-b' })).toBe(false)
    expect(shouldApplyRunStreamEvent({ eventThreadId: 'thread-a', eventRunId: 'run-a', selectedThreadId: 'thread-a', currentRunId: 'run-a' })).toBe(true)
  })

  test('state hook wires the real stream subscription and recoverable error state', () => {
    const source = Bun.file(new URL('./state.ts', import.meta.url)).text()
    return Promise.all([
      expect(source).resolves.toContain('apiClient.subscribeRunEvents'),
      expect(source).resolves.toContain('recoverable_error'),
      expect(source).resolves.toContain('mergeRunEvents'),
    ])
  })

  test('dedupes streamed events by id and sequence', () => {
    const merged = mergeRunEvents([
      { id: 'evt-1', type: 'lifecycle.run_created', label: 'lifecycle', detail: 'Run created', time: '1', status: 'running', sequence: 1 },
    ], [
      { id: 'evt-1', type: 'lifecycle.run_created', label: 'lifecycle', detail: 'Run created', time: '1', status: 'running', sequence: 1 },
      { id: 'evt-2', type: 'final.run_completed', label: 'final', detail: 'Run completed', time: '2', status: 'completed', sequence: 2 },
    ])

    expect(merged.map((event) => event.id)).toEqual(['evt-1', 'evt-2'])
    expect(merged.at(-1)?.status).toBe('completed')
  })
})
