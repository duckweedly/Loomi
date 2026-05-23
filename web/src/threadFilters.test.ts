import { describe, expect, test } from 'bun:test'
import type { Thread } from './domain'
import { filterThreadsByMode } from './threadFilters'

const thread = (id: string, mode: Thread['mode']): Thread => ({
  id,
  title: id,
  project: 'Loomi',
  mode,
  updatedAt: 'Now',
  lifecycleStatus: 'active',
  runStatus: 'completed',
})

describe('filterThreadsByMode', () => {
  test('keeps chat and work recent threads separate', () => {
    const threads = [thread('chat-a', 'chat'), thread('work-a', 'work'), thread('chat-b', 'chat')]

    expect(filterThreadsByMode(threads, 'chat').map((item) => item.id)).toEqual(['chat-a', 'chat-b'])
    expect(filterThreadsByMode(threads, 'work').map((item) => item.id)).toEqual(['work-a'])
  })
})
