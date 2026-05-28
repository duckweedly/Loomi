import { describe, expect, test } from 'bun:test'
import { createClientMessageID, getApiBaseUrl, getApiBearerToken, mapApiLocalProviderDetection, mapApiMemoryAuditItem, mapApiMemoryEntry, mapApiProviderCapability, mapApiRun, mapApiRunEvent, mapApiToolCall, mapApiWorkerQueueDiagnostics, realApiClient, selectSendProvider } from './realApiClient'

async function withLocalStorage<T>(run: () => T | Promise<T>): Promise<T> {
  if (globalThis.localStorage) return run()
  const values = new Map<string, string>()
  const storage = {
    getItem: (key: string) => values.get(key) ?? null,
    setItem: (key: string, value: string) => { values.set(key, value) },
    removeItem: (key: string) => { values.delete(key) },
    clear: () => { values.clear() },
    key: (index: number) => Array.from(values.keys())[index] ?? null,
    get length() { return values.size },
  } as Storage
  Object.defineProperty(globalThis, 'localStorage', { value: storage, configurable: true })
  try {
    return await run()
  } finally {
    Reflect.deleteProperty(globalThis, 'localStorage')
  }
}

describe('createClientMessageID', () => {
  test('does not rely on Date.now alone', () => {
    const originalNow = Date.now
    Date.now = () => 123
    try {
      const first = createClientMessageID()
      const second = createClientMessageID()

      expect(first).toStartWith('web-123-')
      expect(second).toStartWith('web-123-')
      expect(second).not.toBe(first)
    } finally {
      Date.now = originalNow
    }
  })
})

describe('real API response validation', () => {
  test('reports the active API base for desktop diagnostics', () => {
    expect(getApiBaseUrl()).toBeTruthy()
  })

  test('reads a browser-local API token without exposing it in output objects', async () => {
    await withLocalStorage(() => {
      const previous = globalThis.localStorage?.getItem('loomi.api_token')
      globalThis.localStorage?.setItem('loomi.api_token', 'browser-secret-token')
      try {
        expect(getApiBearerToken()).toBe('browser-secret-token')
      } finally {
        if (previous === null || previous === undefined) globalThis.localStorage?.removeItem('loomi.api_token')
        else globalThis.localStorage?.setItem('loomi.api_token', previous)
      }
    })
  })

  test('sends bearer token on JSON requests when configured in browser storage', async () => {
    await withLocalStorage(async () => {
      const previous = globalThis.localStorage?.getItem('loomi.api_token')
      const originalFetch = globalThis.fetch
      let authorization = ''
      globalThis.localStorage?.setItem('loomi.api_token', 'browser-secret-token')
      globalThis.fetch = (async (_input: RequestInfo | URL, init?: RequestInit) => {
        authorization = new Headers(init?.headers).get('Authorization') ?? ''
        return new Response(JSON.stringify({ threads: [] }), { status: 200, headers: { 'Content-Type': 'application/json' } })
      }) as typeof fetch
      try {
        await realApiClient.listThreads()
      } finally {
        globalThis.fetch = originalFetch
        if (previous === null || previous === undefined) globalThis.localStorage?.removeItem('loomi.api_token')
        else globalThis.localStorage?.setItem('loomi.api_token', previous)
      }

      expect(authorization).toBe('Bearer browser-secret-token')
    })
  })

  test('converts raw fetch failures into a Loomi API diagnostic', async () => {
    const originalFetch = globalThis.fetch
    globalThis.fetch = (async () => {
      throw new TypeError('Failed to fetch')
    }) as typeof fetch
    try {
      let error: unknown
      try {
        await realApiClient.listThreads()
      } catch (err) {
        error = err
      }

      expect(error).toBeInstanceOf(Error)
      expect((error as { code?: string }).code).toBe('api_unreachable')
      expect((error as Error).message).toContain('Loomi API 未连接')
      expect((error as Error).message).not.toContain('Failed to fetch')
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('reads not-ready schema state from /readyz for desktop readiness', async () => {
    const originalFetch = globalThis.fetch
    globalThis.fetch = (async () => new Response(JSON.stringify({
      status: 'not_ready',
      checks: [{ name: 'schema', status: 'failed', message: 'migration pending' }],
    }), { status: 503, headers: { 'Content-Type': 'application/json' } })) as typeof fetch
    try {
      const readiness = await realApiClient.getReadiness?.()

      expect(readiness?.status).toBe('not_ready')
      expect(readiness?.checks[0]).toMatchObject({ name: 'schema', status: 'failed' })
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('maps assistant message run id from safe metadata for terminal reconciliation', async () => {
    const originalFetch = globalThis.fetch
    globalThis.fetch = (async () => new Response(JSON.stringify({
      messages: [{
        id: 'msg-assistant',
        thread_id: 'thread-1',
        role: 'assistant',
        content: 'Persisted final',
        metadata: { run_id: 'run-1' },
        created_at: '2026-05-27T00:00:00Z',
      }],
    }), { status: 200, headers: { 'Content-Type': 'application/json' } })) as typeof fetch
    try {
      const messages = await realApiClient.getThreadMessages('thread-1')

      expect(messages[0]).toMatchObject({ id: 'msg-assistant', runId: 'run-1', content: 'Persisted final' })
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('desktop dev defaults to the local API port when no env is supplied', async () => {
    const source = await Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()

    expect(source).toContain("import.meta.env.DEV ? 'http://127.0.0.1:18080'")
    expect(source).toContain('configuredApiBaseUrl || devApiBaseUrl')
  })

  test('listThreads reports invalid API bodies without a raw null property crash', async () => {
    const originalFetch = globalThis.fetch
    globalThis.fetch = (async () => new Response('null', { status: 200, headers: { 'Content-Type': 'application/json' } })) as typeof fetch
    try {
      let error: unknown
      try {
        await realApiClient.listThreads()
      } catch (err) {
        error = err
      }

      expect(error).toBeInstanceOf(Error)
      expect((error as { code?: string }).code).toBe('invalid_response')
      expect((error as Error).message).toBe('Thread list response was invalid.')
      expect((error as Error).message).not.toContain('Cannot read')
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('listThreads accepts an empty array after every thread is archived', async () => {
    const originalFetch = globalThis.fetch
    globalThis.fetch = (async () => new Response(JSON.stringify({ threads: [] }), { status: 200, headers: { 'Content-Type': 'application/json' } })) as typeof fetch
    try {
      const threads = await realApiClient.listThreads()

      expect(threads).toEqual([])
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('listModelProviders reports invalid API bodies without pretending providers are missing', async () => {
    const originalFetch = globalThis.fetch
    globalThis.fetch = (async () => new Response('{}', { status: 200, headers: { 'Content-Type': 'application/json' } })) as typeof fetch
    try {
      let error: unknown
      try {
        await realApiClient.listModelProviders?.()
      } catch (err) {
        error = err
      }

      expect(error).toBeInstanceOf(Error)
      expect((error as { code?: string }).code).toBe('invalid_response')
      expect((error as Error).message).toBe('Provider list response was invalid.')
      expect((error as Error).message).not.toContain('providers')
    } finally {
      globalThis.fetch = originalFetch
    }
  })

  test('MCP config writes use safe management endpoints', async () => {
    const originalFetch = globalThis.fetch
    const requests: Array<{ url: string; init?: RequestInit }> = []
    globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
      requests.push({ url: String(input), init })
      if (String(input).endsWith('/discover')) {
        return new Response(JSON.stringify({ server: apiMCPServer('succeeded') }), { status: 200, headers: { 'Content-Type': 'application/json' } })
      }
      if (init?.method === 'DELETE') {
        return new Response(JSON.stringify({ servers: [] }), { status: 200, headers: { 'Content-Type': 'application/json' } })
      }
      return new Response(JSON.stringify({ server: apiMCPServer('not_discovered') }), { status: 200, headers: { 'Content-Type': 'application/json' } })
    }) as typeof fetch
    try {
      await realApiClient.saveMCPServer?.({ slug: 'local-smoke', displayName: 'Local Smoke', enabled: true, transport: 'stdio', command: '/bin/mcp', args: ['--safe'], env: { MODE: 'test' }, timeoutMs: 5000 })
      await realApiClient.discoverMCPServer?.('local-smoke')
      await realApiClient.deleteMCPServer?.('local-smoke')
    } finally {
      globalThis.fetch = originalFetch
    }

    expect(requests.map((request) => request.url)).toEqual(['/v1/mcp/servers', '/v1/mcp/servers/local-smoke/discover', '/v1/mcp/servers/local-smoke'])
    expect(requests[0].init?.method).toBe('POST')
    expect(String(requests[0].init?.body)).toContain('"display_name":"Local Smoke"')
    expect(String(requests[0].init?.body)).toContain('"timeout_ms":5000')
    expect(requests[1].init?.method).toBe('POST')
    expect(requests[2].init?.method).toBe('DELETE')
  })
})

function apiMCPServer(discoveryStatus: string) {
  return {
    server_safe_id: 'mcp:local-smoke',
    server_slug: 'local-smoke',
    display_name: 'Local Smoke',
    transport: 'stdio',
    enabled: true,
    config_source: 'local',
    discovery_status: discoveryStatus,
    candidate_count: discoveryStatus === 'succeeded' ? 1 : 0,
    candidate_names: discoveryStatus === 'succeeded' ? ['mcp.local-smoke.echo'] : [],
    execution_mode: discoveryStatus === 'succeeded' ? 'approval_gated' : 'disabled',
  }
}

describe('M14 memory management mapping', () => {
  test('maps safe memory management fields without raw content', () => {
    const entry = mapApiMemoryEntry({
      id: 'mem_1',
      title: 'Preference',
      summary: 'Prefers short replies',
      scope_type: 'thread',
      scope_id: 'thr_1',
      status: 'tombstoned',
      safety_state: 'redacted',
      source_thread_id: 'thr_1',
      source_run_id: 'run_1',
      source_event_id: 'evt_1',
      source_type: 'run',
      created_at: '2026-05-25T00:00:00Z',
      updated_at: '2026-05-25T00:01:00Z',
      deleted_at: '2026-05-25T00:02:00Z',
      redaction_applied: true,
    })

    expect(entry).toMatchObject({
      id: 'mem_1',
      scopeType: 'thread',
      scopeId: 'thr_1',
      status: 'tombstoned',
      safetyState: 'redacted',
      sourceRunId: 'run_1',
      sourceType: 'run',
      deletedAt: '2026-05-25T00:02:00Z',
      redactionApplied: true,
    })
    expect(JSON.stringify(entry)).not.toContain('content')
  })

  test('maps safe audit events without provider or tool payloads', () => {
    const item = mapApiMemoryAuditItem({
      id: 'evt_1',
      event_type: 'memory_write_approved',
      summary: 'Memory write approved',
      thread_id: 'thr_1',
      run_id: 'run_1',
      memory_entry_id: 'mem_1',
      memory_proposal_id: 'memprop_1',
      status: 'approved',
      scope_type: 'thread',
      source_type: 'run',
      redaction_applied: true,
      occurred_at: '2026-05-25T00:00:00Z',
    })

    expect(item).toMatchObject({
      eventType: 'memory_write_approved',
      memoryEntryId: 'mem_1',
      memoryProposalId: 'memprop_1',
      redactionApplied: true,
    })
    expect(JSON.stringify(item)).not.toContain('provider')
    expect(JSON.stringify(item)).not.toContain('/Users/')
  })

  test('maps terminal-run memory audit events for UI history', () => {
    const item = mapApiMemoryAuditItem({
      id: 'evt_terminal',
      event_type: 'memory_snapshot_loaded',
      summary: 'Snapshot loaded after terminal run',
      thread_id: 'thread_1',
      run_id: 'run_completed',
      status: 'loaded',
      scope_type: 'thread',
      source_type: 'run',
      redaction_applied: true,
      occurred_at: '2026-05-25T00:00:00Z',
    })

    expect(item).toMatchObject({
      eventType: 'memory_snapshot_loaded',
      runId: 'run_completed',
      status: 'loaded',
      redactionApplied: true,
    })
  })

  test('uses grounded snake_case memory scope filters', async () => {
    const source = await Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()

    expect(source).toContain("params.set('source_thread_id'")
    expect(source).toContain('source_thread_id:')
    expect(source).not.toContain('workspace_id')
  })

  test('delete request body omits UI-only search fields', async () => {
    const source = await Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()
    const deleteBody = source.slice(source.indexOf('function memoryDeleteRequestBody'), source.indexOf('export function mapApiRunEvent'))

    expect(deleteBody).toContain('scope_type')
    expect(deleteBody).toContain('source_run_id')
    expect(deleteBody).not.toContain('limit')
    expect(deleteBody).not.toContain('include_tombstoned')
  })

  test('audit request uses the unified memory filter query contract', async () => {
    const originalFetch = globalThis.fetch
    const requested: string[] = []
    globalThis.fetch = (async (input: RequestInfo | URL) => {
      requested.push(String(input))
      return new Response(JSON.stringify({ items: [] }), { status: 200, headers: { 'Content-Type': 'application/json' } })
    }) as typeof fetch
    try {
      await realApiClient.listMemoryAudit?.({ scopeType: 'thread', scopeId: 'thread_1', sourceThreadId: 'thread_1', sourceRunId: 'run_1', sourceType: 'run', includeTombstoned: true, limit: 9 })
    } finally {
      globalThis.fetch = originalFetch
    }

    const url = new URL(requested[0], 'http://loomi.local')
    expect(url.pathname).toBe('/v1/memory/audit')
    expect(url.searchParams.get('scope_type')).toBe('thread')
    expect(url.searchParams.get('scope_id')).toBe('thread_1')
    expect(url.searchParams.get('source_thread_id')).toBe('thread_1')
    expect(url.searchParams.get('source_run_id')).toBe('run_1')
    expect(url.searchParams.get('source_type')).toBe('run')
    expect(url.searchParams.get('limit')).toBe('9')
    expect(url.searchParams.has('thread_id')).toBe(false)
  })
})

describe('M18.5 local provider detection mapping', () => {
  test('maps safe local provider detection fields without secrets', () => {
    const detection = mapApiLocalProviderDetection({
      provider_id: 'local_codex',
      display_name: 'Local Codex',
      provider_kind: 'codex',
      auth_mode: 'oauth',
      status: 'available',
      model_candidates: ['gpt-5.5'],
      source: 'local_config',
      redaction_applied: true,
      message: 'Detected but not enabled. Explicit opt-in is required before use.',
    })

    expect(detection).toEqual({
      providerId: 'local_codex',
      displayName: 'Local Codex',
      providerKind: 'codex',
      authMode: 'oauth',
      status: 'available',
      modelCandidates: ['gpt-5.5'],
      source: 'local_config',
      redactionApplied: true,
      message: 'Detected but not enabled. Explicit opt-in is required before use.',
    })
    expect(JSON.stringify(detection)).not.toContain('access_token')
    expect(JSON.stringify(detection)).not.toContain('sk-')
  })

  test('maps enabled local provider capability without secrets or paths', () => {
    const provider = mapApiProviderCapability({
      id: 'local_codex',
      family: 'openai_compatible',
      model: 'gpt-5.5',
      status: 'available',
      message: 'Local Codex enabled.',
      local_provider: true,
      session_local: true,
      credential_reference: 'redacted',
      execution_state: 'supported',
    })

    expect(provider).toMatchObject({
      id: 'local_codex',
      localProvider: true,
      sessionLocal: true,
      credentialReference: 'redacted',
      executionState: 'supported',
      status: 'available',
    })
    expect(JSON.stringify(provider)).not.toContain('access_token')
    expect(JSON.stringify(provider)).not.toContain('/Users/')
    expect(JSON.stringify(provider)).not.toContain('sk-')
  })

  test('calls the dedicated local provider detection endpoint', async () => {
    const originalFetch = globalThis.fetch
    const requested: string[] = []
    globalThis.fetch = (async (input: RequestInfo | URL) => {
      requested.push(String(input))
      return new Response(JSON.stringify({ providers: [], request_id: 'req_local' }), { status: 200, headers: { 'Content-Type': 'application/json' } })
    }) as typeof fetch
    try {
      await realApiClient.listLocalProviderDetections?.()
    } finally {
      globalThis.fetch = originalFetch
    }

    const url = new URL(requested[0], 'http://loomi.local')
    expect(url.pathname).toBe('/v1/local-provider-detections')
  })

  test('calls dedicated memory provider detection endpoints', async () => {
    const originalFetch = globalThis.fetch
    const requested: string[] = []
    globalThis.fetch = (async (input: RequestInfo | URL) => {
      requested.push(String(input))
      return new Response(JSON.stringify({ detected: true, base_url: 'http://127.0.0.1:8282', message: 'detected', request_id: 'req_memory_detect' }), { status: 200, headers: { 'Content-Type': 'application/json' } })
    }) as typeof fetch
    try {
      await realApiClient.detectNowledgeMemoryProvider?.()
      await realApiClient.detectOpenVikingMemoryProvider?.()
    } finally {
      globalThis.fetch = originalFetch
    }

    expect(new URL(requested[0], 'http://loomi.local').pathname).toBe('/v1/memory/provider/nowledge/detect')
    expect(new URL(requested[1], 'http://loomi.local').pathname).toBe('/v1/memory/provider/openviking/detect')
  })

  test('calls explicit local provider enable and disable endpoints', async () => {
    const originalFetch = globalThis.fetch
    const requested: { url: string; method: string }[] = []
    globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
      requested.push({ url: String(input), method: init?.method ?? 'GET' })
      return new Response(JSON.stringify({
        provider: {
          id: 'local_codex',
          family: 'openai_compatible',
          model: 'gpt-5.5',
          status: 'unavailable',
          local_provider: true,
          session_local: true,
          credential_reference: 'redacted',
          execution_state: 'unsupported',
        },
        request_id: 'req_local_enable',
      }), { status: 200, headers: { 'Content-Type': 'application/json' } })
    }) as typeof fetch
    try {
      await realApiClient.enableLocalProvider?.('local_codex')
      await realApiClient.disableLocalProvider?.('local_codex')
    } finally {
      globalThis.fetch = originalFetch
    }

    expect(new URL(requested[0].url, 'http://loomi.local').pathname).toBe('/v1/local-provider-detections/local_codex/enable')
    expect(requested[0].method).toBe('POST')
    expect(new URL(requested[1].url, 'http://loomi.local').pathname).toBe('/v1/local-provider-detections/local_codex/enable')
    expect(requested[1].method).toBe('DELETE')
  })

  test('calls workspace root endpoints for desktop folder authorization', async () => {
    const originalFetch = globalThis.fetch
    const requested: { url: string; method: string; body?: string }[] = []
    globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
      requested.push({ url: String(input), method: init?.method ?? 'GET', body: init?.body?.toString() })
      return new Response(JSON.stringify({
        config: { configured: true, display_name: 'Downloads' },
        request_id: 'req_workspace_root',
      }), { status: 200, headers: { 'Content-Type': 'application/json' } })
    }) as typeof fetch
    try {
      const current = await realApiClient.getWorkspaceRoot?.()
      const saved = await realApiClient.saveWorkspaceRoot?.({ path: '/Users/xuean/Downloads' })
      expect(current?.displayName).toBe('Downloads')
      expect(saved?.configured).toBe(true)
    } finally {
      globalThis.fetch = originalFetch
    }

    expect(new URL(requested[0].url, 'http://loomi.local').pathname).toBe('/v1/workspace/root')
    expect(requested[0].method).toBe('GET')
    expect(new URL(requested[1].url, 'http://loomi.local').pathname).toBe('/v1/workspace/root')
    expect(requested[1].method).toBe('POST')
    expect(requested[1].body).toContain('/Users/xuean/Downloads')
  })
})

describe('M7 tool-call mapping', () => {
  test('maps scoped tool-call projection without raw provider payload fields', () => {
    const call = mapApiToolCall({
      id: 'tool_1',
      thread_id: 'thread_1',
      run_id: 'run_1',
      tool_call_id: 'tc_1',
      tool_name: 'runtime.get_current_time',
      arguments_summary: { timezone: 'UTC' },
      approval_status: 'required',
      execution_status: 'blocked',
      result_summary: null,
      error_code: null,
      error_message: null,
    })

    expect(call).toEqual({
      id: 'tool_1',
      toolCallId: 'tc_1',
      name: 'runtime.get_current_time',
      status: 'approval_required',
      approvalStatus: 'required',
      executionStatus: 'blocked',
      summary: 'Approval required',
      input: '{"timezone":"UTC"}',
      output: '',
      argumentsSummary: { timezone: 'UTC' },
      resultSummary: null,
      errorCode: null,
      errorMessage: null,
    })
  })

  test('maps approved denied and executing tool-call projections', () => {
    expect(mapApiToolCall({
      id: 'tool_approved',
      thread_id: 'thread_1',
      run_id: 'run_1',
      tool_call_id: 'tc_1',
      tool_name: 'runtime.get_current_time',
      arguments_summary: { timezone: 'UTC' },
      approval_status: 'approved',
      execution_status: 'not_started',
      result_summary: null,
      error_code: null,
      error_message: null,
    })).toMatchObject({ status: 'approved', approvalStatus: 'approved', executionStatus: 'not_started', summary: 'Approved' })

    expect(mapApiToolCall({
      id: 'tool_denied',
      thread_id: 'thread_1',
      run_id: 'run_1',
      tool_call_id: 'tc_1',
      tool_name: 'runtime.get_current_time',
      arguments_summary: { timezone: 'UTC' },
      approval_status: 'denied',
      execution_status: 'cancelled',
      result_summary: null,
      error_code: null,
      error_message: null,
    })).toMatchObject({ status: 'denied', approvalStatus: 'denied', summary: 'Denied' })

    expect(mapApiToolCall({
      id: 'tool_executing',
      thread_id: 'thread_1',
      run_id: 'run_1',
      tool_call_id: 'tc_1',
      tool_name: 'runtime.get_current_time',
      arguments_summary: { timezone: 'UTC' },
      approval_status: 'approved',
      execution_status: 'executing',
      result_summary: null,
      error_code: null,
      error_message: null,
    })).toMatchObject({ status: 'executing', executionStatus: 'executing', summary: 'Executing' })
  })

  test('exposes approve and deny API paths', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()
    return expect(source).resolves.toContain('/tool-calls/${toolCallId}/approve')
  })
})

describe('M11 MCP event mapping', () => {
  test('maps MCP discovery events to worker job or error groups', () => {
    const succeeded = mapApiRunEvent({
      id: 'evt-mcp-ok',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 1,
      category: 'progress',
      type: 'mcp_discovery_succeeded',
      summary: 'MCP discovery succeeded',
      content: null,
      metadata: { mcp_candidate_count: 1, mcp_execution_enabled: false },
      created_at: '2026-05-25T00:00:00Z',
    })
    const failed = mapApiRunEvent({
      id: 'evt-mcp-failed',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 2,
      category: 'error',
      type: 'mcp_discovery_failed',
      summary: 'MCP discovery failed',
      content: null,
      metadata: { error_code: 'mcp_discovery_timeout' },
      created_at: '2026-05-25T00:00:01Z',
    })

    expect(succeeded.type).toBe('mcp.discovery.succeeded')
    expect(succeeded.group).toBe('worker-job')
    expect(failed.type).toBe('mcp.discovery.failed')
    expect(failed.group).toBe('error')
  })
})

describe('M6 worker queue diagnostics mapping', () => {
  test('maps worker queue diagnostics without credential fields', () => {
    const diagnostics = mapApiWorkerQueueDiagnostics({
      queue_status: 'degraded',
      worker_status: 'degraded',
      queued_count: 1,
      leased_count: 2,
      stale_count: 1,
      retrying_count: 1,
      dead_count: 1,
      blocked_tool_approval_count: 3,
      resumable_tool_call_count: 2,
      updated_at: '2026-05-24T10:00:00Z',
    })

    expect(diagnostics).toEqual({
      queueStatus: 'degraded',
      workerStatus: 'degraded',
      queuedCount: 1,
      leasedCount: 2,
      staleCount: 1,
      retryingCount: 1,
      deadCount: 1,
      blockedToolApprovalCount: 3,
      resumableToolCallCount: 2,
      updatedAt: '2026-05-24T10:00:00Z',
    })
  })
})

describe('M5 provider and run mapping', () => {
  test('maps model gateway run source', () => {
    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'model_gateway',
      title: 'Model gateway run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    })

    expect(run.source).toBe('model_gateway')
    expect(run.model).toBe('Model gateway')
    expect(run.context).toBe('model_gateway')
  })

  test('maps provider capability without credential fields', () => {
    const provider = mapApiProviderCapability({ id: 'custom', family: 'openai_compatible', base_url: 'https://example.test/v1', model: 'gpt-5.5', status: 'available' })

    expect(provider.id).toBe('custom')
    expect(provider.family).toBe('openai_compatible')
    expect(provider.baseUrl).toBe('https://example.test/v1')
    expect(provider.model).toBe('gpt-5.5')
  })

  test('maps provider completion diagnosis fields without raw provider body', () => {
    const provider = mapApiProviderCapability({
      id: 'custom',
      family: 'openai_compatible',
      model: 'gpt-5.5',
      status: 'completion-failed',
      check_code: 'completion-failed-503',
      check_stage: 'completion',
      http_status: 503,
      message: 'Provider completion check failed with HTTP 503.',
    })

    expect(provider.status).toBe('completion-failed')
    expect(provider.checkCode).toBe('completion-failed-503')
    expect(provider.checkStage).toBe('completion')
    expect(provider.httpStatus).toBe(503)
    expect(JSON.stringify(provider)).not.toContain('raw provider body')
    expect(JSON.stringify(provider)).not.toContain('sk-')
  })
})

describe('M4 run mapping', () => {
  test('maps local simulated run status without LLM/tool claims', () => {
    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    })

    expect(run.status).toBe('running')
    expect(run.model).toBe('Local simulated')
    expect(run.context).toBe('local_simulated')
    expect(run.assistantDraft).toMatchObject({ content: '', status: 'pending' })
  })

  test('maps model gateway source and recovering events', () => {
    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'model_gateway',
      title: 'Model gateway run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, [mapApiRunEvent({ id: 'evt-recovering', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'lifecycle', type: 'run.recovering', summary: 'Recovering', content: null, metadata: {}, created_at: '2026-05-23T00:00:00Z' })])

    expect(run.model).toBe('Model gateway')
    expect(run.context).toBe('model_gateway')
    expect(run.status).toBe('recovering')
    expect(run.assistantDraft).toMatchObject({ status: 'recovering' })
  })

  test('restores tool-call cards from loaded event history', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-tool-1', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'progress', type: 'tool_call_requested', summary: 'Tool call requested', content: null, metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'required', execution_status: 'blocked' }, created_at: '2026-05-25T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-tool-2', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'progress', type: 'tool_call_approval_required', summary: 'Tool approval required', content: null, metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' }, approval_status: 'required', execution_status: 'blocked' }, created_at: '2026-05-25T00:00:01Z' }),
    ]
    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'blocked_on_tool_approval',
      source: 'model_gateway',
      title: 'Model gateway run',
      created_at: '2026-05-25T00:00:00Z',
      updated_at: '2026-05-25T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, events)

    expect(run.toolCalls?.[0]).toMatchObject({ toolCallId: 'tc_1', status: 'approval_required', approvalStatus: 'required', executionStatus: 'blocked' })
  })

  test('restores assistant draft from loaded model event history', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-delta', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'message', type: 'model.delta', summary: 'Delta', content: 'Hel', metadata: {}, created_at: '2026-05-23T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-delta-2', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'message', type: 'model.delta', summary: 'Delta', content: 'lo', metadata: {}, created_at: '2026-05-23T00:00:01Z' }),
    ]

    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:01Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, events)

    expect(run.assistantDraft).toMatchObject({ content: 'Hello', status: 'streaming', lastEventId: 'evt-delta-2' })
  })

  test('does not restore late final events over terminal stopped history', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-delta', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'message', type: 'model.delta', summary: 'Delta', content: 'Partial', metadata: {}, created_at: '2026-05-23T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-stopped', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'lifecycle', type: 'run.stopped', summary: 'Stopped', content: null, metadata: {}, created_at: '2026-05-23T00:00:01Z' }),
      mapApiRunEvent({ id: 'evt-final', run_id: 'run-1', thread_id: 'thread-1', sequence: 3, category: 'final', type: 'model.final', summary: 'Final', content: 'Final', metadata: {}, created_at: '2026-05-23T00:00:02Z' }),
    ]

    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:02Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, events)

    expect(run.status).toBe('stopped')
    expect(run.assistantDraft).toMatchObject({ content: 'Partial', status: 'stopped', lastEventId: 'evt-stopped' })
  })

  test('keeps loaded event history for already terminal runs', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-context', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'progress', type: 'pipeline_step_completed', summary: 'Pipeline step completed', content: null, metadata: { step: 'prepare_context', persona_name: 'Default', persona_version: '2026-05-25.1' }, created_at: '2026-05-25T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-completed', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'final', type: 'run_completed', summary: 'Run completed', content: null, metadata: {}, created_at: '2026-05-25T00:00:01Z' }),
    ]

    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'completed',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-25T00:00:00Z',
      updated_at: '2026-05-25T00:00:01Z',
      completed_at: '2026-05-25T00:00:01Z',
      error_code: null,
      error_message: null,
    }, events)

    expect(run.status).toBe('completed')
    expect(run.events).toHaveLength(2)
    expect(run.events[0].detail).toContain('persona_version: 2026-05-25.1')
  })

  test('exposes subscribeRunEvents for EventSource-compatible streaming', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()
    return Promise.all([
      expect(source).resolves.toContain('subscribeRunEvents'),
      expect(source).resolves.toContain('stream_closed'),
      expect(source).resolves.toContain('onClosed?.()'),
    ])
  })

  test('streams run events through fetch so bearer auth works in desktop renderer', async () => {
    await withLocalStorage(async () => {
      const previous = globalThis.localStorage?.getItem('loomi.api_token')
      const originalFetch = globalThis.fetch
      const requests: Array<{ url: string; authorization: string }> = []
      globalThis.localStorage?.setItem('loomi.api_token', 'stream-secret-token')
      globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
        requests.push({ url: String(input), authorization: new Headers(init?.headers).get('Authorization') ?? '' })
        const payload = [
          'event: run_event',
          'data: {"event":{"id":"evt-1","run_id":"run-1","thread_id":"thread-1","sequence":1,"category":"final","type":"run_completed","summary":"Run completed","content":null,"metadata":{},"created_at":"2026-05-27T00:00:00Z"}}',
          '',
          'event: stream_closed',
          'data: {}',
          '',
        ].join('\n')
        return new Response(payload, { status: 200, headers: { 'Content-Type': 'text/event-stream' } })
      }) as typeof fetch
      try {
        const events: ReturnType<typeof mapApiRunEvent>[] = []
        let closed = false
        realApiClient.subscribeRunEvents?.('run-1', 7, (event) => events.push(event), () => {}, () => { closed = true })
        await new Promise((resolve) => setTimeout(resolve, 10))

        expect(requests[0].url).toContain('/v1/runs/run-1/events/stream?after_sequence=7')
        expect(requests[0].authorization).toBe('Bearer stream-secret-token')
        expect(events[0]).toMatchObject({ id: 'evt-1', status: 'completed' })
        expect(closed).toBe(true)
      } finally {
        globalThis.fetch = originalFetch
        if (previous === null || previous === undefined) globalThis.localStorage?.removeItem('loomi.api_token')
        else globalThis.localStorage?.setItem('loomi.api_token', previous)
      }
    })
  })

  test('requests replay events with explicit after_sequence cursor', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()
    return Promise.all([
      expect(source).resolves.toContain('getRunEvents(runId: string, afterSequence = 0)'),
      expect(source).resolves.toContain('after_sequence=${afterSequence}'),
    ])
  })

  test('maps model delta final and error events into assistant draft signals', () => {
    const delta = mapApiRunEvent({ id: 'evt-delta', run_id: 'run-1', thread_id: 'thread-1', sequence: 3, category: 'message', type: 'model.delta', summary: 'Delta', content: 'Hel', metadata: {}, created_at: '2026-05-23T00:00:00Z' })
    const final = mapApiRunEvent({ id: 'evt-final', run_id: 'run-1', thread_id: 'thread-1', sequence: 4, category: 'final', type: 'model.final', summary: 'Final', content: 'Hello', metadata: {}, created_at: '2026-05-23T00:00:01Z' })
    const error = mapApiRunEvent({ id: 'evt-error', run_id: 'run-1', thread_id: 'thread-1', sequence: 5, category: 'error', type: 'model.error', summary: 'Provider failed', content: null, metadata: {}, created_at: '2026-05-23T00:00:02Z' })

    expect(delta.type).toBe('model.delta')
    expect(delta.assistantDelta).toBe('Hel')
    expect(delta.status).toBe('running')
    expect(final.type).toBe('model.final')
    expect(final.content).toBe('Hello')
    expect(final.status).toBe('completed')
    expect(error.type).toBe('model.error')
    expect(error.status).toBe('failed')
  })

  test('restores assistant draft from current M4 local simulated event vocabulary', () => {
    const events = [
      mapApiRunEvent({ id: 'evt-drafting', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'progress', type: 'drafting', summary: 'Drafting response', content: null, metadata: {}, created_at: '2026-05-23T00:00:00Z' }),
      mapApiRunEvent({ id: 'evt-message', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'message', type: 'assistant_message', summary: 'Simulated response', content: 'Local simulated response', metadata: {}, created_at: '2026-05-23T00:00:01Z' }),
      mapApiRunEvent({ id: 'evt-final', run_id: 'run-1', thread_id: 'thread-1', sequence: 3, category: 'final', type: 'run_completed', summary: 'Run completed', content: null, metadata: {}, created_at: '2026-05-23T00:00:02Z' }),
    ]

    const run = mapApiRun({
      id: 'run-1',
      thread_id: 'thread-1',
      status: 'running',
      source: 'local_simulated',
      title: 'Local simulated run',
      created_at: '2026-05-23T00:00:00Z',
      updated_at: '2026-05-23T00:00:02Z',
      completed_at: null,
      error_code: null,
      error_message: null,
    }, events)

    expect(events.map((event) => event.type)).toEqual(['assistant.drafting', 'assistant.message.completed', 'run.completed'])
    expect(run.status).toBe('completed')
    expect(run.assistantDraft).toMatchObject({ content: 'Local simulated response', status: 'completed', lastEventId: 'evt-message' })
  })

  test('maps lifecycle progress message error and final event categories', () => {
    const event = mapApiRunEvent({
      id: 'evt-1',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 2,
      category: 'progress',
      type: 'context_loaded',
      summary: 'Context loaded',
      content: null,
      metadata: {},
      created_at: '2026-05-23T00:00:00Z',
    })

    expect(event.type).toBe('progress.context_loaded')
    expect(event.label).toBe('progress')
    expect(event.detail).toBe('Context loaded')
    expect(event.status).toBe('running')
  })

  test('maps model output deltas into assistantDelta for streaming drafts', () => {
    const event = mapApiRunEvent({
      id: 'evt-2',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 3,
      category: 'message',
      type: 'model_output_delta',
      summary: 'Model output delta',
      content: 'hello',
      metadata: {},
      created_at: '2026-05-23T00:00:01Z',
    })

    expect(event.type).toBe('message.model_output_delta')
    expect(event.assistantDelta).toBe('hello')
    expect(event.status).toBe('running')
  })

  test('maps model output completion as terminal assistant content', () => {
    const event = mapApiRunEvent({
      id: 'evt-final-message',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 4,
      category: 'message',
      type: 'model_output_completed',
      summary: 'Model output completed',
      content: '## Final\n\n- rendered',
      metadata: {},
      created_at: '2026-05-27T00:00:00Z',
    })

    expect(event.type).toBe('message.model_output_completed')
    expect(event.content).toBe('## Final\n\n- rendered')
    expect(event.status).toBe('completed')
    expect(event.assistantDelta).toBeUndefined()
  })

  test('real sendMessage starts model gateway runs from durable messages', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()

    return expect(source).resolves.toContain("source: 'model_gateway'")
  })

  test('real sendMessage checks provider and active run before creating a user message', async () => {
    const source = await Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()
    const providerCheck = source.indexOf('const provider = options?.providerId')
    const activeRunCheck = source.indexOf('const currentRun = await this.getThreadRun(threadId)')
    const messageCreate = source.indexOf("requestJSON<{ message: ApiMessage }>(`/v1/threads/${threadId}/messages`")

    expect(providerCheck).toBeGreaterThan(0)
    expect(activeRunCheck).toBeGreaterThan(providerCheck)
    expect(messageCreate).toBeGreaterThan(activeRunCheck)
    expect(source).toContain("model: options?.model || provider.model")
    expect(source).toContain('当前会话还有任务未结束，请先确认或停止当前任务。')
  })

  test('real sendMessage prefers supported local codex over saved custom provider', () => {
    const provider = selectSendProvider([
      { id: 'custom', family: 'openai_compatible', model: 'gpt-5.5', status: 'available' },
      { id: 'local_codex', family: 'openai_compatible', model: 'gpt-5.5', status: 'available', localProvider: true, sessionLocal: true, credentialReference: 'redacted', executionState: 'supported' },
    ])

    expect(provider?.id).toBe('local_codex')
    expect(provider?.model).toBe('gpt-5.5')
  })

  test('exposes saveModelProvider for local desktop provider settings', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()

    return expect(source).resolves.toContain('saveModelProvider')
  })

  test('preserves token usage and provider metadata as event details', () => {
    const usage = mapApiRunEvent({ id: 'evt-usage', run_id: 'run-1', thread_id: 'thread-1', sequence: 6, category: 'message', type: 'model.usage', summary: 'Usage', content: null, metadata: { input_tokens: 5, output_tokens: 8, total_tokens: 13 }, created_at: '2026-05-23T00:00:03Z' })
    const providerError = mapApiRunEvent({ id: 'evt-provider', run_id: 'run-1', thread_id: 'thread-1', sequence: 7, category: 'error', type: 'provider.error', summary: 'Provider unavailable', content: null, metadata: { provider: 'anthropic', code: 'overloaded' }, created_at: '2026-05-23T00:00:04Z' })

    expect(usage.group).toBe('model-stream')
    expect(usage.usage).toEqual({ inputTokens: 5, outputTokens: 8, totalTokens: 13 })
    expect(usage.detail).not.toContain('input_tokens')
    expect(providerError.group).toBe('error')
    expect(providerError.severity).toBe('error')
    expect(providerError.detail).toContain('anthropic')
    expect(providerError.detail).toContain('overloaded')
  })

  test('preserves canonical dotted worker backend and tool event types from real API events', () => {
    const worker = mapApiRunEvent({ id: 'evt-worker', run_id: 'run-1', thread_id: 'thread-1', sequence: 8, category: 'progress', type: 'worker.claimed', summary: 'Worker claimed', content: null, metadata: {}, created_at: '2026-05-23T00:00:05Z' })
    const tool = mapApiRunEvent({ id: 'evt-tool', run_id: 'run-1', thread_id: 'thread-1', sequence: 9, category: 'progress', type: 'tool_call_approval_required', summary: 'Tool approval required', content: null, metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' } }, created_at: '2026-05-23T00:00:06Z' })
    const backend = mapApiRunEvent({ id: 'evt-backend', run_id: 'run-1', thread_id: 'thread-1', sequence: 10, category: 'progress', type: 'backend.unavailable', summary: 'Backend unavailable', content: null, metadata: {}, created_at: '2026-05-23T00:00:07Z' })

    expect(worker.type).toBe('worker.claimed')
    expect(worker.group).toBe('worker-job')
    expect(tool).toMatchObject({ type: 'tool.call.approval_required', group: 'tool-call', status: 'blocked_on_tool_approval', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', arguments_summary: { timezone: 'UTC' } } })
    expect(backend.type).toBe('backend.unavailable')
    expect(backend.group).toBe('error')
    expect(backend.severity).toBe('error')
  })

  test('maps bounded loop metadata without duplicating loop fields in event detail', () => {
    const tool = mapApiRunEvent({ id: 'evt-loop', run_id: 'run-1', thread_id: 'thread-1', sequence: 9, category: 'progress', type: 'tool_call_approval_required', summary: 'Tool approval required', content: null, metadata: { tool_call_id: 'tc_glob_1', tool_name: 'workspace.glob', tool_group: 'workspace', loop_index: 2, loop_max: 3 }, created_at: '2026-05-25T00:00:06Z' })
    const todo = mapApiRunEvent({ id: 'evt-todo', run_id: 'run-1', thread_id: 'thread-1', sequence: 10, category: 'progress', type: 'work_todo_updated', summary: 'Todo updated', content: null, metadata: { updated_by: 'runtime' }, created_at: '2026-05-25T00:00:07Z' })

    expect(tool).toMatchObject({ type: 'tool.call.approval_required', status: 'blocked_on_tool_approval', metadata: { loop_index: 2, loop_max: 3 } })
    expect(tool.detail).not.toContain('loop_index')
    expect(tool.detail).not.toContain('loop_max')
    expect(todo).toMatchObject({ type: 'work.todo.updated', group: 'run-lifecycle' })
  })

  test('maps M6 queue worker and pipeline events into frontend statuses', () => {
    const queued = mapApiRunEvent({ id: 'evt-queued', run_id: 'run-1', thread_id: 'thread-1', sequence: 1, category: 'lifecycle', type: 'run_queued', summary: 'Run queued', content: null, metadata: {}, created_at: '2026-05-24T00:00:00Z' })
    const claimed = mapApiRunEvent({ id: 'evt-claimed', run_id: 'run-1', thread_id: 'thread-1', sequence: 2, category: 'progress', type: 'job_claimed', summary: 'Job claimed', content: null, metadata: {}, created_at: '2026-05-24T00:00:01Z' })
    const pipeline = mapApiRunEvent({ id: 'evt-pipeline', run_id: 'run-1', thread_id: 'thread-1', sequence: 3, category: 'progress', type: 'pipeline_step_started', summary: 'Pipeline step started', content: null, metadata: { step: 'invoke_runtime' }, created_at: '2026-05-24T00:00:02Z' })
    const pipelineFailed = mapApiRunEvent({ id: 'evt-pipeline-failed', run_id: 'run-1', thread_id: 'thread-1', sequence: 4, category: 'error', type: 'pipeline_step_failed', summary: 'Pipeline step failed', content: null, metadata: { step: 'prepare_context', error_code: 'invalid_request' }, created_at: '2026-05-24T00:00:03Z' })
    const stopped = mapApiRunEvent({ id: 'evt-stop', run_id: 'run-1', thread_id: 'thread-1', sequence: 4, category: 'progress', type: 'stop_requested', summary: 'Stop requested', content: null, metadata: {}, created_at: '2026-05-24T00:00:03Z' })

    expect(queued).toMatchObject({ type: 'run.queued', status: 'queued', group: 'run-lifecycle' })
    expect(claimed).toMatchObject({ type: 'job.claimed', status: 'running', group: 'worker-job' })
    expect(pipeline).toMatchObject({ type: 'pipeline.step.started', status: 'running', group: 'worker-job' })
    expect(pipeline.detail).toContain('invoke_runtime')
    expect(pipelineFailed).toMatchObject({ type: 'pipeline.step.failed', status: 'failed', group: 'error' })
    expect(pipelineFailed.detail).toContain('prepare_context')
    expect(stopped).toMatchObject({ type: 'run.stopping', status: 'stopping' })
  })

  test('maps safe persona summary metadata without prompt leakage', () => {
    const event = mapApiRunEvent({
      id: 'evt-persona',
      run_id: 'run-1',
      thread_id: 'thread-1',
      sequence: 1,
      category: 'progress',
      type: 'pipeline_step_completed',
      summary: 'Pipeline step completed',
      content: null,
      metadata: {
        step: 'prepare_context',
        persona_name: 'Default',
        persona_version: '2026-05-25.1',
      },
      created_at: '2026-05-25T00:00:00Z',
    })

    expect(event.detail).toContain('persona_name: Default')
    expect(event.detail).toContain('persona_version: 2026-05-25.1')
    expect(event.detail).not.toContain('system_prompt')
    expect(event.detail).not.toContain('You are')
  })

  test('real client exposes persona and installed skill list APIs', () => {
    const source = Bun.file(new URL('./realApiClient.ts', import.meta.url)).text()

    return source.then((text) => {
      expect(text).toContain("requireArrayField<ApiPersona>(body, 'personas'")
      expect(text).toContain("requireArrayField<ApiInstalledSkill>(body, 'skills'")
      expect(text).toContain('source_label')
      expect(text).toContain('persona_id: input.personaId')
      expect(text).toContain('personaId')
    })
  })
})
