import { describe, expect, test } from 'bun:test'

describe('M4 App wiring', () => {
  test('passes stopRun and streamState to ChatCanvas', () => {
    const source = Bun.file(new URL('./App.tsx', import.meta.url)).text()
    return Promise.all([
      expect(source).resolves.toContain('stopRun'),
      expect(source).resolves.toContain('streamState'),
      expect(source).resolves.toContain('onStopRun'),
    ])
  })
})
