import { describe, expect, test } from 'bun:test'
import { realExecutionAdapter, mapRealRuntimeCapabilitySignal } from './realExecutionAdapter'

describe('realExecutionAdapter', () => {
  test('reports M4 run/event runtime capability as available', () => {
    expect(realExecutionAdapter.runtimeCapability).toBe('available')
  })

  test('documents that execution behavior is delegated through the real API client', async () => {
    await expect(realExecutionAdapter.createRun('thread-a', 'msg-a', 'success')).rejects.toThrow('Use realApiClient.startRun')
    await expect(realExecutionAdapter.sendMessage('thread-a', 'hello')).rejects.toThrow('Use realApiClient.sendMessage')
  })

  test('maps backend setup provider and stream failures to capability signals', () => {
    expect(mapRealRuntimeCapabilitySignal(new Error('Failed to fetch'))).toEqual({ backendUnavailable: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('model setup missing'), { code: 'model_setup_missing' }))).toEqual({ modelSetupMissing: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('provider unavailable'), { code: 'provider_unavailable' }))).toEqual({ providerUnavailable: true })
    expect(mapRealRuntimeCapabilitySignal(Object.assign(new Error('stream disconnected'), { code: 'stream_disconnected' }))).toEqual({ streamDisconnected: true })
  })
})
