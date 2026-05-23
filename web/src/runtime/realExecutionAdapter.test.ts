import { describe, expect, test } from 'bun:test'
import { realExecutionAdapter } from './realExecutionAdapter'

describe('realExecutionAdapter', () => {
  test('reports unavailable runtime capability', () => {
    expect(realExecutionAdapter.runtimeCapability).toBe('unavailable')
  })

  test('does not execute mock runtime behavior while backend runtime is unavailable', async () => {
    await expect(realExecutionAdapter.createRun('thread-a', 'msg-a', 'success')).rejects.toThrow('后端运行能力未接入')
    await expect(realExecutionAdapter.sendMessage('thread-a', 'hello')).rejects.toThrow('后端运行能力未接入')
  })
})
