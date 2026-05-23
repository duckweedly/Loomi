import { describe, expect, test } from 'bun:test'
import type { Message, Run, Thread } from './domain'
import { getWorkspaceRefreshThreadId, shouldApplySendMessageResult, shouldApplyWorkspaceRefresh } from './state'

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
