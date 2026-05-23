import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { apiClient, executionAdapter } from './apiClient'
import { setMockRuntimeScript } from './mockApiClient'
import type { BackendCapabilityState, Message, Run, RunEvent, RuntimeEvent, RuntimeScriptId, StaleEventGuard, StreamState, Thread, ThreadRuntimeState } from './domain'
import { createNextThreadTitle } from './threadTitles'

type RefreshResult = {
  requestedThreadId: string
  currentSelectedThreadId: string
  threads: Thread[]
  messages: Message[]
  run: Run | null
}

export function getWorkspaceRefreshThreadId(requestedThreadId: string, threads: Thread[]) {
  if (!requestedThreadId) return threads[0]?.id || ''
  return threads.some((thread) => thread.id === requestedThreadId) ? requestedThreadId : threads[0]?.id || ''
}

export function shouldApplyWorkspaceRefresh(result: RefreshResult) {
  if (!result.requestedThreadId) return true
  return result.requestedThreadId === result.currentSelectedThreadId
}

export function shouldSelectWorkspaceRefreshThread({ requestedThreadId, resolvedThreadId, currentSelectedThreadId }: { requestedThreadId: string; resolvedThreadId: string; currentSelectedThreadId: string }) {
  return Boolean(resolvedThreadId) && resolvedThreadId !== requestedThreadId && requestedThreadId === currentSelectedThreadId
}

export function shouldApplySendMessageResult({ requestedThreadId, currentSelectedThreadId }: { requestedThreadId: string; currentSelectedThreadId: string }) {
  return requestedThreadId === currentSelectedThreadId
}

export function shouldApplyRunStreamEvent({ eventThreadId, eventRunId, selectedThreadId, currentRunId }: { eventThreadId: string; eventRunId: string; selectedThreadId: string; currentRunId: string }) {
  return eventThreadId === selectedThreadId && eventRunId === currentRunId
}

export function mergeRunEvents(existing: RunEvent[], incoming: RunEvent[]) {
  const byKey = new Map<string, RunEvent>()
  for (const event of [...existing, ...incoming]) {
    byKey.set(event.id || String(event.sequence), event)
  }
  return [...byKey.values()].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0))
}

export function createThreadRuntimeState(input: Partial<ThreadRuntimeState> = {}): ThreadRuntimeState {
  return {
    activeRunId: input.activeRunId ?? null,
    runsById: input.runsById ?? {},
    selectedScriptId: input.selectedScriptId ?? 'success',
    backendCapability: input.backendCapability ?? 'available',
    lastFailureReason: input.lastFailureReason,
  }
}

export function getActiveRuntimeRun(runtimeState: ThreadRuntimeState | null | undefined) {
  if (!runtimeState?.activeRunId) return null
  return runtimeState.runsById[runtimeState.activeRunId] ?? null
}

export function shouldApplyRuntimeEvent(guard: StaleEventGuard) {
  return guard.requestedThreadId === guard.currentSelectedThreadId && guard.runId === guard.activeRunId
}

export function createRuntimeStateForThread(backendCapability: BackendCapabilityState = 'available', selectedScriptId: RuntimeScriptId = 'success') {
  return createThreadRuntimeState({ backendCapability, selectedScriptId })
}

export function shouldBlockRuntimeSubmit(run: Run | null) {
  return run?.status === 'pending' || run?.status === 'running'
}

export function appendRuntimeEventToRun(run: Run, event: RuntimeEvent): Run {
  return {
    ...run,
    status: event.status,
    events: [...run.events, event],
    completedAt: event.status === 'completed' || event.status === 'failed' || event.status === 'stopped' ? event.time : run.completedAt,
  }
}

export function applyAssistantDeltaToRun(run: Run, delta: string): Run {
  const current = run.assistantDraft?.content ?? ''
  return {
    ...run,
    assistantDraft: {
      ...run.assistantDraft,
      content: `${current}${delta}`,
      status: 'drafting',
    },
  }
}

export function shouldIgnoreTerminalRuntimeEvent(run: Run) {
  return run.status === 'completed' || run.status === 'failed' || run.status === 'stopped'
}

export function useWorkspaceState() {
  const [threads, setThreads] = useState<Thread[]>([])
  const [selectedThreadId, setSelectedThreadId] = useState('thread-brief')
  const [messages, setMessages] = useState<Message[]>([])
  const [run, setRun] = useState<Run | null>(null)
  const [streamState, setStreamState] = useState<StreamState>('closed')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [backendUnavailableAttempted, setBackendUnavailableAttempted] = useState(false)
  const [selectedRuntimeScript, setSelectedRuntimeScript] = useState<RuntimeScriptId>('success')
  const selectedThreadIdRef = useRef(selectedThreadId)
  const runRef = useRef<Run | null>(run)

  selectedThreadIdRef.current = selectedThreadId
  runRef.current = run

  const selectedThread = useMemo(
    () => threads.find((thread) => thread.id === selectedThreadId) ?? null,
    [selectedThreadId, threads],
  )

  const refresh = useCallback(async (threadId = selectedThreadId) => {
    setLoading(true)
    setError(null)
    try {
      const nextThreads = await apiClient.listThreads()
      const nextThreadId = getWorkspaceRefreshThreadId(threadId, nextThreads)
      const [nextMessages, nextRun] = nextThreadId
        ? await Promise.all([apiClient.getThreadMessages(nextThreadId), apiClient.getThreadRun(nextThreadId)])
        : [[], null]
      if (!shouldApplyWorkspaceRefresh({ requestedThreadId: threadId, currentSelectedThreadId: selectedThreadIdRef.current, threads: nextThreads, messages: nextMessages, run: nextRun })) return
      setThreads(nextThreads)
      setMessages(nextMessages)
      setRun(nextRun)
      setStreamState(nextRun?.status === 'running' ? 'connecting' : 'closed')
      if (!threadId && nextThreadId) setSelectedThreadId(nextThreadId)
      else if (shouldSelectWorkspaceRefreshThread({ requestedThreadId: threadId, resolvedThreadId: nextThreadId, currentSelectedThreadId: selectedThreadIdRef.current })) setSelectedThreadId(nextThreadId)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setMessages([])
      setRun(null)
    } finally {
      setLoading(false)
    }
  }, [selectedThreadId])

  useEffect(() => {
    void refresh(selectedThreadId)
  }, [refresh, selectedThreadId])

  useEffect(() => {
    if (!run || run.status !== 'running' || !apiClient.subscribeRunEvents) {
      setStreamState((current) => {
        const next = run?.status === 'running' ? 'recoverable_error' : 'closed'
        return current === next ? current : next
      })
      return
    }
    setStreamState((current) => (current === 'connecting' ? current : 'connecting'))
    const afterSequence = run.events.at(-1)?.sequence ?? 0
    const unsubscribe = apiClient.subscribeRunEvents(
      run.id,
      afterSequence,
      (event) => {
        setRun((currentRun) => {
          if (!currentRun || !shouldApplyRunStreamEvent({ eventThreadId: event.threadId ?? '', eventRunId: event.runId ?? '', selectedThreadId: selectedThreadIdRef.current, currentRunId: currentRun.id })) return currentRun
          const status = event.status === 'running' ? currentRun.status : event.status
          const nextRun = { ...currentRun, status, events: mergeRunEvents(currentRun.events, [event]) }
          runRef.current = nextRun
          return nextRun
        })
        setStreamState((current) => {
          const next = event.status === 'running' ? 'live' : 'closed'
          return current === next ? current : next
        })
      },
      () => setStreamState((current) => (current === 'recoverable_error' ? current : 'recoverable_error')),
    )
    return unsubscribe
  }, [run?.id, run?.status])

  const selectThread = useCallback((threadId: string) => {
    setSelectedThreadId(threadId)
  }, [])

  const sendMessage = useCallback(async (content: string) => {
    const trimmed = content.trim()
    if (!trimmed) return
    const requestedThreadId = selectedThreadId
    setError(null)
    setBackendUnavailableAttempted(false)
    try {
      const result = await apiClient.sendMessage(requestedThreadId, trimmed)
      const nextThreads = await apiClient.listThreads()
      if (!shouldApplySendMessageResult({ requestedThreadId, currentSelectedThreadId: selectedThreadIdRef.current })) return
      setMessages(result.messages)
      setRun(result.run)
      setStreamState(result.run.status === 'running' ? 'connecting' : 'closed')
      setThreads(nextThreads)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
    }
  }, [selectedThreadId])

  const createThread = useCallback(async () => {
    if (!apiClient.createThread) return
    setError(null)
    try {
      const thread = await apiClient.createThread(createNextThreadTitle(threads), 'chat')
      setSelectedThreadId(thread.id)
      await refresh(thread.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
    }
  }, [refresh, threads])

  const renameThread = useCallback(async (threadId: string, title: string) => {
    if (!apiClient.updateThread) return
    setError(null)
    try {
      await apiClient.updateThread(threadId, { title })
      await refresh(threadId)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
    }
  }, [refresh])

  const archiveThread = useCallback(async (threadId: string) => {
    if (!apiClient.archiveThread) return
    setError(null)
    try {
      await apiClient.archiveThread(threadId)
      await refresh('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
    }
  }, [refresh])

  const stopRun = useCallback(async () => {
    if (!run || run.status !== 'running') return
    const stopped = await apiClient.stopRun(run.id)
    setRun(stopped)
    setStreamState('closed')
    setThreads(await apiClient.listThreads())
  }, [run])

  const selectRuntimeScript = useCallback((scriptId: RuntimeScriptId) => {
    setSelectedRuntimeScript(scriptId)
    setMockRuntimeScript(scriptId)
  }, [])

  return {
    threads,
    selectedThread,
    selectedThreadId,
    messages,
    run,
    streamState,
    loading,
    error,
    dataSourceMode: apiClient.mode,
    backendCapability: executionAdapter.runtimeCapability,
    backendUnavailableAttempted,
    selectedRuntimeScript,
    selectRuntimeScript,
    refresh,
    selectThread,
    createThread,
    renameThread,
    archiveThread,
    sendMessage,
    stopRun,
  }
}
