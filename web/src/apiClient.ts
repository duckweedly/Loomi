import type { Message, Run, Thread } from './domain'
import { mockApiClient } from './mockApiClient'
import { hasRealApiBase, realApiClient } from './realApiClient'

export type ApiClient = {
  mode: 'mock' | 'real_api'
  listThreads(): Promise<Thread[]>
  getThreadMessages(threadId: string): Promise<Message[]>
  getThreadRun(threadId: string): Promise<Run>
  getRunEvents(runId: string): Promise<Run['events']>
  createThread?(title: string, mode: Thread['mode']): Promise<Thread>
  updateThread?(threadId: string, input: Partial<Pick<Thread, 'title' | 'mode'>>): Promise<Thread>
  archiveThread?(threadId: string): Promise<Thread>
  sendMessage(threadId: string, content: string): Promise<{ messages: Message[]; run: Run }>
  stopRun(runId: string): Promise<Run>
}

export const apiClient: ApiClient = hasRealApiBase() ? realApiClient : mockApiClient
