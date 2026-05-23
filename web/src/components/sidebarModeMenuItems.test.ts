import { describe, expect, test } from 'bun:test'
import { createSidebarModeMenuItems } from './sidebarModeMenuItems'

describe('createSidebarModeMenuItems', () => {
  test('uses a single chat creation row in chat mode', () => {
    const items = createSidebarModeMenuItems('chat')

    expect(items.map((item) => item.id)).toEqual(['new-chat'])
    expect(items.map((item) => item.label)).toEqual(['New Chat'])
    expect(items.some((item) => item.label === 'Search')).toBe(false)
  })

  test('uses project and schedule rows in work mode', () => {
    const items = createSidebarModeMenuItems('work')

    expect(items.map((item) => item.id)).toEqual(['projects', 'scheduled'])
    expect(items.map((item) => item.label)).toEqual(['Projects', 'Scheduled'])
    expect(items.some((item) => item.label === 'Search')).toBe(false)
  })
})
