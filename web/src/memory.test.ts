import { describe, expect, test } from 'bun:test'
import { mapApiMemoryEntry } from './realApiClient'

describe('M13 memory API mapping', () => {
  test('maps safe memory entry summaries without exposing raw content', () => {
    const entry = mapApiMemoryEntry({
      id: 'mem_1',
      title: 'Preference',
      summary: 'Prefers short slices',
      scope_type: 'user',
      source_thread_id: 'thr_1',
      source_run_id: 'run_1',
      created_at: '2026-05-25T00:00:00Z',
      updated_at: '2026-05-25T00:00:01Z',
      redaction_applied: true,
    })

    expect(entry).toEqual({
      id: 'mem_1',
      title: 'Preference',
      summary: 'Prefers short slices',
      scopeType: 'user',
      status: 'approved',
      sourceThreadId: 'thr_1',
      sourceRunId: 'run_1',
      createdAt: '2026-05-25T00:00:00Z',
      updatedAt: '2026-05-25T00:00:01Z',
      redactionApplied: true,
    })
  })
})
