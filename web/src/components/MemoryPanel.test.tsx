import { describe, expect, test } from 'bun:test'
import { renderToStaticMarkup } from 'react-dom/server'
import { MemoryPanel } from './MemoryPanel'

describe('MemoryPanel', () => {
  test('renders safe summaries and delete controls', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        locale="zh"
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

    expect(html).toContain('已保存记忆')
    expect(html).toContain('Prefers safe snapshots')
    expect(html).toContain('Delete Preference')
  })

  test('keeps grounded filters collapsed until useful and renders safe metadata only', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query="preference"
        locale="zh"
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

    expect(html).toContain('范围')
    expect(html).toContain('范围 ID')
    expect(html).toContain('来源会话')
    expect(html).toContain('来源运行')
    expect(html).toContain('来源类型')
    expect(html).toContain('包含已删除')
    expect(html).toContain('数量')
    expect(html).toContain('已脱敏')
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
        locale="en"
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
        locale="en"
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

  test('renders real audit history as product copy without raw ids', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        locale="zh"
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

    expect(html).toContain('记忆历史')
    expect(html).toContain('记忆已保存')
    expect(html).toContain('已脱敏')
    expect(html).not.toContain('memory_write_approved')
    expect(html).not.toContain('run_terminal')
    expect(html).not.toContain('memprop_1')
    expect(html).not.toContain('/Users/xuean')
    expect(html).not.toContain('provider trace')
    expect(html).not.toContain('stdout')
  })

  test('does not fabricate audit history when backend is unavailable', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        locale="en"
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
        locale="en"
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
        locale="en"
        entries={[]}
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(loadingHtml).toContain('Loading')
    expect(emptyHtml).toContain('No saved memories')
  })

  test('folds system snapshot audit events by default', () => {
    const html = renderToStaticMarkup(
      <MemoryPanel
        query=""
        locale="zh"
        entries={[]}
        auditItems={[{
          id: 'audit_snapshot',
          eventType: 'memory_snapshot_loaded',
          threadId: 'thread_1',
          runId: 'run_1',
          status: 'empty',
          sourceType: 'run',
          redactionApplied: true,
          occurredAt: '2026-05-25T00:00:02Z',
          summary: 'snapshot loaded',
        }]}
        onQueryChange={() => {}}
        onFiltersChange={() => {}}
        onOpenDetail={() => {}}
        onRequestDelete={() => {}}
      />,
    )

    expect(html).toContain('运行时读取了记忆快照')
    expect(html).toContain('系统快照事件已折叠')
    expect(html).not.toContain('memory_snapshot_loaded')
  })
})
