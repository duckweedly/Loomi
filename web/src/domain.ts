export type RunStatus = 'pending' | 'running' | 'completed' | 'failed' | 'stopped'

export type RuntimeStatus = RunStatus

export type RunEventCategory = 'lifecycle' | 'progress' | 'message' | 'error' | 'final'

export type StreamState = 'connecting' | 'live' | 'recoverable_error' | 'closed'

export type BackendCapabilityState = 'available' | 'unavailable'

export type ChatCanvasState =
  | 'no-thread'
  | 'empty-thread'
  | 'loading'
  | 'error'
  | 'history'
  | 'waiting-run'
  | 'running'
  | 'completed'
  | 'failed'
  | 'backend-unavailable'

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

export type RuntimeEventType =
  | 'run.created'
  | 'context.loading'
  | 'assistant.thinking'
  | 'assistant.drafting'
  | 'assistant.message.completed'
  | 'run.completed'
  | 'run.failed'
  | 'run.stopped'
  | string

export type RuntimeEvent = {
  id: string
  runId: string
  threadId: string
  sequence?: number
  category?: RunEventCategory
  type: RuntimeEventType
  label: string
  detail: string
  content?: string | null
  time: string
  status: RuntimeStatus
  assistantDelta?: string
}

export type RunEvent = Omit<RuntimeEvent, 'runId' | 'threadId'> & {
  runId?: string
  threadId?: string
}

export type AssistantDraft = {
  content: string
  status: 'empty' | 'drafting' | 'completed' | 'failed' | 'stopped'
  messageId?: string
}

export type Run = {
  id: string
  threadId: string
  status: RunStatus
  model: string
  context: string
  events: RunEvent[]
  scriptId?: RuntimeScriptId
  assistantDraft?: AssistantDraft
  createdAt?: string
  completedAt?: string
}

export type RuntimeScriptId = 'success' | 'failure'

export type RuntimeScriptStep = {
  type: RuntimeEventType
  label: string
  detail: string
  status: RuntimeStatus
  assistantDelta?: string
}

export type RuntimeScript = {
  id: RuntimeScriptId
  name: string
  steps: RuntimeScriptStep[]
  finalAssistantMessage?: string
  terminalStatus: Extract<RuntimeStatus, 'completed' | 'failed'>
}

export type ThreadRuntimeState = {
  activeRunId: string | null
  runsById: Record<string, Run>
  selectedScriptId: RuntimeScriptId
  backendCapability: BackendCapabilityState
  lastFailureReason?: string
}

export type RuntimeStateByThread = Record<string, ThreadRuntimeState>

export type StaleEventGuard = {
  requestedThreadId: string
  currentSelectedThreadId: string
  runId: string
  activeRunId: string | null
}
