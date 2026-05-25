import { describe, expect, test } from 'bun:test'
import type { Message, Run, Thread } from './domain'
import { createWorkspaceSettingsState, mergeRunEvents, redactProviderCheckMessage, shouldApplyRunStreamEvent, getWorkspaceRefreshThreadId, shouldApplySendMessageResult, shouldApplyWorkspaceRefresh, shouldSelectWorkspaceRefreshThread } from './state'

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

describe('workspace settings state', () => {
  test('captures the default workspace mode for future local behavior', () => {
    expect(createWorkspaceSettingsState({ defaultWorkspaceMode: 'work' }).defaultWorkspaceMode).toBe('work')
    expect(createWorkspaceSettingsState().defaultWorkspaceMode).toBe('chat')
  })
})

describe('provider check state helpers', () => {
  test('redacts provider check errors before displaying them', () => {
    const message = redactProviderCheckMessage('Authorization: Bearer sk-secret123 api_key=secret token=hidden')

    expect(message).toContain('[redacted]')
    expect(message).not.toContain('sk-secret123')
    expect(message).not.toContain('secret')
    expect(message).not.toContain('hidden')
  })

  test('state hook exposes provider check action and result state', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('providerCheckResults')
    expect(source).toContain('checkProvider')
    expect(source).toContain('apiClient.checkModelProvider')
    expect(source).toContain('redactProviderCheckMessage')
  })
})

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

  test('state hook wires one real stream subscription per active run', () => {
    const source = Bun.file(new URL('./state.ts', import.meta.url)).text()
    return Promise.all([
      expect(source).resolves.toContain('apiClient.subscribeRunEvents'),
      expect(source).resolves.toContain('recoverable_error'),
      expect(source).resolves.toContain('mergeRunEvents'),
      expect(source).resolves.not.toContain('run?.events.length'),
    ])
  })

  test('dedupes streamed events by id and sequence without reordering arrivals', () => {
    const merged = mergeRunEvents([
      { id: 'evt-10', type: 'job_attempt_failed', label: 'Job', detail: 'Attempt failed', time: '1', status: 'failed', sequence: 10 },
    ], [
      { id: 'evt-10', type: 'job_attempt_failed', label: 'Job', detail: 'Attempt failed', time: '1', status: 'failed', sequence: 10 },
      { id: 'evt-2', type: 'job_claimed', label: 'Job', detail: 'Claimed after late delivery', time: '2', status: 'running', sequence: 2 },
    ])

    expect(merged.map((event) => event.id)).toEqual(['evt-10', 'evt-2'])
    expect(merged.at(-1)?.status).toBe('running')
  })
})
