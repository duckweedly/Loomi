import type { ExecutionAdapter } from './executionAdapter'

function delegated(method: string): never {
  throw new Error(`Use realApiClient.${method} for M4 run/event execution`)
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
