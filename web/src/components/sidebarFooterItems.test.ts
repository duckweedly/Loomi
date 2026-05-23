import { describe, expect, test } from 'bun:test'
import { createSidebarFooterItems } from './sidebarFooterItems'

describe('createSidebarFooterItems', () => {
  test('keeps only the settings entry in the sidebar footer', () => {
    expect(createSidebarFooterItems().map((item) => item.id)).toEqual(['settings'])
  })
})
