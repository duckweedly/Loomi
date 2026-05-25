import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { apiClient, executionAdapter } from './apiClient'
import { setMockRuntimeScript } from './mockApiClient'
import type { BackendCapabilityState, MemoryEntry, Message, Persona, ProviderCapability, Run, RunEvent, RuntimeEvent, RuntimeScriptId, StaleEventGuard, StreamState, Thread, ThreadRuntimeState, ToolCall } from './domain'
import { isRuntimeActive, isRuntimeTerminal } from './runtime/executionAdapter'
import { deriveCapabilitySignalFromEvent } from './runtime/backendCapabilityStatus'
import { applyRealRunEvent, mapRealRuntimeCapabilitySignal } from './runtime/realExecutionAdapter'
import { createNextThreadTitle } from './threadTitles'

type RefreshResult = {
  requestedThreadId: string
  currentSelectedThreadId: string
  threads: Thread[]
  messages: Message[]
  run: Run | null
}

export type ProviderCheckStatus = 'idle' | 'checking' | 'success' | 'failed'

export type ProviderCheckResult = {
  status: ProviderCheckStatus
  message?: string
}

export type ProviderSaveStatus = 'idle' | 'saving' | 'success' | 'failed'

export type ProviderSaveResult = {
  status: ProviderSaveStatus
  message?: string
}

export function redactProviderCheckMessage(message: string) {
  const trimmed = message.trim()
  if (!trimmed) return 'Provider check failed'
  return trimmed
    .replace(/(authorization\s*[:=]\s*)(bearer\s+)?[^\s,;]+/gi, '$1[redacted]')
    .replace(/(api[_-]?key\s*[:=]\s*)[^\s,;]+/gi, '$1[redacted]')
    .replace(/(token\s*[:=]\s*)[^\s,;]+/gi, '$1[redacted]')
    .replace(/sk-[A-Za-z0-9_-]{8,}/g, '[redacted]')
}

function redactProviderCapabilityMessage(provider: ProviderCapability) {
  return provider.message ? { ...provider, message: redactProviderCheckMessage(provider.message) } : provider
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
  const indexesByKey = new Map<string, number>()
  const merged: RunEvent[] = []
  for (const event of [...existing, ...incoming]) {
    const key = event.id || String(event.sequence)
    const existingIndex = indexesByKey.get(key)
    if (existingIndex === undefined) {
      indexesByKey.set(key, merged.length)
      merged.push(event)
    } else {
      merged[existingIndex] = { ...merged[existingIndex], ...event }
    }
  }
  return merged
}

function getMaxRunEventSequence(events: RunEvent[], fallback: number) {
  return events.reduce((max, event) => (event.sequence === undefined ? max : Math.max(max, event.sequence)), fallback)
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
  return run ? isRuntimeActive(run.status) : false
}

export function appendRuntimeEventToRun(run: Run, event: RuntimeEvent): Run {
  if (isRuntimeTerminal(run.status)) return run

  return {
    ...run,
    status: event.status,
    events: [...run.events, event],
    completedAt: isRuntimeTerminal(event.status) ? event.time : run.completedAt,
  }
}

export function applyAssistantDeltaToRun(run: Run, delta: string, eventId?: string): Run {
  if (isRuntimeTerminal(run.status)) return run
  if (eventId && run.assistantDraft?.lastEventId === eventId) return run

  const current = run.assistantDraft?.content ?? ''
  return {
    ...run,
    assistantDraft: {
      ...run.assistantDraft,
      content: `${current}${delta}`,
      status: 'streaming',
      lastEventId: eventId ?? run.assistantDraft?.lastEventId,
    },
  }
}

export function applyModelGatewayEventToRun(run: Run, event: RuntimeEvent): Run {
  return applyRealRunEvent(run, event)
}

export function shouldApplyIncomingRunEvent(run: Run, event: RunEvent) {
  if (shouldIgnoreTerminalRuntimeEvent(run)) return false
  return !run.events.some((existing) => (existing.id || String(existing.sequence)) === (event.id || String(event.sequence)))
}

export function shouldUpdateStreamStateForRunEvent(run: Run, event: RunEvent) {
  return shouldApplyIncomingRunEvent(run, event)
}

export function shouldIgnoreTerminalRuntimeEvent(run: Run) {
  return isRuntimeTerminal(run.status)
}

export function applyRunStreamEventToRun(run: Run, event: RunEvent): Run {
  if (isRuntimeTerminal(run.status)) return run
  if (run.events.some((existing) => existing.id === event.id)) return run

  const maxSequence = getMaxRunEventSequence(run.events, -1)
  const shouldApplyAssistantDelta = !event.assistantDelta || event.sequence === undefined || maxSequence <= event.sequence
  let nextRun: Run = event.type.startsWith('tool.call.') ? applyRealRunEvent(run, { ...event, runId: event.runId ?? run.id, threadId: event.threadId ?? run.threadId }) : { ...run, events: mergeRunEvents(run.events, [event]) }
  if (event.assistantDelta && shouldApplyAssistantDelta) nextRun = applyAssistantDeltaToRun(nextRun, event.assistantDelta, event.id)

  if (event.status === 'running' || event.status === 'blocked_on_tool_approval') return { ...nextRun, status: event.status }
  if (event.status === 'completed') {
    return {
      ...nextRun,
      status: 'completed',
      completedAt: event.time,
      assistantDraft: {
        content: event.content ?? nextRun.assistantDraft?.content ?? '',
        status: 'completed',
        messageId: nextRun.assistantDraft?.messageId,
        lastEventId: event.id,
      },
    }
  }
  if (event.status === 'failed' || event.status === 'stopped') {
    return {
      ...nextRun,
      status: event.status,
      completedAt: event.time,
      assistantDraft: {
        content: nextRun.assistantDraft?.content ?? event.content ?? '',
        status: event.status,
        lastEventId: event.id,
      },
    }
  }
  if (event.status === 'recovering' || event.status === 'queued' || event.status === 'stopping') {
    return {
      ...nextRun,
      status: event.status,
      assistantDraft: {
        content: nextRun.assistantDraft?.content ?? event.content ?? '',
        status: event.status,
        lastEventId: event.id,
      },
    }
  }
  return { ...nextRun, status: event.status }
}

export function createRetryAttemptRun(failedRun: Run): Run {
  return {
    ...failedRun,
    id: `${failedRun.id}-retry`,
    status: 'pending',
    events: [],
    completedAt: undefined,
    assistantDraft: { content: '', status: 'pending' },
  }
}

export function createRegenerateAttemptRun(run: Run, attemptOfMessageId: string): Run {
  return {
    ...run,
    id: `${run.id}-regen`,
    status: 'pending',
    events: [],
    completedAt: undefined,
    attemptOfMessageId,
    assistantDraft: { content: '', status: 'pending' },
  }
}

export function createWorkspaceSettingsState(input: Partial<{ defaultWorkspaceMode: Thread['mode']; selectedRuntimeScript: RuntimeScriptId }> = {}) {
  return {
    defaultWorkspaceMode: input.defaultWorkspaceMode ?? 'chat',
    selectedRuntimeScript: input.selectedRuntimeScript ?? 'success',
  }
}

export function useWorkspaceState(defaultWorkspaceMode: Thread['mode'] = 'chat') {
  const [threads, setThreads] = useState<Thread[]>([])
  const [selectedThreadId, setSelectedThreadId] = useState('thread-brief')
  const [messages, setMessages] = useState<Message[]>([])
  const [run, setRun] = useState<Run | null>(null)
  const [streamState, setStreamState] = useState<StreamState>('closed')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [backendUnavailableAttempted, setBackendUnavailableAttempted] = useState(false)
  const [capabilitySignals, setCapabilitySignals] = useState({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
  const [selectedRuntimeScript, setSelectedRuntimeScript] = useState<RuntimeScriptId>('success')
  const [providerCapabilities, setProviderCapabilities] = useState<ProviderCapability[]>([])
  const [personas, setPersonas] = useState<Persona[]>([])
  const [selectedPersonaId, setSelectedPersonaId] = useState('')
  const [providerCheckResults, setProviderCheckResults] = useState<Record<string, ProviderCheckResult>>({})
  const [providerSaveResult, setProviderSaveResult] = useState<ProviderSaveResult>({ status: 'idle' })
  const [memoryEntries, setMemoryEntries] = useState<MemoryEntry[]>([])
  const [memoryQuery, setMemoryQuery] = useState('')
  const [memoryLoading, setMemoryLoading] = useState(false)
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
      setCapabilitySignals({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
      setStreamState(nextRun && shouldBlockRuntimeSubmit(nextRun) ? 'connecting' : 'closed')
      if (!threadId && nextThreadId) setSelectedThreadId(nextThreadId)
      else if (shouldSelectWorkspaceRefreshThread({ requestedThreadId: threadId, resolvedThreadId: nextThreadId, currentSelectedThreadId: selectedThreadIdRef.current })) setSelectedThreadId(nextThreadId)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
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
    if (!apiClient.listModelProviders) {
      setProviderCapabilities([])
      return
    }
    let cancelled = false
    apiClient.listModelProviders()
      .then((providers) => {
        if (!cancelled) setProviderCapabilities(providers.map(redactProviderCapabilityMessage))
      })
      .catch(() => {
        if (!cancelled) setProviderCapabilities([])
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!apiClient.listPersonas) {
      setPersonas([])
      setSelectedPersonaId('')
      return
    }
    let cancelled = false
    apiClient.listPersonas()
      .then((items) => {
        if (cancelled) return
        setPersonas(items)
        setSelectedPersonaId((current) => current || items.find((persona) => persona.isDefault)?.id || items[0]?.id || '')
      })
      .catch(() => {
        if (!cancelled) {
          setPersonas([])
          setSelectedPersonaId('')
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  const loadMemoryEntries = useCallback(async (query = '') => {
    if (!apiClient.listMemoryEntries || !apiClient.searchMemory) {
      setMemoryEntries([])
      return
    }
    setMemoryLoading(true)
    try {
      const entries = query.trim()
        ? await apiClient.searchMemory(query)
        : await apiClient.listMemoryEntries()
      setMemoryEntries(entries)
    } finally {
      setMemoryLoading(false)
    }
  }, [])

  const setMemorySearchQuery = useCallback((query: string) => {
    setMemoryQuery(query)
    void loadMemoryEntries(query)
  }, [loadMemoryEntries])

  const deleteMemoryEntry = useCallback(async (entryId: string) => {
    if (!apiClient.deleteMemoryEntry) return
    await apiClient.deleteMemoryEntry(entryId)
    await loadMemoryEntries(memoryQuery)
  }, [loadMemoryEntries, memoryQuery])

  useEffect(() => {
    void loadMemoryEntries('')
  }, [loadMemoryEntries])

  useEffect(() => {
    if (!run || !shouldBlockRuntimeSubmit(run) || !apiClient.subscribeRunEvents) {
      setStreamState((current) => {
        const next = run && shouldBlockRuntimeSubmit(run) ? 'recoverable_error' : 'closed'
        return current === next ? current : next
      })
      return
    }
    setStreamState((current) => (current === 'connecting' ? current : 'connecting'))
    const afterSequence = getMaxRunEventSequence(run.events, 0)
    const unsubscribe = apiClient.subscribeRunEvents(
      run.id,
      afterSequence,
      (event) => {
        setRun((currentRun) => {
          if (!currentRun || !shouldApplyRunStreamEvent({ eventThreadId: event.threadId ?? '', eventRunId: event.runId ?? '', selectedThreadId: selectedThreadIdRef.current, currentRunId: currentRun.id })) return currentRun
          const nextRun = applyRunStreamEventToRun(currentRun, event)
          runRef.current = nextRun
          return nextRun
        })
        setCapabilitySignals((current) => ({ ...current, ...deriveCapabilitySignalFromEvent(event), streamDisconnected: isRuntimeActive(event.status) ? current.streamDisconnected : false }))
        setStreamState((current) => {
          const next = isRuntimeActive(event.status) ? 'live' : 'closed'
          return current === next ? current : next
        })
      },
      () => {
        setCapabilitySignals((current) => ({ ...current, streamDisconnected: true }))
        setStreamState((current) => (current === 'recoverable_error' ? current : 'recoverable_error'))
      },
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
    setCapabilitySignals({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
    try {
      const result = await apiClient.sendMessage(requestedThreadId, trimmed, selectedPersonaId || undefined)
      const nextThreads = await apiClient.listThreads()
      if (!shouldApplySendMessageResult({ requestedThreadId, currentSelectedThreadId: selectedThreadIdRef.current })) return
      setMessages(result.messages)
      setRun(result.run)
      setCapabilitySignals({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
      setStreamState(shouldBlockRuntimeSubmit(result.run) ? 'connecting' : 'closed')
      setThreads(nextThreads)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
    }
  }, [selectedPersonaId, selectedThreadId])

  const createThread = useCallback(async () => {
    if (!apiClient.createThread) return
    setError(null)
    try {
      const thread = await apiClient.createThread(createNextThreadTitle(threads), defaultWorkspaceMode)
      setSelectedThreadId(thread.id)
      await refresh(thread.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
    }
  }, [defaultWorkspaceMode, refresh, threads])

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
    if (!run || !shouldBlockRuntimeSubmit(run)) return
    const stopped = await apiClient.stopRun(run.id)
    setRun(stopped)
    setCapabilitySignals((current) => ({ ...current, streamDisconnected: false }))
    setStreamState('closed')
    setThreads(await apiClient.listThreads())
  }, [run])

  const applyToolCallProjection = useCallback((toolCall: ToolCall) => {
    setRun((current) => {
      if (!current || current.id !== toolCall.id && current.id !== toolCall.toolCallId && current.id !== runRef.current?.id) return current
      const existing = current.toolCalls ?? []
      const index = existing.findIndex((candidate) => candidate.toolCallId === toolCall.toolCallId)
      const toolCalls = index >= 0 ? existing.map((candidate, itemIndex) => itemIndex === index ? toolCall : candidate) : [toolCall, ...existing]
      const next = { ...current, toolCalls }
      runRef.current = next
      return next
    })
  }, [])

  const approveToolCall = useCallback(async (toolCall: ToolCall) => {
    if (!run || !apiClient.approveToolCall) return
    const approved = await apiClient.approveToolCall(run.threadId, run.id, toolCall.toolCallId ?? toolCall.id)
    applyToolCallProjection(approved)
    setStreamState('connecting')
  }, [applyToolCallProjection, run])

  const denyToolCall = useCallback(async (toolCall: ToolCall) => {
    if (!run || !apiClient.denyToolCall) return
    const denied = await apiClient.denyToolCall(run.threadId, run.id, toolCall.toolCallId ?? toolCall.id)
    applyToolCallProjection(denied)
  }, [applyToolCallProjection, run])

  const retryRun = useCallback(async () => {
    if (!run || run.status !== 'failed') return
    setError(null)
    try {
      if (apiClient.startRun) {
        const nextRun = await apiClient.startRun(run.threadId)
        setRun(nextRun)
      } else {
        setRun(createRetryAttemptRun(run))
      }
      setCapabilitySignals((current) => ({ ...current, streamDisconnected: false }))
      setStreamState('connecting')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
    }
  }, [run])

  const regenerateRun = useCallback(async () => {
    const lastAssistant = [...messages].reverse().find((message) => message.role === 'assistant')
    if (!run || !lastAssistant || shouldBlockRuntimeSubmit(run)) return
    setError(null)
    try {
      if (apiClient.startRun) {
        const nextRun = await apiClient.startRun(run.threadId)
        setRun({ ...nextRun, attemptOfMessageId: lastAssistant.id })
      } else {
        setRun(createRegenerateAttemptRun(run, lastAssistant.id))
      }
      setCapabilitySignals((current) => ({ ...current, streamDisconnected: false }))
      setStreamState('connecting')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
    }
  }, [messages, run])

  const selectRuntimeScript = useCallback((scriptId: RuntimeScriptId) => {
    setSelectedRuntimeScript(scriptId)
    setMockRuntimeScript(scriptId)
  }, [])

  const checkProvider = useCallback(async (providerId: string) => {
    if (!apiClient.checkModelProvider) return
    setProviderCheckResults((current) => ({ ...current, [providerId]: { status: 'checking' } }))
    try {
      const provider = redactProviderCapabilityMessage(await apiClient.checkModelProvider(providerId))
      setProviderCapabilities((current) => current.map((candidate) => (candidate.id === provider.id ? provider : candidate)))
      setProviderCheckResults((current) => ({
        ...current,
        [providerId]: {
          status: provider.status === 'available' ? 'success' : 'failed',
          message: provider.message ?? provider.status,
        },
      }))
    } catch (err) {
      setProviderCheckResults((current) => ({
        ...current,
        [providerId]: {
          status: 'failed',
          message: redactProviderCheckMessage(err instanceof Error ? err.message : 'Provider check failed'),
        },
      }))
    }
  }, [])

  const saveProvider = useCallback(async (input: { baseUrl: string; model: string; apiKey: string }) => {
    if (!apiClient.saveModelProvider) return
    setProviderSaveResult({ status: 'saving' })
    try {
      const provider = redactProviderCapabilityMessage(await apiClient.saveModelProvider(input))
      setProviderCapabilities((current) => {
        const exists = current.some((candidate) => candidate.id === provider.id)
        return exists ? current.map((candidate) => (candidate.id === provider.id ? provider : candidate)) : [...current, provider]
      })
      setProviderCheckResults((current) => ({ ...current, [provider.id]: { status: provider.status === 'available' ? 'success' : 'failed', message: provider.message ?? provider.status } }))
      setProviderSaveResult({ status: provider.status === 'available' ? 'success' : 'failed', message: provider.message ?? provider.status })
    } catch (err) {
      setProviderSaveResult({ status: 'failed', message: redactProviderCheckMessage(err instanceof Error ? err.message : 'Provider save failed') })
    }
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
    backendUnavailableAttempted: backendUnavailableAttempted || capabilitySignals.backendUnavailable,
    capabilitySignals,
    selectedRuntimeScript,
    providerCapabilities,
    personas,
    selectedPersonaId,
    providerCheckResults,
    providerSaveResult,
    memoryEntries,
    memoryQuery,
    memoryLoading,
    selectRuntimeScript,
    setSelectedPersonaId,
    checkProvider,
    saveProvider,
    setMemorySearchQuery,
    deleteMemoryEntry,
    refresh,
    selectThread,
    createThread,
    renameThread,
    archiveThread,
    sendMessage,
    stopRun,
    approveToolCall,
    denyToolCall,
    retryRun,
    regenerateRun,
  }
}
