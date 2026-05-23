import { describe, expect, test } from 'bun:test'
import { selectExecutionAdapter } from '../apiClient'

describe('selectExecutionAdapter', () => {
  test('selects mock runtime capability for mock mode', () => {
    expect(selectExecutionAdapter(false).runtimeCapability).toBe('available')
  })

  test('selects M4 real runtime capability for configured real API mode', () => {
    expect(selectExecutionAdapter(true).runtimeCapability).toBe('available')
  })
})
