import { describe, expect, test } from 'bun:test'
import { createClientMessageID } from './realApiClient'

describe('createClientMessageID', () => {
  test('does not rely on Date.now alone', () => {
    const originalNow = Date.now
    Date.now = () => 123
    try {
      const first = createClientMessageID()
      const second = createClientMessageID()

      expect(first).toStartWith('web-123-')
      expect(second).toStartWith('web-123-')
      expect(second).not.toBe(first)
    } finally {
      Date.now = originalNow
    }
  })
})
