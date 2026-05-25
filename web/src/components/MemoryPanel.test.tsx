import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryPanel } from './MemoryPanel'

describe('MemoryPanel', () => {
  test('renders safe summaries and delete controls', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        entries={[{
          id: 'mem_1',
          title: 'Preference',
          summary: 'Prefers safe snapshots',
          scopeType: 'user',
          createdAt: '2026-05-25T00:00:00Z',
          updatedAt: '2026-05-25T00:00:01Z',
          redactionApplied: true,
        }]}
        onQueryChange={() => {}}
        onDelete={() => {}}
      />,
    )

    expect(html).toContain('Prefers safe snapshots')
    expect(html).toContain('Delete Preference')
  })
})
