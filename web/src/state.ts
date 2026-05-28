import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { apiClient, executionAdapter } from './apiClient'
import { setMockRuntimeScript } from './mockApiClient'
import type { BackendCapabilityState, InstalledSkill, LocalProviderDetection, MCPServerConfigInput, MCPServerStatus, MemoryAuditItem, MemoryEntry, MemoryErrorEvent, MemoryFilters, MemoryImpressionSnapshot, MemoryOverviewSnapshot, MemoryProviderStatus, MemoryProviderUpdate, MemoryWriteProposal, Message, Persona, ProviderCapability, Run, RunEvent, RuntimeEvent, RuntimeScriptId, StaleEventGuard, StreamState, Thread, ThreadRuntimeState, ToolCall, ToolCatalogItem, WebSearchConfig, WorkspaceRootConfig } from './domain'
import { isRuntimeActive, isRuntimeTerminal } from './runtime/executionAdapter'
import { deriveCapabilitySignalFromEvent } from './runtime/backendCapabilityStatus'
import { deriveDesktopReadiness } from './runtime/desktopReadiness'
import { applyRealRunEvent, mapRealRuntimeCapabilitySignal } from './runtime/realExecutionAdapter'
import { selectSendProvider } from './realApiClient'
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

function runMessageId(run: Run, messages: Message[]) {
  for (const event of run.events) {
    const messageId = event.metadata?.message_id
    if (typeof messageId === 'string' && messageId.trim()) return messageId
  }
  return [...messages].reverse().find((message) => message.role === 'user')?.id
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

export function getThreadIdAfterArchive({ archivedThreadId, currentSelectedThreadId, threads }: { archivedThreadId: string; currentSelectedThreadId: string; threads: Thread[] }) {
  if (archivedThreadId !== currentSelectedThreadId) return currentSelectedThreadId
  const archivedIndex = threads.findIndex((thread) => thread.id === archivedThreadId)
  const remaining = threads.filter((thread) => thread.id !== archivedThreadId)
  if (remaining.length === 0) return ''
  return remaining[Math.min(Math.max(archivedIndex, 0), remaining.length - 1)]?.id || ''
}

export function shouldApplySendMessageResult({ requestedThreadId, currentSelectedThreadId }: { requestedThreadId: string; currentSelectedThreadId: string }) {
  return requestedThreadId === currentSelectedThreadId
}

export function shouldApplyRunStreamEvent({ eventThreadId, eventRunId, selectedThreadId, currentRunId }: { eventThreadId: string; eventRunId: string; selectedThreadId: string; currentRunId: string }) {
  return eventThreadId === selectedThreadId && eventRunId === currentRunId
}

export function mergeRunEvents(existing: RunEvent[], incoming: RunEvent[]) {
  const indexesById = new Map<string, number>()
  const indexesBySequence = new Map<number, number>()
  const merged: RunEvent[] = []
  for (const event of [...existing, ...incoming]) {
    const existingIndex = event.id && indexesById.has(event.id)
      ? indexesById.get(event.id)
      : event.sequence !== undefined
        ? indexesBySequence.get(event.sequence)
        : undefined
    if (existingIndex === undefined) {
      if (event.id) indexesById.set(event.id, merged.length)
      if (event.sequence !== undefined) indexesBySequence.set(event.sequence, merged.length)
      merged.push(event)
    } else if (event.id && merged[existingIndex]?.id === event.id) {
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

export function shouldApplyLatestRequest(requestID: number, latestRequestID: number) {
  return requestID === latestRequestID
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
  if (shouldIgnoreTerminalRuntimeEvent(run) && !isAssistantFinalContentEvent(event)) return false
  return !hasRunEventIdentity(run, event)
}

export function shouldUpdateStreamStateForRunEvent(run: Run, event: RunEvent) {
  return shouldApplyIncomingRunEvent(run, event)
}

export function shouldIgnoreTerminalRuntimeEvent(run: Run) {
  return isRuntimeTerminal(run.status)
}

function isAssistantFinalContentEvent(event: RunEvent) {
  return Boolean(event.content?.trim()) && (event.type === 'assistant.message.completed' || event.type === 'message.model_output_completed' || event.type === 'model.final')
}

export function applyRunStreamEventToRun(run: Run, event: RunEvent): Run {
  if (isRuntimeTerminal(run.status)) {
    if (run.status !== 'completed') return run
    if (!isAssistantFinalContentEvent(event) || hasRunEventIdentity(run, event)) return run
    return {
      ...run,
      events: mergeRunEvents(run.events, [event]),
      assistantDraft: {
        content: event.content ?? run.assistantDraft?.content ?? '',
        status: 'completed',
        messageId: run.assistantDraft?.messageId,
        lastEventId: event.id,
      },
    }
  }
  if (hasRunEventIdentity(run, event)) return run

  const maxSequence = getMaxRunEventSequence(run.events, -1)
  const shouldApplyAssistantDelta = !event.assistantDelta || event.sequence === undefined || maxSequence <= event.sequence
  let nextRun: Run = event.type.startsWith('tool.call.') ? applyRealRunEvent(run, { ...event, runId: event.runId ?? run.id, threadId: event.threadId ?? run.threadId }) : { ...run, events: mergeRunEvents(run.events, [event]) }
  if (event.assistantDelta && shouldApplyAssistantDelta) nextRun = applyAssistantDeltaToRun(nextRun, event.assistantDelta, event.id)
  if (isAssistantFinalContentEvent(event)) {
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

export function reconcileRunWithPersistedAssistant(run: Run, messages: Message[]): Run {
  if (!isRuntimeTerminal(run.status)) return run
  const assistant = [...messages].reverse().find((message) => (
    message.role === 'assistant'
    && message.threadId === run.threadId
    && message.runId === run.id
    && message.content.trim()
  ))
  if (!assistant) return run
  const current = run.assistantDraft
  if (current?.content === assistant.content && current.messageId === assistant.id && current.status === 'completed') return run
  return {
    ...run,
    assistantDraft: {
      content: assistant.content,
      status: 'completed',
      messageId: assistant.id,
      lastEventId: current?.lastEventId,
    },
  }
}

function hasRunEventIdentity(run: Run, event: RunEvent) {
  return run.events.some((existing) => {
    if (event.id && existing.id === event.id) return true
    return event.sequence !== undefined && existing.sequence === event.sequence
  })
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

function memoryContextForEntry(entry: MemoryEntry): MemoryFilters {
  const context = {
    scopeType: entry.scopeType,
    scopeId: entry.scopeId,
    sourceThreadId: entry.sourceThreadId,
    sourceRunId: entry.sourceRunId,
    sourceType: entry.sourceType,
  }
  if (entry.scopeType === 'thread' && !context.scopeId && !context.sourceThreadId && !context.sourceRunId) {
    throw new Error('Memory action needs thread or source context')
  }
  return context
}

type DesktopFolderSelection = { canceled?: boolean; path?: string }

declare global {
  interface Window {
    loomiDesktop?: {
      selectWorkspaceFolder?: () => Promise<DesktopFolderSelection>
    }
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
  const [apiConnected, setApiConnected] = useState(apiClient.mode !== 'real_api')
  const [dbReady, setDBReady] = useState(true)
  const [capabilitySignals, setCapabilitySignals] = useState({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
  const [selectedRuntimeScript, setSelectedRuntimeScript] = useState<RuntimeScriptId>('success')
  const [providerCapabilities, setProviderCapabilities] = useState<ProviderCapability[]>([])
  const [toolCatalog, setToolCatalog] = useState<ToolCatalogItem[]>([])
  const [toolCatalogLoaded, setToolCatalogLoaded] = useState(false)
  const [webSearchConfig, setWebSearchConfig] = useState<WebSearchConfig | null>(null)
  const [webSearchSaveResult, setWebSearchSaveResult] = useState<ProviderSaveResult>({ status: 'idle' })
  const [workspaceRootConfig, setWorkspaceRootConfig] = useState<WorkspaceRootConfig | null>(null)
  const [workspaceRootSaveResult, setWorkspaceRootSaveResult] = useState<ProviderSaveResult>({ status: 'idle' })
  const [mcpServers, setMCPServers] = useState<MCPServerStatus[]>([])
  const [mcpActionResult, setMCPActionResult] = useState<ProviderSaveResult>({ status: 'idle' })
  const [localProviderDetections, setLocalProviderDetections] = useState<LocalProviderDetection[]>([])
  const [localProviderDetectionError, setLocalProviderDetectionError] = useState<string | null>(null)
  const [personas, setPersonas] = useState<Persona[]>([])
  const [installedSkills, setInstalledSkills] = useState<InstalledSkill[]>([])
  const [skillsLoading, setSkillsLoading] = useState(false)
  const [skillsError, setSkillsError] = useState<string | null>(null)
  const [selectedPersonaId, setSelectedPersonaId] = useState('')
  const [providerCheckResults, setProviderCheckResults] = useState<Record<string, ProviderCheckResult>>({})
  const [providerSaveResult, setProviderSaveResult] = useState<ProviderSaveResult>({ status: 'idle' })
  const [memoryEntries, setMemoryEntries] = useState<MemoryEntry[]>([])
  const [memoryQuery, setMemoryQuery] = useState('')
  const [memoryFilters, setMemoryFilters] = useState<MemoryFilters>({ limit: 20 })
  const [memoryLoading, setMemoryLoading] = useState(false)
  const [memoryError, setMemoryError] = useState<string | null>(null)
  const [memoryDetail, setMemoryDetail] = useState<MemoryEntry | null>(null)
  const [memoryDetailLoading, setMemoryDetailLoading] = useState(false)
  const [memoryDetailError, setMemoryDetailError] = useState<string | null>(null)
  const [memoryAuditItems, setMemoryAuditItems] = useState<MemoryAuditItem[]>([])
  const [memoryAuditLoading, setMemoryAuditLoading] = useState(false)
  const [memoryAuditError, setMemoryAuditError] = useState<string | null>(null)
  const [memoryWriteProposals, setMemoryWriteProposals] = useState<MemoryWriteProposal[]>([])
  const [memoryProposalsLoading, setMemoryProposalsLoading] = useState(false)
  const [memoryProposalsError, setMemoryProposalsError] = useState<string | null>(null)
  const [memoryProviderStatus, setMemoryProviderStatus] = useState<MemoryProviderStatus | null>(null)
  const [memoryErrors, setMemoryErrors] = useState<MemoryErrorEvent[]>([])
  const [memoryProviderSaveResult, setMemoryProviderSaveResult] = useState<ProviderSaveResult>({ status: 'idle' })
  const [memoryOverviewSnapshot, setMemoryOverviewSnapshot] = useState<MemoryOverviewSnapshot | null>(null)
  const [memoryImpressionSnapshot, setMemoryImpressionSnapshot] = useState<MemoryImpressionSnapshot | null>(null)
  const [memorySnapshotLoading, setMemorySnapshotLoading] = useState(false)
  const [pendingDeleteMemoryEntry, setPendingDeleteMemoryEntry] = useState<MemoryEntry | null>(null)
  const selectedThreadIdRef = useRef(selectedThreadId)
  const runRef = useRef<Run | null>(run)
  const memoryEntriesRequestRef = useRef(0)
  const memoryAuditRequestRef = useRef(0)
  const memoryProposalsRequestRef = useRef(0)
  const memoryDetailRequestRef = useRef(0)

  selectedThreadIdRef.current = selectedThreadId
  runRef.current = run

  const selectedThread = useMemo(
    () => threads.find((thread) => thread.id === selectedThreadId) ?? null,
    [selectedThreadId, threads],
  )

  const desktopReadiness = useMemo(() => deriveDesktopReadiness({
    apiConnected,
    dbReady,
    providerCapabilities,
    localProviderDetections,
    toolCatalog,
    toolCatalogLoaded,
    workspaceRootConfig,
  }), [apiConnected, dbReady, localProviderDetections, providerCapabilities, toolCatalog, toolCatalogLoaded, workspaceRootConfig])

  const refreshDesktopReadiness = useCallback(async () => {
    if (!apiClient.getReadiness) {
      setApiConnected(true)
      setDBReady(true)
      return
    }
    try {
      const readiness = await apiClient.getReadiness()
      setApiConnected(true)
      setDBReady(readiness.status === 'ready' && !readiness.checks.some((check) => check.status === 'failed'))
      setCapabilitySignals((current) => ({ ...current, backendUnavailable: false }))
    } catch (err) {
      setApiConnected(false)
      setDBReady(false)
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err), backendUnavailable: true }))
    }
  }, [])

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
      const reconciledRun = nextRun ? reconcileRunWithPersistedAssistant(nextRun, nextMessages) : null
      setThreads(nextThreads)
      setMessages(nextMessages)
      setRun(reconciledRun)
      setApiConnected(true)
      setDBReady(true)
      setCapabilitySignals({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
      setStreamState(reconciledRun && shouldBlockRuntimeSubmit(reconciledRun) ? 'connecting' : 'closed')
      if (!threadId && nextThreadId) setSelectedThreadId(nextThreadId)
      else if (shouldSelectWorkspaceRefreshThread({ requestedThreadId: threadId, resolvedThreadId: nextThreadId, currentSelectedThreadId: selectedThreadIdRef.current })) setSelectedThreadId(nextThreadId)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setApiConnected(false)
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
      setMessages([])
      setRun(null)
    } finally {
      setLoading(false)
    }
  }, [selectedThreadId])

  useEffect(() => {
    void refreshDesktopReadiness()
  }, [refreshDesktopReadiness])

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
        if (!cancelled) {
          setProviderCapabilities(providers.map(redactProviderCapabilityMessage))
          setApiConnected(true)
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Provider list request failed')
          setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  const saveMCPServer = useCallback(async (input: MCPServerConfigInput) => {
    if (!apiClient.saveMCPServer) return
    setMCPActionResult({ status: 'saving' })
    try {
      const server = await apiClient.saveMCPServer(input)
      setMCPServers((current) => [...current.filter((item) => item.serverSlug !== server.serverSlug), server].sort((a, b) => a.serverSlug.localeCompare(b.serverSlug)))
      setMCPActionResult({ status: 'success', message: 'Saved' })
    } catch (err) {
      setMCPActionResult({ status: 'failed', message: err instanceof Error ? redactProviderCheckMessage(err.message) : 'MCP save failed' })
    }
  }, [])

  const deleteMCPServer = useCallback(async (slug: string) => {
    if (!apiClient.deleteMCPServer) return
    setMCPActionResult({ status: 'saving' })
    try {
      setMCPServers(await apiClient.deleteMCPServer(slug))
      setMCPActionResult({ status: 'success', message: 'Deleted' })
    } catch (err) {
      setMCPActionResult({ status: 'failed', message: err instanceof Error ? redactProviderCheckMessage(err.message) : 'MCP delete failed' })
    }
  }, [])

  const discoverMCPServer = useCallback(async (slug: string) => {
    if (!apiClient.discoverMCPServer) return
    setMCPActionResult({ status: 'saving' })
    try {
      const server = await apiClient.discoverMCPServer(slug)
      setMCPServers((current) => current.map((item) => (item.serverSlug === slug ? server : item)))
      setMCPActionResult({ status: server.discoveryStatus === 'succeeded' ? 'success' : 'failed', message: server.discoveryStatus })
    } catch (err) {
      setMCPActionResult({ status: 'failed', message: err instanceof Error ? redactProviderCheckMessage(err.message) : 'MCP discovery failed' })
    }
  }, [])

  useEffect(() => {
    if (!apiClient.listMCPServers) {
      setMCPServers([])
      return
    }
    let cancelled = false
    apiClient.listMCPServers()
      .then((servers) => {
        if (!cancelled) setMCPServers(servers)
      })
      .catch(() => {
        if (!cancelled) setMCPServers([])
      })
    return () => {
      cancelled = true
    }
  }, [])

  const detectLocalProviders = useCallback(async () => {
    if (!apiClient.listLocalProviderDetections) {
      setLocalProviderDetections([])
      setLocalProviderDetectionError('Local provider detection endpoint unavailable')
      return
    }
    setLocalProviderDetectionError(null)
    try {
      setLocalProviderDetections(await apiClient.listLocalProviderDetections())
    } catch (err) {
      setLocalProviderDetections([])
      setLocalProviderDetectionError(err instanceof Error ? redactProviderCheckMessage(err.message) : 'Local provider detection unavailable')
    }
  }, [])

  useEffect(() => {
    void detectLocalProviders()
  }, [detectLocalProviders])

  const enableLocalProvider = useCallback(async (providerId: string) => {
    if (!apiClient.enableLocalProvider) {
      setLocalProviderDetectionError('Local provider enable endpoint unavailable')
      return
    }
    setLocalProviderDetectionError(null)
    try {
      const provider = redactProviderCapabilityMessage(await apiClient.enableLocalProvider(providerId))
      setProviderCapabilities((current) => {
        const exists = current.some((candidate) => candidate.id === provider.id)
        return exists ? current.map((candidate) => (candidate.id === provider.id ? provider : candidate)) : [...current, provider]
      })
    } catch (err) {
      setLocalProviderDetectionError(err instanceof Error ? redactProviderCheckMessage(err.message) : 'Local provider enable failed')
    }
  }, [])

  const disableLocalProvider = useCallback(async (providerId: string) => {
    if (!apiClient.disableLocalProvider) {
      setLocalProviderDetectionError('Local provider disable endpoint unavailable')
      return
    }
    setLocalProviderDetectionError(null)
    try {
      await apiClient.disableLocalProvider(providerId)
      setProviderCapabilities((current) => current.filter((provider) => provider.id !== providerId))
    } catch (err) {
      setLocalProviderDetectionError(err instanceof Error ? redactProviderCheckMessage(err.message) : 'Local provider disable failed')
    }
  }, [])

  useEffect(() => {
    if (!apiClient.listToolCatalog) {
      setToolCatalog([])
      return
    }
    let cancelled = false
    apiClient.listToolCatalog()
      .then((tools) => {
        if (!cancelled) {
          setToolCatalog(tools)
          setToolCatalogLoaded(true)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setToolCatalog([])
          setToolCatalogLoaded(true)
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!apiClient.getWebSearchConfig) {
      setWebSearchConfig(null)
      return
    }
    let cancelled = false
    apiClient.getWebSearchConfig()
      .then((config) => {
        if (!cancelled) setWebSearchConfig(config)
      })
      .catch(() => {
        if (!cancelled) setWebSearchConfig(null)
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (!apiClient.getWorkspaceRoot) {
      setWorkspaceRootConfig(null)
      return
    }
    let cancelled = false
    apiClient.getWorkspaceRoot()
      .then((config) => {
        if (!cancelled) {
          setWorkspaceRootConfig(config)
          setApiConnected(true)
        }
      })
      .catch(() => {
        if (!cancelled) setWorkspaceRootConfig(null)
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

  useEffect(() => {
    if (!apiClient.listSkills) {
      setInstalledSkills([])
      setSkillsError(null)
      return
    }
    let cancelled = false
    setSkillsLoading(true)
    setSkillsError(null)
    apiClient.listSkills()
      .then((items) => {
        if (!cancelled) setInstalledSkills(items)
      })
      .catch((err) => {
        if (!cancelled) {
          setInstalledSkills([])
          setSkillsError(err instanceof Error ? err.message : 'Skills failed to load')
        }
      })
      .finally(() => {
        if (!cancelled) setSkillsLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  const loadMemoryEntries = useCallback(async (query = '', filters = memoryFilters) => {
    const requestID = memoryEntriesRequestRef.current + 1
    memoryEntriesRequestRef.current = requestID
    if (!apiClient.listMemoryEntries || !apiClient.searchMemory) {
      setMemoryEntries([])
      return
    }
    setMemoryLoading(true)
    setMemoryError(null)
    try {
      const entries = query.trim()
        ? await apiClient.searchMemory(query, filters)
        : await apiClient.listMemoryEntries(filters)
      if (!shouldApplyLatestRequest(requestID, memoryEntriesRequestRef.current)) return
      setMemoryEntries(entries)
    } catch (err) {
      if (!shouldApplyLatestRequest(requestID, memoryEntriesRequestRef.current)) return
      setMemoryEntries([])
      setMemoryError(err instanceof Error ? err.message : 'Memory failed to load')
    } finally {
      if (shouldApplyLatestRequest(requestID, memoryEntriesRequestRef.current)) setMemoryLoading(false)
    }
  }, [memoryFilters])

  const setMemorySearchQuery = useCallback((query: string) => {
    setMemoryQuery(query)
    void loadMemoryEntries(query, memoryFilters)
  }, [loadMemoryEntries, memoryFilters])

  const loadMemoryAudit = useCallback(async (filters = memoryFilters) => {
    const requestID = memoryAuditRequestRef.current + 1
    memoryAuditRequestRef.current = requestID
    if (!apiClient.listMemoryAudit) {
      setMemoryAuditItems([])
      setMemoryAuditError('Memory history endpoint unavailable')
      return
    }
    setMemoryAuditLoading(true)
    setMemoryAuditError(null)
    try {
      const auditItems = await apiClient.listMemoryAudit(filters)
      if (!shouldApplyLatestRequest(requestID, memoryAuditRequestRef.current)) return
      setMemoryAuditItems(auditItems)
    } catch (err) {
      if (!shouldApplyLatestRequest(requestID, memoryAuditRequestRef.current)) return
      setMemoryAuditItems([])
      setMemoryAuditError(err instanceof Error ? err.message : 'Memory history failed to load')
    } finally {
      if (shouldApplyLatestRequest(requestID, memoryAuditRequestRef.current)) setMemoryAuditLoading(false)
    }
  }, [memoryFilters])

  const loadMemoryWriteProposals = useCallback(async (filters = memoryFilters) => {
    const requestID = memoryProposalsRequestRef.current + 1
    memoryProposalsRequestRef.current = requestID
    if (!apiClient.listMemoryWriteProposals) {
      setMemoryWriteProposals([])
      return
    }
    setMemoryProposalsLoading(true)
    setMemoryProposalsError(null)
    try {
      const proposals = await apiClient.listMemoryWriteProposals(filters)
      if (!shouldApplyLatestRequest(requestID, memoryProposalsRequestRef.current)) return
      setMemoryWriteProposals(proposals)
    } catch (err) {
      if (!shouldApplyLatestRequest(requestID, memoryProposalsRequestRef.current)) return
      setMemoryWriteProposals([])
      setMemoryProposalsError(err instanceof Error ? err.message : 'Memory proposals failed to load')
    } finally {
      if (shouldApplyLatestRequest(requestID, memoryProposalsRequestRef.current)) setMemoryProposalsLoading(false)
    }
  }, [memoryFilters])

  const updateMemoryFilters = useCallback((filters: MemoryFilters) => {
    setMemoryFilters(filters)
    void loadMemoryEntries(memoryQuery, filters)
    void loadMemoryAudit(filters)
    void loadMemoryWriteProposals(filters)
  }, [loadMemoryAudit, loadMemoryEntries, loadMemoryWriteProposals, memoryQuery])

  const openMemoryDetail = useCallback(async (entry: MemoryEntry) => {
    const requestID = memoryDetailRequestRef.current + 1
    memoryDetailRequestRef.current = requestID
    if (!apiClient.getMemoryEntry) {
      setMemoryDetail(entry)
      return
    }
    setMemoryDetail(entry)
    setMemoryDetailLoading(true)
    setMemoryDetailError(null)
    try {
      const detail = await apiClient.getMemoryEntry(entry.id, memoryContextForEntry(entry))
      if (!shouldApplyLatestRequest(requestID, memoryDetailRequestRef.current)) return
      setMemoryDetail(detail)
    } catch (err) {
      if (!shouldApplyLatestRequest(requestID, memoryDetailRequestRef.current)) return
      setMemoryDetail(null)
      setMemoryDetailError(err instanceof Error ? err.message : 'Memory detail could not be loaded')
    } finally {
      if (shouldApplyLatestRequest(requestID, memoryDetailRequestRef.current)) setMemoryDetailLoading(false)
    }
  }, [])

  const requestDeleteMemoryEntry = useCallback((entry: MemoryEntry) => {
    setPendingDeleteMemoryEntry(entry)
  }, [])

  const cancelDeleteMemoryEntry = useCallback(() => {
    setPendingDeleteMemoryEntry(null)
  }, [])

  const deleteMemoryEntry = useCallback(async (entry: MemoryEntry) => {
    if (!apiClient.deleteMemoryEntry) return
    setMemoryError(null)
    try {
      await apiClient.deleteMemoryEntry(entry.id, memoryContextForEntry(entry))
      setPendingDeleteMemoryEntry(null)
      setMemoryDetail((current) => (current?.id === entry.id ? null : current))
      await loadMemoryEntries(memoryQuery, memoryFilters)
      await loadMemoryAudit(memoryFilters)
    } catch (err) {
      setPendingDeleteMemoryEntry(null)
      setMemoryError(err instanceof Error ? err.message : 'Memory delete failed')
    }
  }, [loadMemoryAudit, loadMemoryEntries, memoryFilters, memoryQuery])

  useEffect(() => {
    if (!apiConnected) return
    void loadMemoryEntries('', memoryFilters)
  }, [apiConnected, loadMemoryEntries, memoryFilters])

  useEffect(() => {
    if (!apiConnected) return
    void loadMemoryAudit(memoryFilters)
  }, [apiConnected, loadMemoryAudit, memoryFilters])

  useEffect(() => {
    if (!apiConnected) return
    void loadMemoryWriteProposals(memoryFilters)
  }, [apiConnected, loadMemoryWriteProposals, memoryFilters])

  const approveMemoryWriteProposal = useCallback(async (proposal: MemoryWriteProposal) => {
    if (!apiClient.approveMemoryWriteProposal) return
    setMemoryProposalsError(null)
    try {
      await apiClient.approveMemoryWriteProposal(proposal.id)
      await loadMemoryWriteProposals(memoryFilters)
      await loadMemoryEntries(memoryQuery, memoryFilters)
      await loadMemoryAudit(memoryFilters)
    } catch (err) {
      setMemoryProposalsError(err instanceof Error ? err.message : 'Memory proposal approval failed')
    }
  }, [loadMemoryAudit, loadMemoryEntries, loadMemoryWriteProposals, memoryFilters, memoryQuery])

  const updateMemoryWriteProposal = useCallback(async (proposal: MemoryWriteProposal, input: { title: string; summary: string }) => {
    if (!apiClient.updateMemoryWriteProposal) return
    setMemoryProposalsError(null)
    try {
      const updated = await apiClient.updateMemoryWriteProposal(proposal.id, input)
      setMemoryWriteProposals((current) => current.map((item) => (item.id === updated.id ? updated : item)))
    } catch (err) {
      setMemoryProposalsError(err instanceof Error ? err.message : 'Memory proposal update failed')
    }
  }, [])

  const denyMemoryWriteProposal = useCallback(async (proposal: MemoryWriteProposal) => {
    if (!apiClient.denyMemoryWriteProposal) return
    setMemoryProposalsError(null)
    try {
      await apiClient.denyMemoryWriteProposal(proposal.id)
      await loadMemoryWriteProposals(memoryFilters)
      await loadMemoryAudit(memoryFilters)
    } catch (err) {
      setMemoryProposalsError(err instanceof Error ? err.message : 'Memory proposal denial failed')
    }
  }, [loadMemoryAudit, loadMemoryWriteProposals, memoryFilters])

  const refreshMemoryProviderStatus = useCallback(async () => {
    if (!apiClient.getMemoryProviderStatus) {
      setMemoryProviderStatus(null)
      return
    }
    try {
      setMemoryProviderStatus(await apiClient.getMemoryProviderStatus())
    } catch {
      setMemoryProviderStatus(null)
    }
  }, [])

  const refreshMemoryErrors = useCallback(async () => {
    if (!apiClient.listMemoryErrors) {
      setMemoryErrors([])
      return
    }
    try {
      setMemoryErrors(await apiClient.listMemoryErrors())
    } catch {
      setMemoryErrors([])
    }
  }, [])

  const refreshMemorySnapshots = useCallback(async () => {
    setMemorySnapshotLoading(true)
    try {
      const [overview, impression] = await Promise.all([
        apiClient.getMemoryOverviewSnapshot?.(),
        apiClient.getMemoryImpressionSnapshot?.(),
      ])
      if (overview) setMemoryOverviewSnapshot(overview)
      if (impression) setMemoryImpressionSnapshot(impression)
    } catch {
      setMemoryOverviewSnapshot(null)
      setMemoryImpressionSnapshot(null)
    } finally {
      setMemorySnapshotLoading(false)
    }
  }, [])

  const rebuildMemoryOverviewSnapshot = useCallback(async () => {
    if (!apiClient.rebuildMemoryOverviewSnapshot) return
    setMemorySnapshotLoading(true)
    try {
      setMemoryOverviewSnapshot(await apiClient.rebuildMemoryOverviewSnapshot())
    } finally {
      setMemorySnapshotLoading(false)
    }
  }, [])

  const rebuildMemoryImpressionSnapshot = useCallback(async () => {
    if (!apiClient.rebuildMemoryImpressionSnapshot) return
    setMemorySnapshotLoading(true)
    try {
      setMemoryImpressionSnapshot(await apiClient.rebuildMemoryImpressionSnapshot())
    } finally {
      setMemorySnapshotLoading(false)
    }
  }, [])

  const getMemoryContent = useCallback(async (uri: string, layer: 'overview' | 'read' = 'overview') => {
    return apiClient.getMemoryContent?.(uri, layer) ?? ''
  }, [])

  const detectNowledgeMemoryProvider = useCallback(async () => {
    return apiClient.detectNowledgeMemoryProvider?.() ?? { detected: false, message: 'Nowledge detect endpoint unavailable' }
  }, [])

  const detectOpenVikingMemoryProvider = useCallback(async () => {
    return apiClient.detectOpenVikingMemoryProvider?.() ?? { detected: false, message: 'OpenViking detect endpoint unavailable' }
  }, [])

  const createMemoryEntry = useCallback(async (input: { title: string; content: string; scopeType?: 'user' | 'thread'; scopeId?: string }) => {
    if (!apiClient.createMemoryEntry) return
    setMemoryError(null)
    try {
      await apiClient.createMemoryEntry(input)
      await loadMemoryEntries(memoryQuery, memoryFilters)
      await loadMemoryAudit(memoryFilters)
      await refreshMemorySnapshots()
    } catch (err) {
      setMemoryError(err instanceof Error ? err.message : 'Memory create failed')
    }
  }, [loadMemoryAudit, loadMemoryEntries, memoryFilters, memoryQuery, refreshMemorySnapshots])

  useEffect(() => {
    if (!apiConnected) return
    void refreshMemoryProviderStatus()
    void refreshMemoryErrors()
    void refreshMemorySnapshots()
  }, [apiConnected, refreshMemoryErrors, refreshMemoryProviderStatus, refreshMemorySnapshots])

  useEffect(() => {
    if (!run || !shouldBlockRuntimeSubmit(run) || !apiClient.subscribeRunEvents) {
      setStreamState((current) => {
        const next = run && shouldBlockRuntimeSubmit(run) ? 'recoverable_error' : 'closed'
        return current === next ? current : next
      })
      return
    }
    let cancelled = false
    const activeRunID = run.id
    const reconcileActiveRun = async () => {
      const threadId = selectedThreadIdRef.current
      if (!threadId) return
      try {
        const [nextMessages, nextRun] = await Promise.all([apiClient.getThreadMessages(threadId), apiClient.getThreadRun(threadId)])
        if (cancelled || !nextRun || nextRun.id !== activeRunID) return
        const reconciledRun = reconcileRunWithPersistedAssistant(nextRun, nextMessages)
        setMessages(nextMessages)
        setRun(reconciledRun)
        runRef.current = reconciledRun
        if (!shouldBlockRuntimeSubmit(reconciledRun)) {
          setCapabilitySignals((current) => ({ ...current, streamDisconnected: false }))
          setStreamState((current) => (current === 'closed' ? current : 'closed'))
        }
      } catch {
        if (!cancelled) setCapabilitySignals((current) => ({ ...current, streamDisconnected: true }))
      }
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
        void reconcileActiveRun()
      },
      () => {
        setStreamState((current) => (current === 'closed' ? current : 'closed'))
        void reconcileActiveRun()
      },
    )
    const reconcileInterval = window.setInterval(() => {
      void reconcileActiveRun()
    }, 2500)
    return () => {
      cancelled = true
      window.clearInterval(reconcileInterval)
      unsubscribe()
    }
  }, [run?.id, run?.status])

  const selectThread = useCallback((threadId: string) => {
    setSelectedThreadId(threadId)
  }, [])

  const sendMessage = useCallback(async (content: string, options?: { providerId?: string; model?: string }) => {
    const trimmed = content.trim()
    if (!trimmed) return
    const requestedThreadId = selectedThreadId
    setError(null)
    setBackendUnavailableAttempted(false)
    setCapabilitySignals({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
    try {
      const result = await apiClient.sendMessage(requestedThreadId, trimmed, selectedPersonaId || undefined, options)
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

  const createThread = useCallback(async (mode: Thread['mode'] = defaultWorkspaceMode) => {
    if (!apiClient.createThread) return
    setError(null)
    try {
      const thread = await apiClient.createThread(createNextThreadTitle(threads), mode)
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
      const nextThreadId = getThreadIdAfterArchive({ archivedThreadId: threadId, currentSelectedThreadId: selectedThreadIdRef.current, threads })
      setThreads((current) => current.filter((thread) => thread.id !== threadId))
      if (nextThreadId !== selectedThreadIdRef.current) {
        setSelectedThreadId(nextThreadId)
        if (!nextThreadId) {
          setMessages([])
          setRun(null)
        }
      }
      await refresh(nextThreadId)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
    }
  }, [refresh, threads])

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
        const providers = await apiClient.listModelProviders?.()
        const provider = selectSendProvider(providers)
        const messageId = runMessageId(run, messages)
        if (apiClient.mode === 'real_api' && (!provider || !messageId)) throw new Error('Model provider is unavailable.')
        const nextRun = provider && messageId
          ? await apiClient.startRun(run.threadId, { messageId, source: 'model_gateway', providerId: provider.id, model: provider.model, personaId: selectedPersonaId || undefined })
          : await apiClient.startRun(run.threadId)
        setRun(nextRun)
      } else {
        setRun(createRetryAttemptRun(run))
      }
      setCapabilitySignals({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
      setStreamState('connecting')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
    }
  }, [messages, run, selectedPersonaId])

  const regenerateRun = useCallback(async () => {
    const lastAssistant = [...messages].reverse().find((message) => message.role === 'assistant')
    if (!run || !lastAssistant || shouldBlockRuntimeSubmit(run)) return
    setError(null)
    try {
      if (apiClient.startRun) {
        const providers = await apiClient.listModelProviders?.()
        const provider = selectSendProvider(providers)
        const messageId = runMessageId(run, messages)
        if (apiClient.mode === 'real_api' && (!provider || !messageId)) throw new Error('Model provider is unavailable.')
        const nextRun = provider && messageId
          ? await apiClient.startRun(run.threadId, { messageId, source: 'model_gateway', providerId: provider.id, model: provider.model, personaId: selectedPersonaId || undefined })
          : await apiClient.startRun(run.threadId)
        setRun({ ...nextRun, attemptOfMessageId: lastAssistant.id })
      } else {
        setRun(createRegenerateAttemptRun(run, lastAssistant.id))
      }
      setCapabilitySignals({ backendUnavailable: false, modelSetupMissing: false, providerUnavailable: false, streamDisconnected: false })
      setStreamState('connecting')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'API request failed')
      setCapabilitySignals((current) => ({ ...current, ...mapRealRuntimeCapabilitySignal(err) }))
    }
  }, [messages, run, selectedPersonaId])

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
          status: ['available', 'configured', 'reachable', 'completion-ok'].includes(provider.status) ? 'success' : 'failed',
          message: provider.checkCode ?? provider.message ?? provider.status,
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
      setProviderCheckResults((current) => ({ ...current, [provider.id]: { status: ['available', 'configured', 'reachable', 'completion-ok'].includes(provider.status) ? 'success' : 'failed', message: provider.checkCode ?? provider.message ?? provider.status } }))
      setProviderSaveResult({ status: ['available', 'configured', 'reachable', 'completion-ok'].includes(provider.status) ? 'success' : 'failed', message: provider.checkCode ?? provider.message ?? provider.status })
    } catch (err) {
      setProviderSaveResult({ status: 'failed', message: redactProviderCheckMessage(err instanceof Error ? err.message : 'Provider save failed') })
    }
  }, [])

  const saveWebSearchKeys = useCallback(async (input: { tavilyApiKey?: string; braveApiKey?: string }) => {
    if (!apiClient.saveWebSearchKeys) return
    setWebSearchSaveResult({ status: 'saving' })
    try {
      const config = await apiClient.saveWebSearchKeys(input)
      setWebSearchConfig(config)
      setWebSearchSaveResult({ status: config.enabled ? 'success' : 'failed', message: config.enabled ? 'Saved' : 'At least one key is required' })
    } catch (err) {
      setWebSearchSaveResult({ status: 'failed', message: redactProviderCheckMessage(err instanceof Error ? err.message : 'Web search save failed') })
    }
  }, [])

  const updateMemoryProvider = useCallback(async (input: MemoryProviderUpdate) => {
    if (!apiClient.updateMemoryProvider) return
    setMemoryProviderSaveResult({ status: 'saving' })
    try {
      const status = await apiClient.updateMemoryProvider(input)
      setMemoryProviderStatus(status)
      void refreshMemoryErrors()
      setMemoryProviderSaveResult({ status: status.state === 'healthy' || status.state === 'available' ? 'success' : 'failed', message: status.diagnostic.message })
    } catch (err) {
      setMemoryProviderSaveResult({ status: 'failed', message: redactProviderCheckMessage(err instanceof Error ? err.message : 'Memory provider update failed') })
    }
  }, [refreshMemoryErrors])

  const chooseWorkspaceFolder = useCallback(async () => {
    if (!apiClient.saveWorkspaceRoot) {
      setWorkspaceRootSaveResult({ status: 'failed', message: 'Workspace folder endpoint unavailable' })
      return
    }
    const picker = window.loomiDesktop?.selectWorkspaceFolder
    if (!picker) {
      setWorkspaceRootSaveResult({ status: 'failed', message: '请在桌面端选择目录' })
      return
    }
    setWorkspaceRootSaveResult({ status: 'saving' })
    try {
      const selected = await picker()
      if (selected.canceled || !selected.path) {
        setWorkspaceRootSaveResult({ status: 'idle' })
        return
      }
      const config = await apiClient.saveWorkspaceRoot({ path: selected.path })
      setWorkspaceRootConfig(config)
      setWorkspaceRootSaveResult({ status: 'success', message: `已授权 ${config.displayName}` })
      setError(null)
    } catch (err) {
      setWorkspaceRootSaveResult({ status: 'failed', message: err instanceof Error ? err.message : 'Workspace folder save failed' })
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
    desktopReadiness,
    refreshDesktopReadiness,
    selectedRuntimeScript,
    providerCapabilities,
    toolCatalog,
    webSearchConfig,
    webSearchSaveResult,
    workspaceRootConfig,
    workspaceRootSaveResult,
    mcpServers,
    mcpActionResult,
    localProviderDetections,
    localProviderDetectionError,
    personas,
    installedSkills,
    skillsLoading,
    skillsError,
    selectedPersonaId,
    providerCheckResults,
    providerSaveResult,
    memoryEntries,
    memoryQuery,
    memoryFilters,
    memoryLoading,
    memoryError,
    memoryDetail,
    memoryDetailLoading,
    memoryDetailError,
    memoryAuditItems,
    memoryAuditLoading,
    memoryAuditError,
    memoryWriteProposals,
    memoryProposalsLoading,
    memoryProposalsError,
    memoryProviderStatus,
    memoryErrors,
    memoryProviderSaveResult,
    memoryOverviewSnapshot,
    memoryImpressionSnapshot,
    memorySnapshotLoading,
    pendingDeleteMemoryEntry,
    selectRuntimeScript,
    setSelectedPersonaId,
    checkProvider,
    detectLocalProviders,
    enableLocalProvider,
    disableLocalProvider,
    saveProvider,
    saveWebSearchKeys,
    refreshMemoryProviderStatus,
    refreshMemoryErrors,
    updateMemoryProvider,
    detectNowledgeMemoryProvider,
    detectOpenVikingMemoryProvider,
    refreshMemorySnapshots,
    rebuildMemoryOverviewSnapshot,
    rebuildMemoryImpressionSnapshot,
    getMemoryContent,
    chooseWorkspaceFolder,
    saveMCPServer,
    deleteMCPServer,
    discoverMCPServer,
    setMemorySearchQuery,
    updateMemoryFilters,
    openMemoryDetail,
    closeMemoryDetail: () => setMemoryDetail(null),
    requestDeleteMemoryEntry,
    cancelDeleteMemoryEntry,
    createMemoryEntry,
    deleteMemoryEntry,
    approveMemoryWriteProposal,
    updateMemoryWriteProposal,
    denyMemoryWriteProposal,
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
