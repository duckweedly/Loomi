import { describe, expect, test } from 'bun:test'
import { rightPanelItems } from './rightPanelItems'

describe('rightPanelItems', () => {
  test('keeps the titlebar panel list scoped to preview only', () => {
    expect(rightPanelItems.map((item) => item.id)).toEqual(['preview'])
  })

  test('keeps the run details panel separate from tool placeholders', () => {
    expect(rightPanelItems.some((item) => item.id === 'run-details')).toBe(false)
  })
})
