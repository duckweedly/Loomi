import type { BackendCapabilityState, Message, Run, RuntimeEvent, RuntimeScriptId } from '../domain'

export type ExecutionAdapter = {
  readonly runtimeCapability: BackendCapabilityState
  sendMessage(threadId: string, content: string): Promise<Message>
  createRun(threadId: string, messageId: string, scriptId?: RuntimeScriptId): Promise<Run>
  subscribeRunEvents(threadId: string, runId: string, onEvent: (event: RuntimeEvent) => void): Promise<() => void>
  appendAssistantDelta(threadId: string, runId: string, delta: string): Promise<Run>
  completeRun(threadId: string, runId: string, finalAssistantContent: string): Promise<{ run: Run; message: Message }>
  failRun(threadId: string, runId: string, reason: string): Promise<Run>
  stopRun(threadId: string, runId: string): Promise<Run>
}

export type ExecutionAdapterMode = 'mock' | 'real_api'

export function isRuntimeTerminal(status: Run['status']) {
  return status === 'completed' || status === 'failed' || status === 'stopped'
}
