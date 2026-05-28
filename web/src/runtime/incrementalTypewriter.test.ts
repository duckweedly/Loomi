import { describe, expect, test } from 'bun:test'
import { nextTypewriterFrame } from './incrementalTypewriter'

describe('nextTypewriterFrame', () => {
  test('reveals only the changed segment between two status labels', () => {
    expect(nextTypewriterFrame('组织回复...', '组织回复 12s', 4)).toBe('组织回复 1...')
    expect(nextTypewriterFrame('梳理线索 9s', '梳理线索 10s', 20)).toBe('梳理线索 10s')
  })
})
