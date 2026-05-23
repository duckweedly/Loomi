import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('RunRail restrained runtime polish', () => {
  test('uses compact Scenario controls instead of prominent script buttons', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('Scenario')
    expect(source).toContain('Success')
    expect(source).toContain('Fail')
    expect(source).not.toContain('成功剧本')
    expect(source).not.toContain('失败剧本')
  })

  test('uses a ghost stop action label', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('Stop run')
    expect(source).toContain('runtime-stop-button ghost')
  })

  test('styles timeline with quiet dots and compact agent card', () => {
    const css = readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8')

    expect(css).toContain('.progress-row::before')
    expect(css).toContain('width: 7px')
    expect(css).toContain('.agent-motion-card.compact')
    expect(css).toContain('.runtime-script-switch.compact')
  })
})
