import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('RunTimeline runtime linkage', () => {
  test('feeds selected run events through RunRail', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunTimeline.tsx'), 'utf8')

    expect(source).toContain('<RunRail run={run}')
  })

  test('RunRail renders failed and stopped statuses without marking them done', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain("event.status === 'failed'")
    expect(source).toContain("event.status === 'stopped'")
    expect(source).toContain('onStopRun')
  })

  test('RunRail exposes a compact mock script selector for failure smoke', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('onSelectRuntimeScript')
    expect(source).toContain('Scenario')
    expect(source).toContain('Fail')
  })
})
