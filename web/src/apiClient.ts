import type { Message, Run, Thread } from './domain'
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
  startRun?(threadId: string): Promise<Run>
  subscribeRunEvents?(runId: string, afterSequence: number, onEvent: (event: Run['events'][number]) => void, onError: () => void): () => void
  createThread?(title: string, mode: Thread['mode']): Promise<Thread>
  updateThread?(threadId: string, input: Partial<Pick<Thread, 'title' | 'mode'>>): Promise<Thread>
  archiveThread?(threadId: string): Promise<Thread>
  sendMessage(threadId: string, content: string): Promise<{ messages: Message[]; run: Run }>
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
