import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('M4 run UI labels', () => {
  test('keeps local simulated language and stopped state visible without LLM/tool claims', () => {
    const chatCanvas = readFileSync(resolve(import.meta.dir, 'ChatCanvas.tsx'), 'utf8')
    const runRail = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')
    const dictionary = readFileSync(resolve(import.meta.dir, '../i18n.ts'), 'utf8')

    expect(chatCanvas).not.toContain('copy.localSimulated')
    expect(runRail).toContain('stopped')
    expect(chatCanvas).not.toContain('LLM')
    expect(dictionary).toContain("'worker-job': 'Worker/job'")
  })
})
