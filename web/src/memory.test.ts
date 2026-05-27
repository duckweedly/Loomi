import { describe, expect, test } from 'bun:test'
import { mapApiMemoryEntry, mapApiMemoryErrorEvent, mapApiMemoryImpressionSnapshot, mapApiMemoryOverviewSnapshot, mapApiMemoryProviderStatus, mapApiMemoryWriteProposal } from './realApiClient'

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

  test('maps provider foundation status without leaking provider setup', () => {
    const status = mapApiMemoryProviderStatus({
      enabled: true,
      provider: 'semantic',
      label: 'Semantic',
      state: 'unconfigured',
      configured: false,
      commit_after_run: true,
      checked_at: '2026-05-26T12:00:00Z',
      openviking: {
        base_url: 'http://127.0.0.1:8282',
        root_api_key_set: true,
        embedding_model: 'text-embedding-3-large',
        embedding_api_key_set: true,
        vlm_model: 'gpt-5.5',
        vlm_api_key_set: true,
      },
      nowledge: {
        base_url: 'http://127.0.0.1:7727',
        api_key_set: true,
        request_timeout_ms: 30000,
      },
      diagnostic: { code: 'semantic_unconfigured', message: 'Semantic memory provider is not configured.' },
    })

    expect(status).toMatchObject({
      enabled: true,
      provider: 'semantic',
      label: 'Semantic',
      state: 'unconfigured',
      configured: false,
      commitAfterRun: true,
      checkedAt: '2026-05-26T12:00:00Z',
      openviking: {
        baseUrl: 'http://127.0.0.1:8282',
        rootApiKeySet: true,
        embeddingModel: 'text-embedding-3-large',
        embeddingApiKeySet: true,
        vlmModel: 'gpt-5.5',
        vlmApiKeySet: true,
      },
      nowledge: {
        baseUrl: 'http://127.0.0.1:7727',
        apiKeySet: true,
        requestTimeoutMs: 30000,
      },
      diagnostic: { code: 'semantic_unconfigured', message: 'Semantic memory provider is not configured.' },
    })
    expect(JSON.stringify(status)).not.toContain('sk-')
    expect(JSON.stringify(status)).not.toContain('Authorization')
  })

  test('maps pending write proposals without raw content', () => {
    const proposal = mapApiMemoryWriteProposal({
      id: 'memprop_1',
      title: 'Run outcome',
      summary: 'Assistant outcome summary',
      scope_type: 'thread',
      scope_id: 'thr_1',
      status: 'pending',
      safety_state: 'safe',
      source_thread_id: 'thr_1',
      source_run_id: 'run_1',
      created_at: '2026-05-26T12:00:00Z',
    })

    expect(proposal).toEqual({
      id: 'memprop_1',
      title: 'Run outcome',
      summary: 'Assistant outcome summary',
      scopeType: 'thread',
      scopeId: 'thr_1',
      status: 'pending',
      safetyState: 'safe',
      sourceThreadId: 'thr_1',
      sourceRunId: 'run_1',
      createdAt: '2026-05-26T12:00:00Z',
      redactionApplied: false,
    })
    expect(JSON.stringify(proposal)).not.toContain('content')
  })

  test('maps memory snapshot and impression without raw content fields', () => {
    const overview = mapApiMemoryOverviewSnapshot({
      memory_block: '- Preference: Prefers short answers',
      hits: [{
        uri: 'memory://mem_1',
        entry_id: 'mem_1',
        title: 'Preference',
        abstract: 'Prefers short answers',
        is_leaf: true,
        updated_at: '2026-05-27T12:00:00Z',
      }],
      updated_at: '2026-05-27T12:00:01Z',
      rebuilt: true,
    })
    const impression = mapApiMemoryImpressionSnapshot({
      impression: 'Preference: Prefers short answers',
      updated_at: '2026-05-27T12:00:02Z',
      rebuilt: true,
    })

    expect(overview).toMatchObject({
      memoryBlock: '- Preference: Prefers short answers',
      hits: [{ uri: 'memory://mem_1', entryId: 'mem_1', title: 'Preference', abstract: 'Prefers short answers', isLeaf: true }],
      rebuilt: true,
    })
    expect(impression).toEqual({
      impression: 'Preference: Prefers short answers',
      updatedAt: '2026-05-27T12:00:02Z',
      rebuilt: true,
    })
    expect(JSON.stringify({ overview, impression })).not.toContain('content')
  })

  test('maps memory errors without provider secrets', () => {
    const error = mapApiMemoryErrorEvent({
      code: 'nowledge_unconfigured',
      message: 'Provider is not configured.',
      provider: 'nowledge',
      state: 'unconfigured',
      checked_at: '2026-05-27T12:00:00Z',
      run_id: 'run_1',
      event_type: 'memory_external_snapshot_failed',
    })

    expect(error).toEqual({
      code: 'nowledge_unconfigured',
      message: 'Provider is not configured.',
      provider: 'nowledge',
      state: 'unconfigured',
      checkedAt: '2026-05-27T12:00:00Z',
      runId: 'run_1',
      eventType: 'memory_external_snapshot_failed',
    })
    expect(JSON.stringify(error)).not.toContain('api_key')
  })
})
