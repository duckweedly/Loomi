import type { ApiClient } from './apiClient'
import type { InstalledSkill, LocalProviderDetection, MCPServerConfigInput, MCPServerStatus, MemoryEntry, MemoryErrorEvent, MemoryImpressionSnapshot, MemoryOverviewSnapshot, MemoryProviderStatus, MemoryProviderUpdate, MemoryWriteProposal, Message, Persona, ProviderCapability, Run, RuntimeScriptId, ToolCatalogItem } from './domain'
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
let mockMemoryProposals: MemoryWriteProposal[] = []
let mockMemoryEntries: MemoryEntry[] = []
let mockMemoryErrors: MemoryErrorEvent[] = []
let mockMemoryProviderStatus: MemoryProviderStatus = {
  enabled: true,
  provider: 'local',
  label: 'Local',
  state: 'available',
  configured: true,
  commitAfterRun: false,
  diagnostic: { code: 'ok', message: 'Ready.' },
}
let mockMemoryOverviewSnapshot: MemoryOverviewSnapshot = {
  memoryBlock: 'No approved memories yet.',
  hits: [],
  updatedAt: new Date().toISOString(),
  rebuilt: false,
}
let mockMemoryImpressionSnapshot: MemoryImpressionSnapshot = {
  impression: 'No approved memories have been saved yet.',
  updatedAt: new Date().toISOString(),
  rebuilt: false,
}

const mockPersonas: Persona[] = [{
  id: 'persona-default',
  slug: 'default',
  name: 'Loomi Default',
  description: 'Default local persona with bounded tool access.',
  activeVersion: '2026-05-26.1',
  isDefault: true,
}]

const mockInstalledSkills: InstalledSkill[] = [{
  id: 'project:speckit-implement',
  name: 'speckit-implement',
  description: 'Execute implementation tasks from the project Spec Kit plan.',
  source: 'project',
  sourceLabel: 'Project .agents',
  package: 'speckit',
  path: '.agents/skills/speckit-implement/SKILL.md',
  installed: true,
}, {
  id: 'codex:skill-creator',
  name: 'skill-creator',
  description: 'Guide for creating effective Codex skills.',
  source: 'codex',
  sourceLabel: 'Codex',
  path: '~/.codex/skills/.system/skill-creator/SKILL.md',
  installed: true,
}]

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
}, {
  name: 'workspace.glob',
  displayName: 'Workspace glob',
  description: 'Find files under the configured workspace root.',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'workspace', arguments: ['pattern', 'path', 'limit'] },
}, {
  name: 'workspace.grep',
  displayName: 'Workspace grep',
  description: 'Search text files under the configured workspace root.',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'workspace', arguments: ['query', 'path', 'include', 'case_sensitive', 'limit'] },
}, {
  name: 'workspace.read',
  displayName: 'Workspace read',
  description: 'Read a bounded UTF-8 text slice from one workspace file.',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'workspace', arguments: ['path', 'offset', 'limit', 'max_bytes'] },
}, {
  name: 'workspace.write_file',
  displayName: 'Workspace write file',
  description: 'Create a bounded UTF-8 text file under the configured workspace root.',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'high',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: false, write_capable: true, scope: 'workspace', arguments: ['path', 'content', 'max_bytes'] },
}, {
  name: 'workspace.edit',
  displayName: 'Workspace edit',
  description: 'Apply one bounded exact text replacement inside a workspace file.',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'high',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: {
    read_only: false,
    write_capable: true,
    requires_read_before_edit: true,
    returns_diff: true,
    normalizes_line_endings: true,
    preserves_indentation: true,
    strips_trailing_whitespace_except_markdown: true,
    scope: 'workspace',
    arguments: ['path', 'old_text', 'new_text', 'max_bytes'],
  },
}, {
  name: 'workspace.patch_preview',
  displayName: 'Workspace patch preview',
  description: 'Preview one bounded exact text replacement before applying it.',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'high',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: {
    read_only: true,
    write_capable: false,
    requires_read_before_preview: true,
    returns_diff: true,
    preview_only: true,
    normalizes_line_endings: true,
    preserves_indentation: true,
    strips_trailing_whitespace_except_markdown: true,
    scope: 'workspace',
    arguments: ['path', 'old_text', 'new_text', 'max_bytes'],
  },
}, {
  name: 'workspace.patch_apply',
  displayName: 'Workspace patch apply',
  description: 'Apply one previously previewed bounded text replacement.',
  source: 'builtin',
  group: 'workspace',
  riskLevel: 'high',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: {
    read_only: false,
    write_capable: true,
    requires_patch_preview: true,
    returns_diff: true,
    normalizes_line_endings: true,
    preserves_indentation: true,
    strips_trailing_whitespace_except_markdown: true,
    scope: 'workspace',
    arguments: ['path', 'old_text', 'new_text', 'max_bytes'],
  },
}, {
  name: 'sandbox.exec_command',
  displayName: 'Sandbox exec command',
  description: 'Run one bounded argv-form command under the configured workspace root.',
  source: 'builtin',
  group: 'sandbox',
  riskLevel: 'high',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { argv_only: true, exec_capable: true, validation_capable: true, read_only: false, isolated_sandbox: false, scope: 'bounded_command', allowed_commands: ['pwd', 'ls', 'cat', 'head', 'tail', 'sed -n', 'wc', 'rg', 'git status', 'git diff', 'git log', 'git show', 'go test', 'bun test', 'bun run build'], arguments: ['argv', 'cwd', 'timeout_ms', 'max_output_bytes'] },
}, {
  name: 'lsp.diagnostics',
  displayName: 'LSP diagnostics',
  description: 'Read bounded diagnostics for a workspace source file.',
  source: 'builtin',
  group: 'lsp',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'lsp', arguments: ['path', 'language', 'limit'] },
}, {
  name: 'lsp.symbols',
  displayName: 'LSP symbols',
  description: 'Read bounded symbol summaries for a workspace source file.',
  source: 'builtin',
  group: 'lsp',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'lsp', arguments: ['path', 'query', 'language', 'limit'] },
}, {
  name: 'lsp.references',
  displayName: 'LSP references',
  description: 'Read bounded workspace references for a source position.',
  source: 'builtin',
  group: 'lsp',
  riskLevel: 'low',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'lsp', arguments: ['path', 'line', 'column', 'include_declaration', 'limit'] },
}, {
  name: 'web.fetch',
  displayName: 'Web fetch',
  description: 'Fetch one bounded public HTTP(S) URL and return a safe text summary.',
  source: 'builtin',
  group: 'web',
  riskLevel: 'medium',
  approvalPolicy: 'read_only',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'web', network_access: 'public_http_only', arguments: ['url', 'max_bytes', 'timeout_ms'] },
}, {
  name: 'web.search',
  displayName: 'Web search',
  description: 'Search the public web through configured Brave or Tavily provider and return bounded safe results.',
  source: 'builtin',
  group: 'web',
  riskLevel: 'medium',
  approvalPolicy: 'read_only',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'web', network_access: 'search_provider_api', providers: ['tavily', 'brave'], arguments: ['query', 'provider', 'limit', 'timeout_ms'] },
}, {
  name: 'browser.open',
  displayName: 'Browser open',
  description: 'Open one bounded public HTTP(S) page in a run-scoped browser session.',
  source: 'builtin',
  group: 'browser',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'browser', network_access: 'public_http_only', stateful_session: true, arguments: ['url', 'max_bytes', 'timeout_ms'] },
}, {
  name: 'browser.snapshot',
  displayName: 'Browser snapshot',
  description: 'Return the current page title, URL, text excerpt, and safe links from a browser session.',
  source: 'builtin',
  group: 'browser',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'browser', network_access: 'public_http_only', stateful_session: true, arguments: ['session_id'] },
}, {
  name: 'browser.click_link',
  displayName: 'Browser click link',
  description: 'Navigate one safe link index inside an existing run-scoped browser session.',
  source: 'builtin',
  group: 'browser',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'browser', network_access: 'public_http_only', stateful_session: true, arguments: ['session_id', 'link_index', 'max_bytes', 'timeout_ms'] },
}, {
  name: 'artifact.create_text',
  displayName: 'Artifact create text',
  description: 'Create one bounded non-executable text artifact.',
  source: 'builtin',
  group: 'artifact',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: false, scope: 'artifact', non_executable: true, arguments: ['title', 'content', 'max_bytes'] },
}, {
  name: 'artifact.read',
  displayName: 'Artifact read',
  description: 'Read one bounded text artifact excerpt.',
  source: 'builtin',
  group: 'artifact',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'artifact', non_executable: true, arguments: ['artifact_id', 'max_bytes'] },
}, {
  name: 'artifact.list',
  displayName: 'Artifact list',
  description: 'List bounded safe artifact summaries.',
  source: 'builtin',
  group: 'artifact',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'artifact', non_executable: true, arguments: ['limit'] },
}, {
  name: 'agent.spawn',
  displayName: 'Agent spawn',
  description: 'Create one bounded coordination task record for another agent role.',
  source: 'builtin',
  group: 'agent',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: false, scope: 'agent', coordination_only: true, autonomous_execution: false, arguments: ['role', 'goal'] },
}, {
  name: 'agent.list',
  displayName: 'Agent list',
  description: 'List bounded coordination task summaries for the current thread.',
  source: 'builtin',
  group: 'agent',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'agent', coordination_only: true, autonomous_execution: false, arguments: ['limit'] },
}, {
  name: 'agent.complete',
  displayName: 'Agent complete',
  description: 'Mark one bounded coordination task complete with a safe result summary.',
  source: 'builtin',
  group: 'agent',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: false, scope: 'agent', coordination_only: true, autonomous_execution: false, arguments: ['task_id', 'result_summary'] },
}, {
  name: 'memory.search',
  displayName: 'Memory search',
  description: 'Search approved memory summaries in the current safe scope.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['query', 'limit'] },
}, {
  name: 'memory.list',
  displayName: 'Memory list',
  description: 'List approved memory summaries in the current safe scope.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['limit'] },
}, {
  name: 'memory.read',
  displayName: 'Memory read',
  description: 'Read one approved memory summary without raw content.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['entry_id'] },
}, {
  name: 'memory.write',
  displayName: 'Memory write',
  description: 'Create one approval-gated memory write proposal.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: false, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['title', 'content'] },
}, {
  name: 'memory.edit',
  displayName: 'Memory edit',
  description: 'Edit a pending memory proposal or create an approval-gated replacement proposal.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: false, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['proposal_id', 'entry_id', 'title', 'content'] },
}, {
  name: 'memory.forget',
  displayName: 'Memory forget',
  description: 'Tombstone one approved memory entry through the audited memory boundary.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: false, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['entry_id', 'reason'] },
}, {
  name: 'memory.context',
  displayName: 'Memory context',
  description: 'Return provider status plus bounded relevant memory summaries.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['query', 'limit'] },
}, {
  name: 'memory.timeline',
  displayName: 'Memory timeline',
  description: 'List safe memory audit timeline items.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['limit'] },
}, {
  name: 'memory.connections',
  displayName: 'Memory connections',
  description: 'Return bounded related memory summaries for one entry or query.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['entry_id', 'query', 'limit'] },
}, {
  name: 'memory.thread_search',
  displayName: 'Memory thread search',
  description: 'Search local thread and message history with safe excerpts.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['query', 'limit'] },
}, {
  name: 'memory.thread_fetch',
  displayName: 'Memory thread fetch',
  description: 'Fetch safe local thread message excerpts.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: ['thread_id', 'limit'] },
}, {
  name: 'memory.status',
  displayName: 'Memory status',
  description: 'Return memory provider readiness and configuration state.',
  source: 'builtin',
  group: 'memory',
  riskLevel: 'medium',
  approvalPolicy: 'always_required',
  enabled: true,
  executionState: 'executable',
  safeMetadata: { read_only: true, scope: 'memory', approval_gated: true, returns_raw_content: false, arguments: [] },
}]

let mockMCPServers: MCPServerStatus[] = [{
  serverSafeId: 'mcp:local-smoke',
  serverSlug: 'local-smoke',
  displayName: 'Local Smoke',
  transport: 'stdio',
  enabled: true,
  configSource: 'local',
  discoveryStatus: 'succeeded',
  candidateCount: 1,
  candidateNames: ['mcp.local-smoke.echo'],
  executionMode: 'approval_gated',
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

  async listPersonas() {
    return mockPersonas
  },

  async listSkills() {
    return mockInstalledSkills
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

  async listMemoryEntries() {
    return mockMemoryEntries
  },

  async searchMemory(query: string) {
    const needle = query.trim().toLowerCase()
    if (!needle) return mockMemoryEntries
    return mockMemoryEntries.filter((entry) => `${entry.title} ${entry.summary}`.toLowerCase().includes(needle))
  },

  async createMemoryEntry(input: { title: string; content: string; scopeType?: 'user' | 'thread'; scopeId?: string }) {
    const now = new Date().toISOString()
    const entry: MemoryEntry = {
      id: `mem_mock_${mockId++}`,
      title: input.title,
      summary: input.content,
      scopeType: input.scopeType ?? 'user',
      scopeId: input.scopeId ?? '',
      status: 'approved',
      safetyState: 'safe',
      sourceType: 'manual',
      createdAt: now,
      updatedAt: now,
      redactionApplied: false,
    }
    mockMemoryEntries = [entry, ...mockMemoryEntries]
    mockMemoryOverviewSnapshot = {
      memoryBlock: mockMemoryEntries.map((item) => `- ${item.title}: ${item.summary}`).join('\n') || 'No approved memories yet.',
      hits: mockMemoryEntries.map((item) => ({ uri: `memory://${item.id}`, entryId: item.id, title: item.title, abstract: item.summary, isLeaf: true, updatedAt: item.updatedAt })),
      updatedAt: now,
      rebuilt: false,
    }
    return entry
  },

  async getMemoryProviderStatus() {
    return mockMemoryProviderStatus
  },

  async updateMemoryProvider(input: MemoryProviderUpdate) {
    const openviking = input.openviking ?? mockMemoryProviderStatus.openviking ?? {}
    const nowledge = input.nowledge ?? mockMemoryProviderStatus.nowledge ?? {}
    const provider = input.provider
    const configured = provider === 'local'
      || (provider === 'openviking' && Boolean(openviking.baseUrl && (openviking.rootApiKey || openviking.rootApiKeySet) && openviking.embeddingModel && openviking.vlmModel))
      || (provider === 'nowledge' && Boolean(nowledge.baseUrl))
      || (provider === 'semantic' && Boolean(input.semanticEndpoint))
    mockMemoryProviderStatus = {
      enabled: input.enabled,
      provider,
      label: provider === 'openviking' ? 'OpenViking' : provider === 'nowledge' ? 'Nowledge' : provider === 'semantic' ? 'Semantic' : 'Local',
      state: input.enabled ? (provider === 'local' ? 'available' : configured ? 'healthy' : 'unconfigured') : 'disabled',
      configured: input.enabled && configured,
      commitAfterRun: input.commitAfterRun,
      checkedAt: new Date().toISOString(),
      openviking: {
        ...openviking,
        rootApiKey: '',
        rootApiKeySet: Boolean(openviking.rootApiKey || openviking.rootApiKeySet),
        embeddingApiKey: '',
        embeddingApiKeySet: Boolean(openviking.embeddingApiKey || openviking.embeddingApiKeySet),
        vlmApiKey: '',
        vlmApiKeySet: Boolean(openviking.vlmApiKey || openviking.vlmApiKeySet),
        rerankApiKey: '',
        rerankApiKeySet: Boolean(openviking.rerankApiKey || openviking.rerankApiKeySet),
      },
      nowledge: { ...nowledge, apiKey: '', apiKeySet: Boolean(nowledge.apiKey || nowledge.apiKeySet) },
      diagnostic: { code: configured ? 'ok' : `${provider}_unconfigured`, message: configured ? 'Ready.' : 'Provider is not configured.' },
    }
    mockMemoryErrors = configured ? [] : [{ code: mockMemoryProviderStatus.diagnostic.code, message: mockMemoryProviderStatus.diagnostic.message, provider, state: mockMemoryProviderStatus.state, checkedAt: mockMemoryProviderStatus.checkedAt ?? undefined }]
    return mockMemoryProviderStatus
  },

  async listMemoryErrors() {
    return mockMemoryErrors
  },

  async detectNowledgeMemoryProvider() {
    return { detected: true, baseUrl: 'http://127.0.0.1:14242', message: 'Nowledge local instance detected.' }
  },

  async detectOpenVikingMemoryProvider() {
    return { detected: true, baseUrl: 'http://127.0.0.1:8282', message: 'OpenViking local instance detected.' }
  },

  async getMemoryOverviewSnapshot() {
    return mockMemoryOverviewSnapshot
  },

  async rebuildMemoryOverviewSnapshot() {
    mockMemoryOverviewSnapshot = { ...mockMemoryOverviewSnapshot, rebuilt: true, updatedAt: new Date().toISOString() }
    return mockMemoryOverviewSnapshot
  },

  async getMemoryImpressionSnapshot() {
    return mockMemoryImpressionSnapshot
  },

  async rebuildMemoryImpressionSnapshot() {
    mockMemoryImpressionSnapshot = { ...mockMemoryImpressionSnapshot, rebuilt: true, updatedAt: new Date().toISOString() }
    return mockMemoryImpressionSnapshot
  },

  async getMemoryContent(uri: string, layer: 'overview' | 'read' = 'overview') {
    const hit = mockMemoryOverviewSnapshot.hits.find((item) => item.uri === uri)
    if (!hit) return ''
    return layer === 'read' ? `${hit.title}\n\n${hit.abstract}` : hit.abstract
  },

  async listMemoryWriteProposals() {
    return mockMemoryProposals.filter((proposal) => proposal.status === 'pending')
  },

  async updateMemoryWriteProposal(proposalId: string, input: { title: string; summary: string }) {
    mockMemoryProposals = mockMemoryProposals.map((proposal) => (proposal.id === proposalId ? { ...proposal, title: input.title.trim(), summary: input.summary.trim() } : proposal))
    const proposal = mockMemoryProposals.find((item) => item.id === proposalId)
    if (!proposal) throw new Error('Memory proposal not found.')
    return proposal
  },

  async approveMemoryWriteProposal(proposalId: string) {
    mockMemoryProposals = mockMemoryProposals.map((proposal) => (proposal.id === proposalId ? { ...proposal, status: 'approved', decidedAt: new Date().toISOString(), decisionReason: 'approved in settings' } : proposal))
    const proposal = mockMemoryProposals.find((item) => item.id === proposalId)
    if (!proposal) throw new Error('Memory proposal not found.')
    return proposal
  },

  async denyMemoryWriteProposal(proposalId: string) {
    mockMemoryProposals = mockMemoryProposals.map((proposal) => (proposal.id === proposalId ? { ...proposal, status: 'denied', decidedAt: new Date().toISOString(), decisionReason: 'denied in settings' } : proposal))
    const proposal = mockMemoryProposals.find((item) => item.id === proposalId)
    if (!proposal) throw new Error('Memory proposal not found.')
    return proposal
  },

  async listMCPServers() {
    return mockMCPServers
  },

  async saveMCPServer(input: MCPServerConfigInput) {
    const server: MCPServerStatus = {
      serverSafeId: `mcp:${input.slug}`,
      serverSlug: input.slug,
      displayName: input.displayName,
      transport: input.transport,
      enabled: input.enabled,
      configSource: 'local',
      discoveryStatus: input.enabled ? 'not_discovered' : 'disabled',
      candidateCount: 0,
      candidateNames: [],
      executionMode: 'disabled',
    }
    mockMCPServers = [...mockMCPServers.filter((item) => item.serverSlug !== input.slug), server].sort((a, b) => a.serverSlug.localeCompare(b.serverSlug))
    return server
  },

  async deleteMCPServer(slug: string) {
    mockMCPServers = mockMCPServers.filter((item) => item.serverSlug !== slug)
    return mockMCPServers
  },

  async discoverMCPServer(slug: string) {
    const server = mockMCPServers.find((item) => item.serverSlug === slug)
    if (!server) throw new Error('MCP server config was not found.')
    const discovered: MCPServerStatus = { ...server, discoveryStatus: server.enabled ? 'succeeded' : 'disabled', candidateNames: server.enabled ? [`mcp.${slug}.echo`] : [], candidateCount: server.enabled ? 1 : 0, executionMode: server.enabled ? 'approval_gated' : 'disabled', lastDiscoveredAt: new Date().toISOString() }
    mockMCPServers = mockMCPServers.map((item) => (item.serverSlug === slug ? discovered : item))
    return discovered
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
