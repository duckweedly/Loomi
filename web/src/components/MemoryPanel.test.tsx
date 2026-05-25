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
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(html).toContain('Prefers safe snapshots')
    expect(html).toContain('Delete Preference')
  })

  test('renders grounded search filters and safe metadata only', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query="preference"
        filters={{ scopeType: 'thread', scopeId: 'thread_1', sourceThreadId: 'thread_1', sourceRunId: 'run_1', sourceType: 'run', includeTombstoned: true, limit: 10 }}
        entries={[{
          id: 'mem_1',
          title: 'Preference',
          summary: 'Safe summary only',
          scopeType: 'thread',
          scopeId: 'thread_1',
          status: 'approved',
          safetyState: 'redacted',
          sourceThreadId: 'thread_1',
          sourceRunId: 'run_1',
          sourceType: 'run',
          createdAt: '2026-05-25T00:00:00Z',
          updatedAt: '2026-05-25T00:00:01Z',
          redactionApplied: true,
        }]}
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(html).toContain('Scope type')
    expect(html).toContain('Scope id')
    expect(html).toContain('Source thread')
    expect(html).toContain('Source run')
    expect(html).toContain('Source type')
    expect(html).toContain('Include deleted')
    expect(html).toContain('Limit')
    expect(html).toContain('Redacted')
    expect(html).not.toMatch(/raw content|Authorization|provider trace|tool output|\/Users\//i)
  })

  test('renders detail drawer and delete confirmation before destructive action', () => {
    const entry = {
      id: 'mem_1',
      title: 'Preference',
      summary: 'Safe detail summary',
      scopeType: 'thread' as const,
      scopeId: 'thread_1',
      status: 'tombstoned' as const,
      safetyState: 'redacted' as const,
      sourceThreadId: 'thread_1',
      sourceRunId: 'run_1',
      sourceType: 'run' as const,
      createdAt: '2026-05-25T00:00:00Z',
      updatedAt: '2026-05-25T00:00:01Z',
      deletedAt: '2026-05-25T00:00:02Z',
      redactionApplied: true,
    }
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        entries={[entry]}
        detailEntry={entry}
        pendingDeleteEntry={entry}
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onCloseDetail={() => {}}
        onRequestDelete={() => {}}
        onCancelDelete={() => {}}
        onConfirmDelete={() => {}}
      />,
    )

    expect(html).toContain('Memory detail')
    expect(html).toContain('Deleted')
    expect(html).toContain('Confirm deletion')
    expect(html).toContain('Preference')
    expect(html).toContain('thread · thread_1')
    expect(html).toContain('run · thread_1 · run_1')
    expect(html).toContain('Cancel')
    expect(html).toContain('Delete memory')
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
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(html).toContain('Memory failed to load')
    expect(html).toContain('Memory could not be loaded')
    expect(html).not.toContain('Should not masquerade as success')
  })

  test('renders real audit history event metadata', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        entries={[]}
        auditItems={[{
          id: 'audit_1',
          eventType: 'memory_write_approved',
          threadId: 'thread_1',
          runId: 'run_terminal',
          memoryEntryId: 'mem_1',
          memoryProposalId: 'memprop_1',
          status: 'approved',
          scopeType: 'thread',
          sourceType: 'run',
          redactionApplied: true,
          occurredAt: '2026-05-25T00:00:02Z',
          summary: 'unsafe /Users/xuean/.env provider trace stdout',
        }]}
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(html).toContain('Memory history')
    expect(html).toContain('memory_write_approved')
    expect(html).toContain('run_terminal')
    expect(html).toContain('memprop_1')
    expect(html).toContain('redacted')
    expect(html).not.toContain('/Users/xuean')
    expect(html).not.toContain('provider trace')
    expect(html).not.toContain('stdout')
  })

  test('does not fabricate audit history when backend is unavailable', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        entries={[]}
        auditItems={[]}
        auditError="Memory history endpoint unavailable"
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(html).toContain('Memory history endpoint unavailable')
    expect(html).toContain('Memory history could not be loaded')
    expect(html).not.toContain('memory_write_approved')
    expect(html).not.toContain('memory_snapshot_loaded')
  })

  test('renders loading and empty states without implying success', () => {
    const loadingHtml = renderToStaticMarkup(
      <MemoryPanel
        query=""
        loading
        entries={[]}
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )
    const emptyHtml = renderToStaticMarkup(
      <MemoryPanel
        query=""
        entries={[]}
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(loadingHtml).toContain('Loading')
    expect(emptyHtml).toContain('No memory entries')
  })
})
