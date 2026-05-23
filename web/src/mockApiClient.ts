import type { ApiClient } from './apiClient'
import type { Message, Run } from './domain'
import { messages, runs, threads } from './mockData'

let mockId = 0
let threadStore = [...threads]
let messageStore = [...messages]
let runStore = runs.map((run) => ({ ...run, events: [...run.events] }))

function nextMockId(prefix: string) {
  mockId += 1
  return `${prefix}-${mockId}`
}

function cloneRun(run: Run): Run {
  return { ...run, events: [...run.events] }
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

  async sendMessage(threadId: string, content: string) {
    const now = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    const userMessage: Message = {
      id: `msg-${nextMockId('mock')}`,
      threadId,
      role: 'user',
      content,
      createdAt: now,
    }
    const assistantMessage: Message = {
      id: `msg-${nextMockId('mock')}-assistant`,
      threadId,
      role: 'assistant',
      content: 'Queued. I will align the shell, update the run rail, and keep the workspace compact.',
      createdAt: now,
      toolCalls: [
        {
          id: `tool-${nextMockId('mock')}`,
          name: 'compose_workspace',
          status: 'running',
          summary: 'Updating mock workspace state.',
          input: content,
          output: 'Pending',
        },
      ],
    }
    messageStore = [...messageStore, userMessage, assistantMessage]

    const currentRun = runStore.find((run) => run.threadId === threadId)
    const run: Run = currentRun
      ? {
          ...currentRun,
          status: 'running',
          events: [
            ...currentRun.events,
            {
              id: `evt-${nextMockId('mock')}`,
              type: 'message.queued',
              label: 'Queued',
              detail: 'New message received',
              time: 'Now',
              status: 'running',
            },
          ],
        }
      : {
          ...createIdleRun(threadId),
          status: 'running',
          context: '12k / 128k',
          events: [
            {
              id: `evt-${nextMockId('mock')}`,
              type: 'message.queued',
              label: 'Queued',
              detail: 'New message received',
              time: 'Now',
              status: 'running',
            },
          ],
        }
    runStore = currentRun ? runStore.map((item) => (item.threadId === threadId ? run : item)) : [run, ...runStore]
    threadStore = threadStore.map((thread) => (thread.id === threadId ? { ...thread, runStatus: 'running', updatedAt: 'Now' } : thread))

    return { messages: await this.getThreadMessages(threadId), run: cloneRun(run) }
  },

  async stopRun(runId: string) {
    const run = runStore.find((item) => item.id === runId)
    if (!run) throw new Error('Run not found')
    const stopped: Run = {
      ...run,
      status: 'stopped',
      events: [
        ...run.events,
        {
          id: `evt-${nextMockId('mock')}`,
          type: 'run.stopped',
          label: 'Stopped',
          detail: 'Stopped by user',
          time: 'Now',
          status: 'stopped',
        },
      ],
    }
    runStore = runStore.map((item) => (item.id === runId ? stopped : item))
    threadStore = threadStore.map((thread) => (thread.id === stopped.threadId ? { ...thread, runStatus: 'stopped' } : thread))
    return cloneRun(stopped)
  },
}
