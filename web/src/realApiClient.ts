import type { ApiClient } from './apiClient'
import type { Message, Run, Thread } from './domain'

const apiBaseUrl = (import.meta.env.VITE_LOOMI_API_BASE_URL ?? '').replace(/\/$/, '')

export function hasRealApiBase() {
  return apiBaseUrl.length > 0
}

type ApiThread = {
  id: string
  title: string
  mode: 'chat' | 'work'
  lifecycle_status: 'active' | 'archived'
  created_at: string
  updated_at: string
  archived_at?: string | null
}

type ApiMessage = {
  id: string
  thread_id: string
  role: 'user'
  content: string
  created_at: string
  client_message_id?: string | null
}

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })
  const body = await response.json().catch(() => null)
  if (!response.ok) {
    const message = body?.error?.message ?? `Request failed with ${response.status}`
    throw new Error(message)
  }
  return body as T
}

function mapThread(thread: ApiThread): Thread {
  return {
    id: thread.id,
    title: thread.title,
    project: 'Loomi',
    mode: thread.mode,
    updatedAt: thread.updated_at,
    lifecycleStatus: thread.lifecycle_status,
    runStatus: 'completed',
  }
}

function mapMessage(message: ApiMessage): Message {
  return {
    id: message.id,
    threadId: message.thread_id,
    role: message.role,
    content: message.content,
    createdAt: message.created_at,
  }
}

export function createClientMessageID() {
  const random = globalThis.crypto?.randomUUID?.() ?? Math.random().toString(36).slice(2)
  return `web-${Date.now()}-${random}`
}

function deferredRun(threadId: string): Run {
  return {
    id: `deferred-${threadId}`,
    threadId,
    status: 'completed',
    model: 'Deferred',
    context: 'M3 thread/message only',
    events: [],
  }
}

export const realApiClient: ApiClient = {
  mode: 'real_api',

  async listThreads() {
    const body = await requestJSON<{ threads: ApiThread[] }>('/v1/threads')
    return body.threads.map(mapThread)
  },

  async getThreadMessages(threadId: string) {
    const body = await requestJSON<{ messages: ApiMessage[] }>(`/v1/threads/${threadId}/messages`)
    return body.messages.map(mapMessage)
  },

  async getThreadRun(threadId: string) {
    return deferredRun(threadId)
  },

  async getRunEvents() {
    return []
  },

  async createThread(title: string, mode: Thread['mode']) {
    const body = await requestJSON<{ thread: ApiThread }>('/v1/threads', {
      method: 'POST',
      body: JSON.stringify({ title, mode }),
    })
    return mapThread(body.thread)
  },

  async updateThread(threadId: string, input: Partial<Pick<Thread, 'title' | 'mode'>>) {
    const body = await requestJSON<{ thread: ApiThread }>(`/v1/threads/${threadId}`, {
      method: 'PATCH',
      body: JSON.stringify(input),
    })
    return mapThread(body.thread)
  },

  async archiveThread(threadId: string) {
    const body = await requestJSON<{ thread: ApiThread }>(`/v1/threads/${threadId}/archive`, { method: 'POST' })
    return mapThread(body.thread)
  },

  async sendMessage(threadId: string, content: string) {
    await requestJSON<{ message: ApiMessage }>(`/v1/threads/${threadId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content, client_message_id: createClientMessageID() }),
    })
    return {
      messages: await this.getThreadMessages(threadId),
      run: deferredRun(threadId),
    }
  },

  async stopRun(runId: string) {
    return deferredRun(runId.replace(/^deferred-/, ''))
  },
}
