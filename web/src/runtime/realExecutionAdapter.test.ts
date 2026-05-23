import { describe, expect, test } from 'bun:test'
import { realExecutionAdapter } from './realExecutionAdapter'

describe('realExecutionAdapter', () => {
  test('reports M4 run/event runtime capability as available', () => {
    expect(realExecutionAdapter.runtimeCapability).toBe('available')
  })

  test('documents that execution behavior is delegated through the real API client', async () => {
    await expect(realExecutionAdapter.createRun('thread-a', 'msg-a', 'success')).rejects.toThrow('Use realApiClient.startRun')
    await expect(realExecutionAdapter.sendMessage('thread-a', 'hello')).rejects.toThrow('Use realApiClient.sendMessage')
  })
})
