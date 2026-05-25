import type { LocalProviderDetection, MemoryAuditItem, MemoryEntry, MemoryFilters, Message, Persona, ProviderCapability, Run, Thread, ToolCall, ToolCatalogItem, WorkerQueueDiagnostics } from './domain'
import { mockApiClient } from './mockApiClient'
import { hasRealApiBase, realApiClient } from './realApiClient'
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
  listModelProviders?(): Promise<ProviderCapability[]>
  listToolCatalog?(): Promise<ToolCatalogItem[]>
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
  subscribeRunEvents?(runId: string, afterSequence: number, onEvent: (event: Run['events'][number]) => void, onError: () => void): () => void
  createThread?(title: string, mode: Thread['mode']): Promise<Thread>
  updateThread?(threadId: string, input: Partial<Pick<Thread, 'title' | 'mode'>>): Promise<Thread>
  archiveThread?(threadId: string): Promise<Thread>
  sendMessage(threadId: string, content: string, personaId?: string): Promise<{ messages: Message[]; run: Run }>
  stopRun(runId: string): Promise<Run>
}

export function selectExecutionAdapter(realApiMode: boolean): ExecutionAdapter {
  return realApiMode ? realExecutionAdapter : mockExecutionAdapter
}

export function createExecutionAdapter(): ExecutionAdapter {
  return selectExecutionAdapter(hasRealApiBase())
}

export const apiClient: ApiClient = hasRealApiBase() ? realApiClient : mockApiClient
export const executionAdapter = createExecutionAdapter()
