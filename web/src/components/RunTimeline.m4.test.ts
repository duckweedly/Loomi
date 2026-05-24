import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('M4 run UI labels', () => {
  test('keeps local simulated language and stopped state visible without LLM/tool claims', () => {
    const chatCanvas = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    const runRail = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')
    const eventGroups = readFileSync(resolve(import.meta.dir, '../runtime/runtimeEventGroups.ts'), 'utf8')

    expect(chatCanvas).toContain('Local simulated')
    expect(runRail).toContain('stopped')
    expect(chatCanvas).not.toContain('LLM')
    expect(eventGroups).toContain("title: 'Worker/job'")
  })
})
