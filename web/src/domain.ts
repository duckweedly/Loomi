export type RunStatus = 'running' | 'completed' | 'stopped'

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
  type: string
  label: string
  detail: string
  time: string
  status: 'pending' | 'running' | 'completed' | 'stopped'
}

export type Run = {
  id: string
  threadId: string
  status: RunStatus
  model: string
  context: string
  events: RunEvent[]
}
