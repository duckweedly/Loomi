import type { Message, Run, RuntimeEvent, RuntimeScriptId } from '../domain'
import type { ExecutionAdapter } from './executionAdapter'
import { createRuntimeEvent, getRuntimeScript, getRuntimeScriptSteps } from './runtimeScripts'

type MockExecutionStore = {
  id: number
  messages: Message[]
  runs: Record<string, Run>
}

function nextId(store: MockExecutionStore, prefix: string) {
  store.id += 1
  return `${prefix}-${store.id}`
}

function nowLabel() {
  return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function cloneRun(run: Run): Run {
  return { ...run, events: [...run.events], assistantDraft: run.assistantDraft ? { ...run.assistantDraft } : undefined }
}

function appendRuntimeEvent(run: Run, event: RuntimeEvent): Run {
  return {
    ...run,
    status: event.status,
    events: [...run.events, event],
    completedAt: event.status === 'completed' || event.status === 'failed' || event.status === 'stopped' ? event.time : run.completedAt,
  }
}

function appendAssistantDelta(run: Run, delta: string): Run {
  return {
    ...run,
    assistantDraft: {
      ...run.assistantDraft,
      content: `${run.assistantDraft?.content ?? ''}${delta}`,
      status: 'drafting',
    },
  }
}

export function createMockExecutionAdapter(store: MockExecutionStore = { id: 0, messages: [], runs: {} }): ExecutionAdapter {
  return {
    runtimeCapability: 'available',

    async sendMessage(threadId, content) {
      const message: Message = {
        id: nextId(store, 'msg'),
        threadId,
        role: 'user',
        content,
        createdAt: nowLabel(),
      }
      store.messages = [...store.messages, message]
      return message
    },

    async createRun(threadId, _messageId, scriptId: RuntimeScriptId = 'success') {
      const run: Run = {
        id: nextId(store, 'run'),
        threadId,
        status: 'pending',
        model: 'Mock Runtime',
        context: 'M3.5 mock',
        events: [],
        scriptId,
        assistantDraft: { content: '', status: 'empty' },
        createdAt: nowLabel(),
      }
      store.runs[run.id] = run
      return cloneRun(run)
    },

    async subscribeRunEvents(threadId, runId, onEvent) {
      const run = store.runs[runId]
      if (!run || run.threadId !== threadId) return () => {}
      getRuntimeScriptSteps(run.scriptId ?? 'success').forEach((step, index) => {
        const event = createRuntimeEvent({ threadId, runId, sequence: index, step })
        onEvent(event)
        if (store.runs[runId].status === 'failed' || store.runs[runId].status === 'stopped') return
        store.runs[runId] = appendRuntimeEvent(store.runs[runId], event)
        if (event.assistantDelta) store.runs[runId] = appendAssistantDelta(store.runs[runId], event.assistantDelta)
      })
      return () => {}
    },

    async appendAssistantDelta(threadId, runId, delta) {
      const run = store.runs[runId]
      if (!run || run.threadId !== threadId) throw new Error('Run not found')
      store.runs[runId] = appendAssistantDelta(run, delta)
      return cloneRun(store.runs[runId])
    },

    async completeRun(threadId, runId, finalAssistantContent) {
      const run = store.runs[runId]
      if (!run || run.threadId !== threadId) throw new Error('Run not found')
      const message: Message = {
        id: nextId(store, 'msg'),
        threadId,
        role: 'assistant',
        content: finalAssistantContent,
        createdAt: nowLabel(),
      }
      store.messages = [...store.messages, message]
      store.runs[runId] = {
        ...run,
        status: 'completed',
        completedAt: message.createdAt,
        assistantDraft: { content: finalAssistantContent, status: 'completed', messageId: message.id },
      }
      return { run: cloneRun(store.runs[runId]), message }
    },

    async failRun(threadId, runId, reason) {
      const run = store.runs[runId]
      if (!run || run.threadId !== threadId) throw new Error('Run not found')
      store.runs[runId] = {
        ...run,
        status: 'failed',
        assistantDraft: { content: run.assistantDraft?.content ?? '', status: 'failed' },
        events: [
          ...run.events,
          { id: `${runId}-failed`, runId, threadId, type: 'run.failed', label: 'Failed', detail: reason, time: 'Now', status: 'failed' },
        ],
      }
      return cloneRun(store.runs[runId])
    },

    async stopRun(threadId, runId) {
      const run = store.runs[runId]
      if (!run || run.threadId !== threadId) throw new Error('Run not found')
      store.runs[runId] = {
        ...run,
        status: 'stopped',
        assistantDraft: { content: run.assistantDraft?.content ?? '', status: 'stopped' },
        events: [
          ...run.events,
          { id: `${runId}-stopped`, runId, threadId, type: 'run.stopped', label: 'Stopped', detail: '已停止', time: 'Now', status: 'stopped' },
        ],
      }
      return cloneRun(store.runs[runId])
    },
  }
}

export const mockExecutionAdapter = createMockExecutionAdapter()
