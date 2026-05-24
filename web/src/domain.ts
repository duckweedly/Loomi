export type RunStatus = 'pending' | 'queued' | 'running' | 'recovering' | 'stopping' | 'completed' | 'failed' | 'stopped' | 'cancelled' | 'retrying'

export type RuntimeStatus = RunStatus

export type RuntimeEventGroup = 'run-lifecycle' | 'model-stream' | 'worker-job' | 'error'

export type RuntimeEventSeverity = 'info' | 'progress' | 'warning' | 'error'

export type RuntimeUsage = {
  inputTokens?: number
  outputTokens?: number
  totalTokens?: number
}

export type RunEventCategory = 'lifecycle' | 'progress' | 'message' | 'error' | 'final'

export type StreamState = 'connecting' | 'live' | 'recoverable_error' | 'closed'

export type BackendCapabilityState = 'available' | 'unavailable' | 'misconfigured'

export type ProviderFamily = 'anthropic' | 'openai' | 'gemini' | 'openai_compatible'

export type ProviderCapability = {
  id: string
  family: ProviderFamily
  baseUrl?: string | null
  model: string
  status: BackendCapabilityState
  message?: string | null
}

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
  | 'stopped'
  | 'recovering'
  | 'stopping'

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
  runId?: string
  attemptOfMessageId?: string
  toolCalls?: ToolCall[]
}

export type RuntimeEventType =
  | 'run.created'
  | 'run.queued'
  | 'run.stopping'
  | 'context.loading'
  | 'assistant.thinking'
  | 'assistant.drafting'
  | 'assistant.message.completed'
  | 'job.claimed'
  | 'job.recovering'
  | 'job.retry_scheduled'
  | 'job.attempt_failed'
  | 'job.retry_exhausted'
  | 'worker.lease_renewed'
  | 'pipeline.step.started'
  | 'pipeline.step.completed'
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
  group?: RuntimeEventGroup
  severity?: RuntimeEventSeverity
  usage?: RuntimeUsage
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
  status: 'empty' | 'drafting' | 'pending' | 'queued' | 'streaming' | 'completed' | 'failed' | 'stopped' | 'recovering' | 'stopping'
  messageId?: string
  lastEventId?: string
}

export type RunSource = 'local_simulated' | 'model_gateway'

export type Run = {
  id: string
  threadId: string
  status: RunStatus
  model: string
  context: string
  source?: RunSource
  events: RunEvent[]
  scriptId?: RuntimeScriptId
  attemptOfMessageId?: string
  assistantDraft?: AssistantDraft
  createdAt?: string
  completedAt?: string
}

export type RuntimeScriptId = 'success' | 'failure' | 'model-stream' | 'model-error' | 'stopped' | 'replayed'

export type RuntimeScriptStep = {
  type: RuntimeEventType
  label: string
  detail: string
  status: RuntimeStatus
  group?: RuntimeEventGroup
  severity?: RuntimeEventSeverity
  usage?: RuntimeUsage
  assistantDelta?: string
}

export type RuntimeScript = {
  id: RuntimeScriptId
  name: string
  steps: RuntimeScriptStep[]
  finalAssistantMessage?: string
  terminalStatus: Extract<RuntimeStatus, 'completed' | 'failed' | 'stopped'>
}

export type WorkerQueueStatus = 'ready' | 'paused' | 'unhealthy' | 'degraded'

export type WorkerStatus = 'ready' | 'paused' | 'unhealthy' | 'degraded' | 'stopped'

export type WorkerQueueDiagnostics = {
  queueStatus: WorkerQueueStatus
  workerStatus: WorkerStatus
  queuedCount: number
  leasedCount: number
  staleCount: number
  retryingCount: number
  deadCount: number
  updatedAt: string
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
