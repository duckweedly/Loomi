import type { Run, RuntimeEvent } from '../domain'
import type { ExecutionAdapter } from './executionAdapter'
import { isRuntimeTerminal } from './executionAdapter'

function delegated(method: string): never {
  throw new Error(`Use realApiClient.${method} for M4 run/event execution`)
}

export function applyRealRunEvent(run: Run, event: RuntimeEvent): Run {
  if (isRuntimeTerminal(run.status)) return run
  const events = [...run.events, event]
  const completedAt = event.status === 'completed' || event.status === 'failed' || event.status === 'stopped' ? event.time : run.completedAt
  if (event.type === 'message.model_output_delta') {
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: `${run.assistantDraft?.content ?? ''}${event.assistantDelta ?? event.content ?? ''}`,
        status: 'drafting',
      },
    }
  }
  if (event.type === 'message.model_output_completed') {
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: event.content ?? run.assistantDraft?.content ?? '',
        status: 'completed',
      },
    }
  }
  return { ...run, status: event.status, events }
}

export const realExecutionAdapter: ExecutionAdapter = {
  runtimeCapability: 'available',
  async sendMessage() {
    delegated('sendMessage')
  },
  async createRun() {
    delegated('startRun')
  },
  async subscribeRunEvents() {
    return () => {}
  },
  async appendAssistantDelta() {
    delegated('subscribeRunEvents')
  },
  async completeRun() {
    delegated('subscribeRunEvents')
  },
  async failRun() {
    delegated('subscribeRunEvents')
  },
  async stopRun() {
    delegated('stopRun')
  },
}
