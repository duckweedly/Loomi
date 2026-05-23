import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('Vite production chunking', () => {
  test('splits stable vendor groups away from the app entry chunk', () => {
    const source = readFileSync(resolve(import.meta.dir, 'vite.config.ts'), 'utf8')

    expect(source).toContain('rolldownOptions')
    expect(source).toContain('codeSplitting')
    expect(source).toContain("name: 'react-vendor'")
    expect(source).toContain("name: 'ui-vendor'")
    expect(source).toContain("name: 'icons-vendor'")
  })
})
