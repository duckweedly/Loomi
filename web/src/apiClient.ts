import type { ApiReadiness, InstalledSkill, LocalProviderDetection, MCPServerConfigInput, MCPServerStatus, MemoryAuditItem, MemoryEntry, MemoryErrorEvent, MemoryFilters, MemoryImpressionSnapshot, MemoryOverviewSnapshot, MemoryProviderStatus, MemoryProviderUpdate, MemoryWriteProposal, Message, Persona, ProviderCapability, Run, Thread, ToolCall, ToolCatalogItem, WebSearchConfig, WorkerQueueDiagnostics, WorkspaceRootConfig } from './domain'
import type { PreviewArtifact } from './runtime/artifactPreview'
import { realApiClient } from './realApiClient'
import type { ExecutionAdapter } from './runtime/executionAdapter'
import { mockExecutionAdapter } from './runtime/mockExecutionAdapter'
import { realExecutionAdapter } from './runtime/realExecutionAdapter'

export type ApiClient = {
  mode: 'mock' | 'real_api'
  listThreads(): Promise<Thread[]>
  getThreadMessages(threadId: string): Promise<Message[]>
  getThreadRun(threadId: string, options?: { afterSequence?: number; existingEvents?: Run['events'] }): Promise<Run>
  getRunEvents(runId: string, afterSequence?: number): Promise<Run['events']>
  listArtifacts?(threadId: string): Promise<PreviewArtifact[]>
  readArtifact?(threadId: string, artifactId: string): Promise<PreviewArtifact>
  listPersonas?(): Promise<Persona[]>
  listSkills?(): Promise<InstalledSkill[]>
  listModelProviders?(): Promise<ProviderCapability[]>
  getReadiness?(): Promise<ApiReadiness>
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
  createMemoryEntry?(input: { title: string; content: string; scopeType?: 'user' | 'thread'; scopeId?: string }): Promise<MemoryEntry>
  searchMemory?(query: string, filters?: MemoryFilters): Promise<MemoryEntry[]>
  getMemoryEntry?(entryId: string, filters?: MemoryFilters): Promise<MemoryEntry>
  deleteMemoryEntry?(entryId: string, filters?: MemoryFilters): Promise<void>
  listMemoryAudit?(filters?: MemoryFilters): Promise<MemoryAuditItem[]>
  listMemoryWriteProposals?(filters?: MemoryFilters): Promise<MemoryWriteProposal[]>
  updateMemoryWriteProposal?(proposalId: string, input: { title: string; summary: string }): Promise<MemoryWriteProposal>
  approveMemoryWriteProposal?(proposalId: string): Promise<MemoryWriteProposal>
  denyMemoryWriteProposal?(proposalId: string): Promise<MemoryWriteProposal>
  getMemoryProviderStatus?(): Promise<MemoryProviderStatus>
  listMemoryErrors?(): Promise<MemoryErrorEvent[]>
  detectNowledgeMemoryProvider?(): Promise<{ detected: boolean; baseUrl?: string; message: string }>
  detectOpenVikingMemoryProvider?(): Promise<{ detected: boolean; baseUrl?: string; message: string }>
  updateMemoryProvider?(input: MemoryProviderUpdate): Promise<MemoryProviderStatus>
  getMemoryOverviewSnapshot?(): Promise<MemoryOverviewSnapshot>
  rebuildMemoryOverviewSnapshot?(): Promise<MemoryOverviewSnapshot>
  getMemoryImpressionSnapshot?(): Promise<MemoryImpressionSnapshot>
  rebuildMemoryImpressionSnapshot?(): Promise<MemoryImpressionSnapshot>
  getMemoryContent?(uri: string, layer?: 'overview' | 'read'): Promise<string>
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
