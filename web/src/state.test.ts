import { describe, expect, test } from 'bun:test'
import type { Message, Run, Thread } from './domain'
import { areThreadSnapshotsEqual, createWorkspaceSettingsState, getThreadIdAfterArchive, mergeRunEvents, redactProviderCheckMessage, shouldApplyLatestRequest, shouldApplyRunStreamEvent, getWorkspaceRefreshThreadId, shouldSendWorkspaceRefreshIntoLoading, shouldApplySendMessageResult, shouldApplyWorkspaceRefresh, shouldSelectWorkspaceRefreshThread } from './state'

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

  test('local provider detections require an explicit action and stay separate from configured providers', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('localProviderDetections')
    expect(source).toContain('detectLocalProviders')
    expect(source).toContain('enableLocalProvider')
    expect(source).toContain('disableLocalProvider')
    expect(source).toContain('apiClient.listLocalProviderDetections')
    expect(source).toContain('setLocalProviderDetections')
    expect(source).not.toContain('apiClient.listLocalProviderDetections()\\n      .then')
    expect(source).not.toContain('setProviderCapabilities(localProviderDetections')
  })

  test('local provider enablement refreshes configured providers without exposing secrets', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('apiClient.enableLocalProvider')
    expect(source).toContain('apiClient.disableLocalProvider')
    expect(source).toContain('setProviderCapabilities')
    expect(source).toContain('redactProviderCapabilityMessage')
    expect(source).not.toContain('access_token')
    expect(source).not.toContain('refresh_token')
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

describe('getThreadIdAfterArchive', () => {
  test('keeps the current selection when archiving another thread', () => {
    expect(getThreadIdAfterArchive({ archivedThreadId: 'thread-a', currentSelectedThreadId: 'thread-b', threads: [threadA, threadB] })).toBe('thread-b')
  })

  test('selects an adjacent thread when archiving the selected thread', () => {
    expect(getThreadIdAfterArchive({ archivedThreadId: 'thread-a', currentSelectedThreadId: 'thread-a', threads: [threadA, threadB] })).toBe('thread-b')
    expect(getThreadIdAfterArchive({ archivedThreadId: 'thread-b', currentSelectedThreadId: 'thread-b', threads: [threadA, threadB] })).toBe('thread-a')
  })
})

describe('areThreadSnapshotsEqual', () => {
  test('treats identical thread snapshots as stable to avoid redundant redraws', () => {
    expect(areThreadSnapshotsEqual({ messages: [messageA], run: runA }, { messages: [{ ...messageA }], run: { ...runA, events: [] } })).toBe(true)
  })

  test('detects changed message or run content before applying a fresh thread snapshot', () => {
    expect(areThreadSnapshotsEqual({ messages: [messageA], run: runA }, { messages: [{ ...messageA, content: 'B' }], run: runA })).toBe(false)
    expect(areThreadSnapshotsEqual({ messages: [messageA], run: runA }, { messages: [messageA], run: { ...runA, status: 'running' } })).toBe(false)
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

describe('shouldSendWorkspaceRefreshIntoLoading', () => {
  test('only uses the blocking loading state for the first empty workspace load', () => {
    expect(shouldSendWorkspaceRefreshIntoLoading({ threads: [], messages: [], run: null })).toBe(true)
    expect(shouldSendWorkspaceRefreshIntoLoading({ threads: [threadA], messages: [], run: null })).toBe(false)
    expect(shouldSendWorkspaceRefreshIntoLoading({ threads: [], messages: [messageA], run: null })).toBe(false)
    expect(shouldSendWorkspaceRefreshIntoLoading({ threads: [], messages: [], run: runA })).toBe(false)
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

describe('thread selection visual stability', () => {
  test('waits for an uncached thread snapshot before swapping the selected conversation', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).not.toContain("else {\n      skipNextSelectedThreadRefreshRef.current = true\n      setSelectedThreadId(threadId)\n    }")
    expect(source).not.toContain('if (!cached) setSelectedThreadId(threadId)')
    expect(source).toContain('if (cached) applyThreadSnapshot(threadId, cached)')
  })

  test('archives the selected thread after the next snapshot is ready', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('if (!archivedSelectedThread) setThreads((current) => current.filter((thread) => thread.id !== threadId))')
    expect(source).toContain('await apiClient.archiveThread(threadId)')
    expect(source).toContain('if (archivedSelectedThread) setThreads((current) => current.filter((thread) => thread.id !== threadId))')
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
      expect(source).resolves.toContain('reconcileActiveRun'),
      expect(source).resolves.toContain('window.setInterval'),
      expect(source).resolves.toContain('afterSequence: getMaxRunEventSequence'),
      expect(source).resolves.toContain('existingEvents: currentRun.events'),
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

describe('memory request freshness', () => {
  test('rejects older memory list audit and detail responses after a newer request starts', () => {
    expect(shouldApplyLatestRequest(1, 2)).toBe(false)
    expect(shouldApplyLatestRequest(2, 2)).toBe(true)
  })

  test('state hook guards memory list audit and detail async results', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('memoryEntriesRequestRef')
    expect(source).toContain('memoryAuditRequestRef')
    expect(source).toContain('memoryDetailRequestRef')
    expect(source).toContain('shouldApplyLatestRequest(requestID, memoryEntriesRequestRef.current)')
    expect(source).toContain('shouldApplyLatestRequest(requestID, memoryAuditRequestRef.current)')
    expect(source).toContain('shouldApplyLatestRequest(requestID, memoryDetailRequestRef.current)')
  })
})
