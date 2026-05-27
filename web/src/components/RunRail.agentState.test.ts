import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('RunRail agent state integration', () => {
  test('renders the compact agent state motion badge inside the progress list', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain("import { AgentStateMotion } from './AgentStateMotion'")
    expect(source).toContain('<AgentStateMotion run={run} compact locale={locale} />')
  })
})
