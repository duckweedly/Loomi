import { describe, expect, test } from 'bun:test'
import type { Thread } from './domain'
import { createNextThreadTitle } from './threadTitles'

const thread = (title: string): Thread => ({
  id: title,
  title,
  project: 'Loomi',
  mode: 'chat',
  updatedAt: 'Now',
  lifecycleStatus: 'active',
  runStatus: 'completed',
})

describe('createNextThreadTitle', () => {
  test('uses the base title when no matching thread exists', () => {
    expect(createNextThreadTitle([thread('M1 interface pass')])).toBe('New thread')
  })

  test('increments the title after existing New thread entries', () => {
    expect(createNextThreadTitle([
      thread('New thread'),
      thread('Roadmap alignment'),
      thread('New thread 2'),
    ])).toBe('New thread 3')
  })
})
