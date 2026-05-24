import type { Run, RuntimeEvent } from '../domain'
import type { ExecutionAdapter } from './executionAdapter'
import { isRuntimeTerminal } from './executionAdapter'

function delegated(method: string): never {
  throw new Error(`Use realApiClient.${method} for M4 run/event execution`)
}

export type RealRuntimeCapabilitySignal = {
  backendUnavailable?: boolean
  modelSetupMissing?: boolean
  providerUnavailable?: boolean
  streamDisconnected?: boolean
}

export function applyRealRunEvent(run: Run, event: RuntimeEvent): Run {
  if (isRuntimeTerminal(run.status)) return run
  if (run.events.some((existing) => existing.id === event.id)) return run

  const events = [...run.events, event].sort((a, b) => (a.sequence ?? 0) - (b.sequence ?? 0))
  const completedAt = event.status === 'completed' || event.status === 'failed' || event.status === 'stopped' ? event.time : run.completedAt
  if (event.type === 'model.delta' || event.type === 'message.model_output_delta') {
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: `${run.assistantDraft?.content ?? ''}${event.assistantDelta ?? event.content ?? ''}`,
        status: 'streaming',
        lastEventId: event.id,
      },
    }
  }
  if (event.type === 'assistant.message.completed' || event.type === 'message.model_output_completed' || event.type === 'model.final') {
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: event.content ?? run.assistantDraft?.content ?? '',
        status: 'completed',
        lastEventId: event.id,
      },
    }
  }
  if (event.status === 'failed' || event.status === 'stopped') {
    return {
      ...run,
      status: event.status,
      events,
      completedAt,
      assistantDraft: {
        ...run.assistantDraft,
        content: run.assistantDraft?.content ?? event.content ?? '',
        status: event.status,
        lastEventId: event.id,
      },
    }
  }
  if (event.status === 'recovering') {
    return {
      ...run,
      status: 'recovering',
      events,
      assistantDraft: {
        ...run.assistantDraft,
        content: run.assistantDraft?.content ?? event.content ?? '',
        status: 'recovering',
        lastEventId: event.id,
      },
    }
  }
  return { ...run, status: event.status, events }
}

export function mapRealRuntimeCapabilitySignal(error: unknown): RealRuntimeCapabilitySignal {
  const code = typeof error === 'object' && error !== null && 'code' in error ? String(error.code) : ''
  const message = error instanceof Error ? error.message.toLowerCase() : ''
  if (code === 'stream_disconnected') return { streamDisconnected: true }
  if (code === 'provider_unavailable' || message.includes('provider')) return { providerUnavailable: true }
  if (code === 'model_setup_missing' || message.includes('model setup')) return { modelSetupMissing: true }
  return { backendUnavailable: true }
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
