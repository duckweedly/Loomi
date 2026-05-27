import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('AgentStateMotion reduced motion CSS', () => {
  test('disables badge animations when the user prefers reduced motion', () => {
    const css = [
      readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/30-runtime-panels.css'), 'utf8'),
    ].join('\n')

    expect(css).toContain('@media (prefers-reduced-motion: reduce)')
    expect(css).toContain('.agent-motion-card *')
    expect(css).toContain('animation: none !important;')
  })
})
