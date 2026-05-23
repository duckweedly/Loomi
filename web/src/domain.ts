export type RunStatus = 'pending' | 'running' | 'completed' | 'failed' | 'stopped'

export type RunEventCategory = 'lifecycle' | 'progress' | 'message' | 'error' | 'final'

export type StreamState = 'connecting' | 'live' | 'recoverable_error' | 'closed'

export type Thread = {
  id: string
  title: string
  project: string
  mode: 'chat' | 'work'
  updatedAt: string
  lifecycleStatus?: 'active' | 'archived'
  runStatus: RunStatus
}

export type ToolCall = {
  id: string
  name: string
  status: 'running' | 'completed'
  summary: string
  input: string
  output: string
}

export type Message = {
  id: string
  threadId: string
  role: 'user' | 'assistant'
  content: string
  createdAt: string
  toolCalls?: ToolCall[]
}

export type RunEvent = {
  id: string
  runId?: string
  threadId?: string
  sequence?: number
  category?: RunEventCategory
  type: string
  label: string
  detail: string
  content?: string | null
  time: string
  status: RunStatus
}

export type Run = {
  id: string
  threadId: string
  status: RunStatus
  model: string
  context: string
  events: RunEvent[]
}
