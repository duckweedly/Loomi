import type { ApiClient } from './apiClient'
import type { Message, Run, RuntimeScriptId } from './domain'
import { messages, runs, threads } from './mockData'
import { mockExecutionAdapter } from './runtime/mockExecutionAdapter'

let mockId = 0
let threadStore = [...threads]
let messageStore = [...messages]
let runStore = runs.map((run) => ({ ...run, events: [...run.events] }))
let selectedRuntimeScriptId: RuntimeScriptId = 'success'

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

export function setMockRuntimeScript(scriptId: RuntimeScriptId) {
  selectedRuntimeScriptId = scriptId
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
    const userMessage = await mockExecutionAdapter.sendMessage(threadId, content)
    messageStore = [...messageStore, userMessage]

    const run = await mockExecutionAdapter.createRun(threadId, userMessage.id, selectedRuntimeScriptId)
    updateRunStore(run)
    await mockExecutionAdapter.subscribeRunEvents(threadId, run.id, (event) => {
      const current = runStore.find((item) => item.id === run.id) ?? run
      const next: Run = {
        ...current,
        status: event.status,
        events: [...current.events, event],
        assistantDraft: event.assistantDelta
          ? { content: `${current.assistantDraft?.content ?? ''}${event.assistantDelta}`, status: 'drafting' }
          : current.assistantDraft,
        completedAt: event.status === 'completed' || event.status === 'failed' || event.status === 'stopped' ? event.time : current.completedAt,
      }
      updateRunStore(next)
    })

    const scriptRun = runStore.find((item) => item.id === run.id) ?? run
    let finalRun = scriptRun
    if (selectedRuntimeScriptId === 'success') {
      const completed = await mockExecutionAdapter.completeRun(threadId, run.id, scriptRun.assistantDraft?.content || '已完成一次模拟执行。')
      messageStore = [...messageStore, completed.message]
      finalRun = completed.run
      updateRunStore(finalRun)
    } else if (selectedRuntimeScriptId === 'failure') {
      finalRun = { ...scriptRun, assistantDraft: { content: scriptRun.assistantDraft?.content ?? '', status: 'failed' } }
      updateRunStore(finalRun)
    }

    threadStore = threadStore.map((thread) => (thread.id === threadId ? { ...thread, runStatus: finalRun.status, updatedAt: 'Now' } : thread))
    return { messages: await this.getThreadMessages(threadId), run: cloneRun(finalRun) }
  },

  async stopRun(runId: string) {
    const run = runStore.find((item) => item.id === runId)
    if (!run) throw new Error('Run not found')
    const stopped = await mockExecutionAdapter.stopRun(run.threadId, runId)
    updateRunStore(stopped)
    threadStore = threadStore.map((thread) => (thread.id === stopped.threadId ? { ...thread, runStatus: 'stopped' } : thread))
    return cloneRun(stopped)
  },
}
