import { describe, expect, test } from 'bun:test'
import type { Message, Run } from './domain'
import { createRuntimeEvent, getRuntimeScriptSteps } from './runtime/runtimeScripts'
import { appendRuntimeEventToRun, applyAssistantDeltaToRun, applyModelGatewayEventToRun, applyRunStreamEventToRun, createOptimisticSendSnapshot, createRegenerateAttemptRun, createRetryAttemptRun, createWorkspaceSettingsState, isOptimisticSendRun, minimumOptimisticThinkingMs, reconcileRunWithPersistedAssistant, shouldApplyIncomingRunEvent, shouldBlockRuntimeSubmit, shouldIgnoreTerminalRuntimeEvent, shouldPromoteThreadForWorkspaceSend, shouldReconcileTerminalStreamEvent, shouldUpdateStreamStateForRunEvent } from './state'

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

  test('creates an optimistic user message and pending draft before the backend run returns', () => {
    const snapshot = createOptimisticSendSnapshot({ threadId: 'thread-a', content: '继续', model: 'gpt-5.5' })

    expect(snapshot.messages).toMatchObject([{ threadId: 'thread-a', role: 'user', content: '继续' }])
    expect(snapshot.run).toMatchObject({
      threadId: 'thread-a',
      status: 'pending',
      model: 'gpt-5.5',
      assistantDraft: { content: '', status: 'pending' },
    })
    expect(isOptimisticSendRun(snapshot.run)).toBe(true)
    expect(shouldBlockRuntimeSubmit(snapshot.run)).toBe(true)
  })

  test('keeps the optimistic thinking state visible briefly before applying a fast final result', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()
    const sendStart = source.indexOf('const sendMessage = useCallback(async')
    const sendSource = source.slice(sendStart)

    expect(minimumOptimisticThinkingMs).toBeGreaterThanOrEqual(500)
    expect(sendSource).toContain('const optimisticStartedAt = Date.now()')
    expect(sendSource).toContain('await waitForMinimumOptimisticThinking(optimisticStartedAt)')
    expect(sendSource.indexOf('await waitForMinimumOptimisticThinking(optimisticStartedAt)')).toBeLessThan(sendSource.indexOf('setMessages(result.messages)'))
  })

  test('desktop workspace authorization promotes the current thread to work mode', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()
    const chooseFolder = source.indexOf('const chooseWorkspaceFolder = useCallback(async () => {')
    const saveWorkspace = source.indexOf('const config = await apiClient.saveWorkspaceRoot({ path: selected.path })', chooseFolder)
    const promoteThread = source.indexOf("await apiClient.updateThread(threadId, { mode: 'work' })", saveWorkspace)

    expect(chooseFolder).toBeGreaterThan(0)
    expect(saveWorkspace).toBeGreaterThan(chooseFolder)
    expect(promoteThread).toBeGreaterThan(saveWorkspace)
  })

  test('keeps sidebar thread changes from flashing through stale refreshes', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('const workspaceRefreshRequestRef = useRef(0)')
    expect(source).toContain('const requestID = workspaceRefreshRequestRef.current + 1')
    expect(source).toContain('workspaceRefreshRequestRef.current = requestID')
    expect(source).toContain('if (!shouldApplyLatestRequest(requestID, workspaceRefreshRequestRef.current)) return')
    expect(source).toContain('if (shouldApplyLatestRequest(requestID, workspaceRefreshRequestRef.current)) setLoading(false)')
  })

  test('archives selected threads without issuing a duplicate refresh before selection settles', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()
    const archiveStart = source.indexOf('const archiveThread = useCallback(async (threadId: string) => {')
    const archiveEnd = source.indexOf('const stopRun = useCallback(async () => {', archiveStart)
    const archiveSource = source.slice(archiveStart, archiveEnd)

    expect(archiveSource).toContain('if (!archivedSelectedThread) setThreads((current) => current.filter((thread) => thread.id !== threadId))')
    expect(archiveSource).toContain('threadSelectionRequestRef.current += 1')
    expect(archiveSource).toContain('apiClient.getThreadMessages(nextThreadId)')
    expect(archiveSource).toContain('apiClient.getThreadRun(nextThreadId)')
    expect(archiveSource).toContain('applyThreadSnapshot(nextThreadId, nextSnapshot)')
    expect(archiveSource.indexOf('if (archivedSelectedThread) setThreads((current) => current.filter((thread) => thread.id !== threadId))')).toBeGreaterThan(archiveSource.indexOf('await apiClient.archiveThread(threadId)'))
    expect(archiveSource).toContain('if (nextThreadId && !nextSnapshot)')
    expect(archiveSource).not.toContain('nextThreadId !== selectedThreadIdRef.current && !nextSnapshot')
    expect(archiveSource).not.toContain('await refresh(nextThreadId)')
  })

  test('defers uncached thread selection until the target snapshot is ready', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()
    const selectStart = source.indexOf('const selectThread = useCallback(async (threadId: string) => {')
    const selectEnd = source.indexOf('const sendMessage = useCallback(async', selectStart)
    const selectSource = source.slice(selectStart, selectEnd)
    const cachedSelection = selectSource.indexOf('if (cached) applyThreadSnapshot(threadId, cached)')
    const fetchMessages = selectSource.indexOf('apiClient.getThreadMessages(threadId)')
    const fetchRun = selectSource.indexOf('apiClient.getThreadRun(threadId)')
    const applySelection = selectSource.indexOf('applyThreadSnapshot(threadId, { messages: nextMessages, run: reconciledRun, artifacts: nextArtifacts })')

    expect(selectSource).toContain('threadSnapshotsRef')
    expect(selectSource).not.toContain('if (!cached) setSelectedThreadId(threadId)')
    expect(selectSource).not.toContain('skipNextSelectedThreadRefreshRef.current = true')
    expect(cachedSelection).toBeGreaterThan(0)
    expect(cachedSelection).toBeLessThan(fetchMessages)
    expect(fetchMessages).toBeGreaterThan(0)
    expect(fetchRun).toBeGreaterThan(fetchMessages)
    expect(applySelection).toBeGreaterThan(fetchRun)
  })

  test('creates a new thread only after its initial snapshot is ready', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()
    const createStart = source.indexOf("const createThread = useCallback(async (mode: Thread['mode'] = defaultWorkspaceMode) => {")
    const createEnd = source.indexOf('const renameThread = useCallback(async', createStart)
    const createSource = source.slice(createStart, createEnd)
    const fetchMessages = createSource.indexOf('apiClient.getThreadMessages(thread.id)')
    const fetchRun = createSource.indexOf('apiClient.getThreadRun(thread.id)')
    const cacheSnapshot = createSource.indexOf('threadSnapshotsRef.current.set(thread.id, nextSnapshot)')
    const applySnapshot = createSource.indexOf('applyThreadSnapshot(thread.id, nextSnapshot)')

    expect(createSource).not.toContain('setSelectedThreadId(thread.id)')
    expect(createSource).not.toContain('await refresh(thread.id)')
    expect(fetchMessages).toBeGreaterThan(0)
    expect(fetchRun).toBeGreaterThan(fetchMessages)
    expect(cacheSnapshot).toBeGreaterThan(fetchRun)
    expect(applySnapshot).toBeGreaterThan(cacheSnapshot)
  })

  test('skips the follow-up selected-thread refresh after applying a ready snapshot', async () => {
    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('const skipNextSelectedThreadRefreshRef = useRef(false)')
    expect(source).toContain('skipNextSelectedThreadRefreshRef.current = true')
    expect(source).toContain('if (skipNextSelectedThreadRefreshRef.current) {')
    expect(source).toContain('skipNextSelectedThreadRefreshRef.current = false')
    expect(source).toContain('return')
  })

  test('continues a workspace authorization flow in work mode after the folder is already granted', () => {
    expect(shouldPromoteThreadForWorkspaceSend({
      thread: { id: 'thread-chat', title: 'Downloads', project: 'Loomi', mode: 'chat', updatedAt: 'Now', lifecycleStatus: 'active' },
      workspaceRootConfig: { configured: true, displayName: 'Downloads' },
      content: '现在呢',
      messages: [
        { ...message, content: '看下我下载目录整理一下' },
        { ...message, id: 'msg-b', role: 'assistant', content: '请选择下载目录授权后我继续。' },
      ],
    })).toBe(true)
  })

  test('blocks a second submit while a selected run is pending, running, retrying, or recovering', () => {
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'pending' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'queued' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'running' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'retrying' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'recovering' })).toBe(true)
    expect(shouldBlockRuntimeSubmit({ ...run, status: 'blocked_on_tool_approval' })).toBe(true)
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

  test('dedupes replayed assistant delta when live stream repeats the same sequence', () => {
    const runningRun: Run = {
      ...run,
      status: 'running',
      events: [{ id: 'evt-replay', runId: run.id, threadId: run.threadId, sequence: 2, type: 'message.model_output_delta', label: 'Model', detail: 'hi', time: 'Replay', status: 'running', assistantDelta: 'hi' }],
      assistantDraft: { content: 'hi', status: 'streaming', lastEventId: 'evt-replay' },
    }

    const next = applyRunStreamEventToRun(runningRun, { id: 'evt-live-duplicate', runId: run.id, threadId: run.threadId, sequence: 2, type: 'message.model_output_delta', label: 'Model', detail: 'hi', time: 'Live', status: 'running', assistantDelta: 'hi' })

    expect(next.assistantDraft?.content).toBe('hi')
    expect(next.events.map((event) => event.id)).toEqual(['evt-replay'])
  })

  test('dedupes replayed tool event when live stream repeats the same sequence', () => {
    const runningRun: Run = {
      ...run,
      status: 'blocked_on_tool_approval',
      events: [{ id: 'evt-tool-replay', runId: run.id, threadId: run.threadId, sequence: 5, type: 'tool.call.approval_required', label: 'Tool', detail: 'approval', time: 'Replay', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'required', execution_status: 'blocked' } }],
      toolCalls: [{ id: 'tc_1', toolCallId: 'tc_1', name: 'runtime.get_current_time', status: 'approval_required', approvalStatus: 'required', executionStatus: 'blocked', argumentsSummary: {} }],
    }

    const next = applyRunStreamEventToRun(runningRun, { id: 'evt-tool-live-duplicate', runId: run.id, threadId: run.threadId, sequence: 5, type: 'tool.call.approval_required', label: 'Tool', detail: 'approval', time: 'Live', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', approval_status: 'required', execution_status: 'blocked' } })

    expect(next.events.map((event) => event.id)).toEqual(['evt-tool-replay'])
    expect(next.toolCalls).toHaveLength(1)
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

  test('reconciles persisted final messages when the stream reaches a terminal event', async () => {
    expect(shouldReconcileTerminalStreamEvent({ id: 'evt-running', runId: run.id, threadId: run.threadId, type: 'message.model_output_delta', label: 'message', detail: 'delta', content: 'hello', assistantDelta: 'hello', time: 'Now', status: 'running' })).toBe(false)
    expect(shouldReconcileTerminalStreamEvent({ id: 'evt-completed', runId: run.id, threadId: run.threadId, type: 'run.completed', label: 'run', detail: 'completed', time: 'Now', status: 'completed' })).toBe(true)

    const source = await Bun.file(new URL('./state.ts', import.meta.url)).text()

    expect(source).toContain('reconcileActiveRun({ allowAfterCleanup: true })')
    expect(source).toContain('cancelled && !options.allowAfterCleanup')
    expect(source).toContain('threadId !== selectedThreadIdRef.current')
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

  test('accepts late assistant final content after run completed terminal event', () => {
    const completedRun: Run = {
      ...run,
      status: 'completed',
      events: [{ id: 'evt-run-completed', runId: run.id, threadId: run.threadId, type: 'run.completed', label: 'Run', detail: 'done', time: 'Now', status: 'completed' }],
      assistantDraft: { content: 'collapsed draft', status: 'completed', lastEventId: 'evt-run-completed' },
    }
    const finalEvent = {
      id: 'evt-final-message',
      runId: run.id,
      threadId: run.threadId,
      type: 'message.model_output_completed',
      label: 'message',
      detail: 'Model output completed',
      content: 'formatted final',
      time: 'Later',
      status: 'completed',
    } as const

    const next = applyRunStreamEventToRun(completedRun, finalEvent)

    expect(next.assistantDraft).toMatchObject({ content: 'formatted final', status: 'completed', lastEventId: 'evt-final-message' })
    expect(next.events.map((event) => event.id)).toEqual(['evt-run-completed', 'evt-final-message'])
    expect(shouldApplyIncomingRunEvent(completedRun, finalEvent)).toBe(true)
  })

  test('uses persisted assistant message as completed run source of truth', () => {
    const completedRun: Run = {
      ...run,
      status: 'completed',
      assistantDraft: { content: '[redacted]', status: 'completed', lastEventId: 'evt-final' },
    }
    const messages: Message[] = [
      message,
      { id: 'msg-assistant', threadId: run.threadId, role: 'assistant', content: '## Final\n\n- rendered from message', createdAt: 'Now', runId: run.id },
    ]

    const reconciled = reconcileRunWithPersistedAssistant(completedRun, messages)

    expect(reconciled.assistantDraft).toEqual({ content: '## Final\n\n- rendered from message', status: 'completed', messageId: 'msg-assistant', lastEventId: 'evt-final' })
  })

  test('promotes live model output completion before run terminal event', () => {
    const runningRun: Run = {
      ...run,
      status: 'running',
      assistantDraft: { content: '##Final###1.collapsed', status: 'streaming', lastEventId: 'evt-delta' },
      events: [{ id: 'evt-delta', runId: run.id, threadId: run.threadId, sequence: 1, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', content: '##Final', assistantDelta: '##Final', time: 'Now', status: 'running' }],
    }
    const finalContent = '## Final\n\n### 1. Summary\n\n| Path | Meaning |\n| --- | --- |\n| `src` | Code |'
    const finalEvent = { id: 'evt-final-message', runId: run.id, threadId: run.threadId, sequence: 2, type: 'message.model_output_completed', label: 'message', detail: 'Model output completed', content: finalContent, time: 'Now', status: 'running' } as const
    const runCompleted = { id: 'evt-run-completed', runId: run.id, threadId: run.threadId, sequence: 3, type: 'run.completed', label: 'Run', detail: 'Run completed', content: null, time: 'Later', status: 'completed' } as const

    const completedWithFinalContent = applyRunStreamEventToRun(runningRun, finalEvent)
    const next = applyRunStreamEventToRun(completedWithFinalContent, runCompleted)

    expect(completedWithFinalContent.assistantDraft).toMatchObject({ content: finalContent, status: 'completed', lastEventId: 'evt-final-message' })
    expect(next.assistantDraft?.content).toBe(finalContent)
    expect(next.status).toBe('completed')
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

  test('replays approval-required tool stream events into tool-call view model', () => {
    const runningRun: Run = { ...run, status: 'running', events: [], toolCalls: [] }
    const events: Run['events'] = [
      { id: 'evt-tool-requested', runId: run.id, threadId: run.threadId, sequence: 1, type: 'tool.call.requested', label: 'Tool', detail: 'Tool call requested', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'required', execution_status: 'blocked' } },
      { id: 'evt-tool-required', runId: run.id, threadId: run.threadId, sequence: 2, type: 'tool.call.approval_required', label: 'Tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'required', execution_status: 'blocked' } },
    ]

    const next = events.reduce(applyRunStreamEventToRun, runningRun)

    expect(next.status).toBe('blocked_on_tool_approval')
    expect(next.toolCalls).toHaveLength(1)
    expect(next.toolCalls?.[0]).toMatchObject({ toolCallId: 'tc_1', name: 'runtime.get_current_time', status: 'approval_required', approvalStatus: 'required', executionStatus: 'blocked', argumentsSummary: { timezone: 'UTC' } })
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
