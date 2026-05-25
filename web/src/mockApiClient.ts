import type { ApiClient } from './apiClient'
import type { LocalProviderDetection, Message, ProviderCapability, Run, RuntimeScriptId, ToolCatalogItem } from './domain'
import { messages, runs, threads } from './mockData'
import { isRuntimeTerminal } from './runtime/executionAdapter'
import { mockExecutionAdapter } from './runtime/mockExecutionAdapter'
import { createRuntimeEvent, getRuntimeScript, getRuntimeScriptSteps } from './runtime/runtimeScripts'

let mockId = 1000
let threadStore = [...threads]
let messageStore = [...messages]
let runStore = runs.map((run) => ({ ...run, events: [...run.events] }))
let selectedRuntimeScriptId: RuntimeScriptId = 'success'
let mockLocalProviderDetections: LocalProviderDetection[] = []
let mockProviderCapabilities: ProviderCapability[] = []

const mockToolCatalog: ToolCatalogItem[] = [{
  name: 'runtime.get_current_time',
  displayName: 'Current time',
  description: 'Returns the current UTC time.',
  source: 'builtin',
  group: 'runtime',
  inputSchemaHash: 'sha256:mock-current-time',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
}]

function nextMockId(prefix: string) {
  mockId += 1
  return `${prefix}-${mockId}`
}

function cloneRun(run: Run): Run {
  return { ...run, events: [...run.events], assistantDraft: run.assistantDraft ? { ...run.assistantDraft } : undefined }
}

function createIdleRun(threadId: string): Run {
  return {
    id: `run-${threadId}`,
    threadId,
    status: 'completed',
    model: 'Claude Sonnet',
    context: 'Ready',
    events: [],
  }
}

function updateRunStore(run: Run) {
  const exists = runStore.some((item) => item.id === run.id || item.threadId === run.threadId)
  runStore = exists ? runStore.map((item) => (item.id === run.id || item.threadId === run.threadId ? cloneRun(run) : item)) : [cloneRun(run), ...runStore]
}

function applyMockRunEvent(run: Run, event: Run['events'][number]): Run {
  if (isRuntimeTerminal(run.status)) return run
  if (run.events.some((existing) => existing.id === event.id)) return run
  const lastSequence = run.events.at(-1)?.sequence ?? -1
  const isOutOfOrder = event.sequence !== undefined && lastSequence > event.sequence
  const events = [...run.events, event].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0))
  const content = event.assistantDelta && !isOutOfOrder ? `${run.assistantDraft?.content ?? ''}${event.assistantDelta}` : run.assistantDraft?.content ?? ''

  if (event.status === 'completed') {
    return { ...run, status: 'completed', events, completedAt: event.time, assistantDraft: { content: event.content ?? content, status: 'completed', messageId: run.assistantDraft?.messageId, lastEventId: event.id } }
  }
  if (event.status === 'failed' || event.status === 'stopped') {
    return { ...run, status: event.status, events, completedAt: event.time, assistantDraft: { content, status: event.status, lastEventId: event.id } }
  }
  if (event.status === 'recovering') {
    return { ...run, status: 'recovering', events, assistantDraft: { content, status: 'recovering', lastEventId: event.id } }
  }
  return { ...run, status: event.status, events, assistantDraft: event.assistantDelta && !isOutOfOrder ? { content, status: 'streaming', lastEventId: event.id } : run.assistantDraft }
}

function completeMockRun(run: Run): Run {
  const script = getRuntimeScript(run.scriptId ?? 'success')
  if (script.terminalStatus !== 'completed') return run
  const assistantMessage: Message = {
    id: nextMockId('msg'),
    threadId: run.threadId,
    role: 'assistant',
    content: run.assistantDraft?.content || script.finalAssistantMessage || '已完成一次模拟执行。',
    createdAt: 'Now',
    runId: run.id,
  }
  messageStore = [...messageStore, assistantMessage]
  return { ...run, assistantDraft: { content: assistantMessage.content, status: 'completed', messageId: assistantMessage.id } }
}

function playMockRunScript(runId: string, stepIndex = 0) {
  const run = runStore.find((item) => item.id === runId)
  if (!run) return
  const steps = getRuntimeScriptSteps(run.scriptId ?? 'success')
  const step = steps[stepIndex]
  if (!step) {
    const terminalRun = runStore.find((item) => item.id === runId)
    if (terminalRun?.status === 'completed') updateRunStore(completeMockRun(terminalRun))
    const finalRun = runStore.find((item) => item.id === runId)
    if (finalRun) updateThreadRunStatus(finalRun.threadId, finalRun.status)
    return
  }

  const current = runStore.find((item) => item.id === runId)
  if (!current || isRuntimeTerminal(current.status)) return
  const event = createRuntimeEvent({ threadId: current.threadId, runId, sequence: stepIndex, step })
  updateRunStore(applyMockRunEvent(current, event))
  notifyRunSubscribers(runId, event)
  setTimeout(() => playMockRunScript(runId, stepIndex + 1), 16)
}

function scheduleMockRunScript(runId: string) {
  if (scheduledRunScripts.has(runId)) return
  scheduledRunScripts.add(runId)
  setTimeout(() => playMockRunScript(runId), 16)
}

export function setMockRuntimeScript(scriptId: RuntimeScriptId) {
  selectedRuntimeScriptId = scriptId
}

const runSubscribers = new Map<string, Set<(event: Run['events'][number]) => void>>()
const scheduledRunScripts = new Set<string>()

function notifyRunSubscribers(runId: string, event: Run['events'][number]) {
  runSubscribers.get(runId)?.forEach((subscriber) => subscriber(event))
}

function updateThreadRunStatus(threadId: string, status: Run['status']) {
  threadStore = threadStore.map((thread) => (thread.id === threadId ? { ...thread, runStatus: status, updatedAt: 'Now' } : thread))
}

export const mockApiClient: ApiClient = {
  mode: 'mock',

  async listThreads() {
    return threadStore.filter((thread) => thread.lifecycleStatus !== 'archived')
  },

  async getThreadMessages(threadId: string) {
    return messageStore.filter((message) => message.threadId === threadId)
  },

  async getThreadRun(threadId: string) {
    const run = runStore.find((item) => item.threadId === threadId)
    if (!run) throw new Error('Run not found')
    return cloneRun(run)
  },

  async getRunEvents(runId: string) {
    return runStore.find((run) => run.id === runId)?.events ?? []
  },

  async listModelProviders() {
    return mockProviderCapabilities
  },

  async listLocalProviderDetections() {
    mockLocalProviderDetections = [
      {
        providerId: 'local_claude_code',
        displayName: 'Local Claude Code',
        providerKind: 'claude_code',
        authMode: 'unknown',
        status: 'unavailable',
        modelCandidates: ['claude-sonnet-4-5'],
        source: 'unknown',
        redactionApplied: true,
        message: 'Not detected.',
      },
      {
        providerId: 'local_codex',
        displayName: 'Local Codex',
        providerKind: 'codex',
        authMode: 'unknown',
        status: 'available',
        modelCandidates: ['gpt-5.5'],
        source: 'unknown',
        redactionApplied: true,
        message: 'Detected but not enabled. Explicit opt-in is required before use.',
      },
    ]
    return mockLocalProviderDetections
  },

  async enableLocalProvider(providerId: string) {
    const detection = mockLocalProviderDetections.find((provider) => provider.providerId === providerId)
    if (!detection || detection.status !== 'available') throw new Error('Local provider is not available.')
    const capability: ProviderCapability = {
      id: detection.providerId,
      family: 'openai_compatible',
      model: detection.modelCandidates[0] ?? 'gpt-5.5',
      status: 'unavailable',
      message: `${detection.displayName} is enabled for this session, but execution is unsupported until the local provider execution bridge is implemented.`,
      localProvider: true,
      sessionLocal: true,
      credentialReference: 'redacted',
      executionState: 'unsupported',
    }
    mockProviderCapabilities = [...mockProviderCapabilities.filter((provider) => provider.id !== providerId), capability]
    return capability
  },

  async disableLocalProvider(providerId: string) {
    const capability = mockProviderCapabilities.find((provider) => provider.id === providerId)
    mockProviderCapabilities = mockProviderCapabilities.filter((provider) => provider.id !== providerId)
    if (!capability) throw new Error('Local provider is not enabled.')
    return capability
  },

  async listToolCatalog() {
    return mockToolCatalog
  },

  subscribeRunEvents(runId: string, afterSequence: number, onEvent) {
    const replay = runStore.find((run) => run.id === runId)?.events.filter((event) => (event.sequence ?? 0) > afterSequence) ?? []
    replay.forEach(onEvent)
    const subscribers = runSubscribers.get(runId) ?? new Set()
    subscribers.add(onEvent)
    runSubscribers.set(runId, subscribers)
    scheduleMockRunScript(runId)
    return () => subscribers.delete(onEvent)
  },

  async createThread(title: string, mode) {
    const thread = {
      id: `thread-${nextMockId('mock')}`,
      title,
      project: 'Loomi',
      mode,
      updatedAt: 'Now',
      lifecycleStatus: 'active' as const,
      runStatus: 'completed' as const,
    }
    threadStore = [thread, ...threadStore]
    runStore = [createIdleRun(thread.id), ...runStore]
    return thread
  },

  async updateThread(threadId: string, input) {
    threadStore = threadStore.map((thread) => (thread.id === threadId ? { ...thread, ...input, updatedAt: 'Now' } : thread))
    const thread = threadStore.find((item) => item.id === threadId)
    if (!thread) throw new Error('Thread not found')
    return thread
  },

  async archiveThread(threadId: string) {
    threadStore = threadStore.map((thread) => (thread.id === threadId ? { ...thread, lifecycleStatus: 'archived' as const } : thread))
    const thread = threadStore.find((item) => item.id === threadId)
    if (!thread) throw new Error('Thread not found')
    return thread
  },

  async startRun(threadId: string) {
    const runningRun: Run = {
      id: nextMockId('run'),
      threadId,
      status: 'running',
      model: 'Mock Runtime',
      context: 'M3.5 mock',
      events: [],
      scriptId: selectedRuntimeScriptId,
      assistantDraft: { content: '', status: 'pending' },
      createdAt: 'Now',
    }
    updateRunStore(runningRun)
    updateThreadRunStatus(threadId, 'running')
    return cloneRun(runningRun)
  },

  async sendMessage(threadId: string, content: string) {
    const userMessage = await mockExecutionAdapter.sendMessage(threadId, content)
    messageStore = [...messageStore, userMessage]

    const runningRun = await this.startRun!(threadId)
    return { messages: await this.getThreadMessages(threadId), run: runningRun }
  },

  async stopRun(runId: string) {
    const run = runStore.find((item) => item.id === runId)
    if (!run) throw new Error('Run not found')
    let stopped: Run
    try {
      stopped = await mockExecutionAdapter.stopRun(run.threadId, runId)
    } catch (err) {
      if (!(err instanceof Error) || err.message !== 'Run not found') throw err
      stopped = {
        ...run,
        status: 'stopped',
        assistantDraft: { content: run.assistantDraft?.content ?? '', status: 'stopped' },
        events: [...run.events, { id: `${runId}-stopped`, type: 'run.stopped', label: 'Stopped', detail: '已停止', time: 'Now', status: 'stopped' }],
      }
    }
    updateRunStore(stopped)
    threadStore = threadStore.map((thread) => (thread.id === stopped.threadId ? { ...thread, runStatus: 'stopped' } : thread))
    return cloneRun(stopped)
  },
}
