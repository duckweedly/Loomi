import type { ExecutionAdapter } from './executionAdapter'

function unavailable(): never {
  throw new Error('后端运行能力未接入')
}

export const realExecutionAdapter: ExecutionAdapter = {
  runtimeCapability: 'unavailable',
  async sendMessage() {
    unavailable()
  },
  async createRun() {
    unavailable()
  },
  async subscribeRunEvents() {
    return () => {}
  },
  async appendAssistantDelta() {
    unavailable()
  },
  async completeRun() {
    unavailable()
  },
  async failRun() {
    unavailable()
  },
  async stopRun() {
    unavailable()
  },
}
