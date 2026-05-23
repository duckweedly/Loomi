import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

describe('ThreadSidebar layout CSS', () => {
  test('allocates a dedicated grid column for the archive action', () => {
    const css = readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8')

    expect(css).toContain('grid-template-columns: minmax(0, 1fr) 24px;')
    expect(css).toContain('grid-template-columns: 10px minmax(0, 1fr);')
    expect(css).toContain('.thread-row > .thread-action')
  })
})
