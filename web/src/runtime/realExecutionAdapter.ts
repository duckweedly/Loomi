import type { ExecutionAdapter } from './executionAdapter'

function delegated(method: string): never {
  throw new Error(`Use realApiClient.${method} for M4 run/event execution`)
}

export type RealRuntimeCapabilitySignal = {
  backendUnavailable?: boolean
  modelSetupMissing?: boolean
  providerUnavailable?: boolean
  streamDisconnected?: boolean
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
