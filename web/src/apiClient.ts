import type { InstalledSkill, LocalProviderDetection, MCPServerConfigInput, MCPServerStatus, MemoryAuditItem, MemoryEntry, MemoryFilters, Message, Persona, ProviderCapability, Run, Thread, ToolCall, ToolCatalogItem, WebSearchConfig, WorkerQueueDiagnostics, WorkspaceRootConfig } from './domain'
import { realApiClient } from './realApiClient'
import type { ExecutionAdapter } from './runtime/executionAdapter'
import { mockExecutionAdapter } from './runtime/mockExecutionAdapter'
import { realExecutionAdapter } from './runtime/realExecutionAdapter'

export type ApiClient = {
  mode: 'mock' | 'real_api'
  listThreads(): Promise<Thread[]>
  getThreadMessages(threadId: string): Promise<Message[]>
  getThreadRun(threadId: string): Promise<Run>
  getRunEvents(runId: string): Promise<Run['events']>
  listPersonas?(): Promise<Persona[]>
  listSkills?(): Promise<InstalledSkill[]>
  listModelProviders?(): Promise<ProviderCapability[]>
  listToolCatalog?(): Promise<ToolCatalogItem[]>
  getWebSearchConfig?(): Promise<WebSearchConfig>
  saveWebSearchKeys?(input: { tavilyApiKey?: string; braveApiKey?: string }): Promise<WebSearchConfig>
  getWorkspaceRoot?(): Promise<WorkspaceRootConfig>
  saveWorkspaceRoot?(input: { path: string }): Promise<WorkspaceRootConfig>
  listMCPServers?(): Promise<MCPServerStatus[]>
  saveMCPServer?(input: MCPServerConfigInput): Promise<MCPServerStatus>
  deleteMCPServer?(slug: string): Promise<MCPServerStatus[]>
  discoverMCPServer?(slug: string): Promise<MCPServerStatus>
  listLocalProviderDetections?(): Promise<LocalProviderDetection[]>
  enableLocalProvider?(providerId: string): Promise<ProviderCapability>
  disableLocalProvider?(providerId: string): Promise<ProviderCapability>
  checkModelProvider?(providerId: string): Promise<ProviderCapability>
  saveModelProvider?(input: { baseUrl: string; model: string; apiKey: string }): Promise<ProviderCapability>
  getWorkerQueueDiagnostics?(): Promise<WorkerQueueDiagnostics>
  getToolCall?(threadId: string, runId: string, toolCallId: string): Promise<ToolCall>
  approveToolCall?(threadId: string, runId: string, toolCallId: string): Promise<ToolCall>
  denyToolCall?(threadId: string, runId: string, toolCallId: string): Promise<ToolCall>
  listMemoryEntries?(filters?: MemoryFilters): Promise<MemoryEntry[]>
  searchMemory?(query: string, filters?: MemoryFilters): Promise<MemoryEntry[]>
  getMemoryEntry?(entryId: string, filters?: MemoryFilters): Promise<MemoryEntry>
  deleteMemoryEntry?(entryId: string, filters?: MemoryFilters): Promise<void>
  listMemoryAudit?(filters?: MemoryFilters): Promise<MemoryAuditItem[]>
  startRun?(threadId: string, input?: { messageId?: string; source?: Run['source']; providerId?: string; model?: string; personaId?: string }): Promise<Run>
  subscribeRunEvents?(runId: string, afterSequence: number, onEvent: (event: Run['events'][number]) => void, onError: () => void, onClosed?: () => void): () => void
  createThread?(title: string, mode: Thread['mode']): Promise<Thread>
  updateThread?(threadId: string, input: Partial<Pick<Thread, 'title' | 'mode'>>): Promise<Thread>
  archiveThread?(threadId: string): Promise<Thread>
  sendMessage(threadId: string, content: string, personaId?: string, options?: { providerId?: string; model?: string }): Promise<{ messages: Message[]; run: Run }>
  stopRun(runId: string): Promise<Run>
}

export function selectExecutionAdapter(realApiMode: boolean): ExecutionAdapter {
  return realApiMode ? realExecutionAdapter : mockExecutionAdapter
}

export function createExecutionAdapter(): ExecutionAdapter {
  return realExecutionAdapter
}

export const apiClient: ApiClient = realApiClient
export const executionAdapter = createExecutionAdapter()
