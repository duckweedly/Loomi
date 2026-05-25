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
          status: 'approved',
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

  test('renders error state instead of stale entries', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query="safe"
        error="Memory failed to load"
        entries={[{
          id: 'mem_1',
          title: 'Stale',
          summary: 'Should not masquerade as success',
          scopeType: 'user',
          status: 'approved',
          createdAt: '2026-05-25T00:00:00Z',
          updatedAt: '2026-05-25T00:00:01Z',
          redactionApplied: false,
        }]}
        onQueryChange={() => {}}
        onDelete={() => {}}
      />,
    )

    expect(html).toContain('Memory failed to load')
    expect(html).toContain('Memory could not be loaded')
    expect(html).not.toContain('Should not masquerade as success')
  })
})
