import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RunRail } from './RunRail'

describe('RunRail tool continuation runtime states', () => {
  test('labels initial model tool execution continuation and final phases', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-initial', sequence: 1, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Now', status: 'running', group: 'model-stream', metadata: { model_phase: 'initial' } },
          { id: 'evt-tool', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_call_id: 'tc_1', tool_name: 'runtime.get_current_time', result_summary: { timezone: 'UTC' } } },
          { id: 'evt-continuation', sequence: 3, type: 'message.model_output_delta', label: 'message', detail: 'Model output delta', time: 'Now', status: 'running', group: 'model-stream', metadata: { model_phase: 'continuation' } },
          { id: 'evt-final', sequence: 4, type: 'run.completed', label: 'run', detail: 'Run completed', time: 'Now', status: 'completed', group: 'run-lifecycle' },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Initial model phase')
    expect(html).toContain('Tool call succeeded')
    expect(html).toContain('Continuation model phase')
    expect(html).toContain('Run completed')
  })

  test('shows M9 pipeline foundation stage rows', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-context', sequence: 1, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: prepare_context', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'prepare_context' } },
          { id: 'evt-tools', sequence: 2, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: resolve_tools', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'resolve_tools' } },
          { id: 'evt-runtime', sequence: 3, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: invoke_runtime', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'invoke_runtime' } },
          { id: 'evt-finalize', sequence: 4, type: 'pipeline.step.completed', label: 'progress', detail: 'Pipeline step completed · step: finalize', time: 'Now', status: 'running', group: 'worker-job', metadata: { step: 'finalize' } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('prepare_context')
    expect(html).toContain('resolve_tools')
    expect(html).toContain('invoke_runtime')
    expect(html).toContain('finalize')
  })

  test('shows safe persona summary without prompt text', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          {
            id: 'evt-context',
            sequence: 1,
            type: 'pipeline.step.completed',
            label: 'progress',
            detail: 'Pipeline step completed · step: prepare_context · persona_name: Default · persona_version: 2026-05-25.1',
            time: 'Now',
            status: 'running',
            group: 'worker-job',
            metadata: { step: 'prepare_context', persona_name: 'Default', persona_version: '2026-05-25.1' },
          },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('persona_name: Default')
    expect(html).toContain('persona_version: 2026-05-25.1')
    expect(html).not.toContain('system_prompt')
    expect(html).not.toContain('You are')
  })

  test('shows workspace tool lifecycle states', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-requested', sequence: 1, type: 'tool.call.requested', label: 'tool', detail: 'Tool call requested', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_group: 'workspace', tool_name: 'workspace.read' } },
          { id: 'evt-required', sequence: 2, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_group: 'workspace', tool_name: 'workspace.read' } },
          { id: 'evt-approved', sequence: 3, type: 'tool.call.approved', label: 'tool', detail: 'Tool call approved', time: 'Now', status: 'queued', group: 'tool-call', metadata: { tool_group: 'workspace', tool_name: 'workspace.read' } },
          { id: 'evt-executing', sequence: 4, type: 'tool.call.executing', label: 'tool', detail: 'Tool call executing', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_group: 'workspace', tool_name: 'workspace.read' } },
          { id: 'evt-succeeded', sequence: 5, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_group: 'workspace', tool_name: 'workspace.read' } },
          { id: 'evt-failed', sequence: 6, type: 'tool.call.failed', label: 'tool', detail: 'Tool call failed', time: 'Now', status: 'failed', group: 'tool-call', severity: 'error', metadata: { tool_group: 'workspace', tool_name: 'workspace.read' } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Workspace tool')
    expect(html).toContain('Tool approval required')
    expect(html).toContain('Tool call executing')
    expect(html).toContain('Tool call succeeded')
    expect(html).toContain('Tool call failed')
  })

  test('shows workspace mutation risk and write-capable lifecycle states without raw content', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-required', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_group: 'workspace', tool_name: 'workspace.write_file', arguments_summary: { path: 'src/generated.txt', content: '[redacted]' }, loop_index: 1, loop_max: 3 } },
          { id: 'evt-succeeded', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_group: 'workspace', tool_name: 'workspace.edit', result_summary: { operation: 'edit', path: 'src/notes.txt', changed: true } } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Workspace mutation tool')
    expect(html).toContain('high risk')
    expect(html).toContain('write-capable')
    expect(html).toContain('Tool approval required')
    expect(html).toContain('Tool call succeeded')
    expect(html).not.toContain('created\\n')
    expect(html).not.toContain('/tmp/')
  })

  test('shows sandbox exec risk and exec-capable lifecycle states without host paths or secrets', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-required', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_group: 'sandbox', tool_name: 'sandbox.exec_command', arguments_summary: { argv: ['printf', 'hello'], cwd: '.', timeout_ms: 5000 }, loop_index: 1, loop_max: 3 } },
          { id: 'evt-succeeded', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_group: 'sandbox', tool_name: 'sandbox.exec_command', result_summary: { operation: 'exec_command', cwd: '.', exit_code: 0, stdout: 'hello' } } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Sandbox exec tool')
    expect(html).toContain('high risk')
    expect(html).toContain('exec-capable')
    expect(html).toContain('Tool approval required')
    expect(html).toContain('Tool call succeeded')
    expect(html).not.toContain('/tmp/')
    expect(html).not.toContain('TOKEN')
  })

  test('shows lsp read-only lifecycle states without host paths or secrets', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-required', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_group: 'lsp', tool_name: 'lsp.symbols', arguments_summary: { path: 'src/main.go', query: 'Tool' }, loop_index: 1, loop_max: 3 } },
          { id: 'evt-succeeded', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_group: 'lsp', tool_name: 'lsp.symbols', result_summary: { operation: 'symbols', path: 'src/main.go', count: 1 } } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('LSP read-only tool')
    expect(html).toContain('low risk')
    expect(html).toContain('workspace-scoped')
    expect(html).toContain('Tool approval required')
    expect(html).toContain('Tool call succeeded')
    expect(html).not.toContain('id_ed25519')
    expect(html).not.toContain('SECRET')
  })

  test('shows web fetch lifecycle states without raw bodies or secrets', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-required', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_group: 'web', tool_name: 'web.fetch', arguments_summary: { url: 'https://example.com/docs' }, loop_index: 1, loop_max: 3 } },
          { id: 'evt-succeeded', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'Tool call succeeded', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_group: 'web', tool_name: 'web.fetch', result_summary: { operation: 'fetch', status_code: 200, final_url: 'https://example.com/docs', truncated: false } } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Web fetch tool')
    expect(html).toContain('medium risk')
    expect(html).toContain('public HTTP only')
    expect(html).toContain('Tool approval required')
    expect(html).toContain('Tool call succeeded')
    expect(html).not.toContain('Set-Cookie')
    expect(html).not.toContain('Authorization')
    expect(html).not.toContain('sk-secret')
  })

  test('shows browser automation lifecycle states without raw HTML or cookies', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-open', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'browser.open waiting for approval', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_group: 'browser', tool_name: 'browser.open', arguments_summary: { url: 'https://example.com/docs' }, loop_index: 1, loop_max: 3 } },
          { id: 'evt-click', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'browser.click_link completed', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_group: 'browser', tool_name: 'browser.click_link', result_summary: { operation: 'click_link', final_url: 'https://example.com/docs/next', title: 'Next', link_count: 2 } } },
          { id: 'evt-snapshot', sequence: 3, type: 'tool.call.succeeded', label: 'tool', detail: 'browser.snapshot completed', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_group: 'browser', tool_name: 'browser.snapshot', result_summary: { operation: 'snapshot', current_url: 'https://example.com/docs/next', title: 'Next', text_excerpt: 'Safe bounded text' } } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Browser automation tool')
    expect(html).toContain('medium risk')
    expect(html).toContain('public HTTP only')
    expect(html).toContain('browser.open waiting for approval')
    expect(html).toContain('browser.click_link completed')
    expect(html).toContain('browser.snapshot completed')
    expect(html).not.toContain('<html')
    expect(html).not.toContain('Set-Cookie')
    expect(html).not.toContain('Authorization')
    expect(html).not.toContain('sk-secret')
  })

  test('shows artifact lifecycle states without raw unbounded content', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-create', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'artifact.create_text waiting for approval', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_group: 'artifact', tool_name: 'artifact.create_text', arguments_summary: { title: 'Notes' }, loop_index: 1, loop_max: 3 } },
          { id: 'evt-list', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'artifact.list completed', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_group: 'artifact', tool_name: 'artifact.list', result_summary: { operation: 'list', count: 1 } } },
          { id: 'evt-read', sequence: 3, type: 'tool.call.succeeded', label: 'tool', detail: 'artifact.read completed', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_group: 'artifact', tool_name: 'artifact.read', result_summary: { operation: 'read', title: 'Notes', text_excerpt: 'Safe bounded text' } } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Artifact runtime tool')
    expect(html).toContain('medium risk')
    expect(html).toContain('non-executable')
    expect(html).toContain('artifact.create_text waiting for approval')
    expect(html).toContain('artifact.list completed')
    expect(html).toContain('artifact.read completed')
    expect(html).not.toContain('raw_result')
    expect(html).not.toContain('sk-secret')
  })

  test('shows agent coordination lifecycle states without autonomous execution', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-spawn', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'agent.spawn waiting for approval', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_group: 'agent', tool_name: 'agent.spawn', arguments_summary: { role: 'reviewer', goal: 'Review implementation' }, loop_index: 1, loop_max: 3 } },
          { id: 'evt-list', sequence: 2, type: 'tool.call.succeeded', label: 'tool', detail: 'agent.list completed', time: 'Now', status: 'running', group: 'tool-call', metadata: { tool_group: 'agent', tool_name: 'agent.list', result_summary: { operation: 'list', count: 1 } } },
          { id: 'evt-complete', sequence: 3, type: 'tool.call.succeeded', label: 'tool', detail: 'agent.complete completed', time: 'Now', status: 'completed', group: 'tool-call', metadata: { tool_group: 'agent', tool_name: 'agent.complete', result_summary: { operation: 'complete', status: 'completed', result_summary: 'No safety issue found', autonomous_execution: false } } },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
    }))

    expect(html).toContain('Agent coordination tool')
    expect(html).toContain('medium risk')
    expect(html).toContain('no autonomous execution')
    expect(html).toContain('agent.spawn waiting for approval')
    expect(html).toContain('agent.list completed')
    expect(html).toContain('agent.complete completed')
    expect(html).not.toContain('raw_result')
    expect(html).not.toContain('sk-secret')
  })

  test('shows bounded loop index continuation limit and stopped state', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'blocked_on_tool_approval',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [
          { id: 'evt-glob', sequence: 1, type: 'tool.call.approval_required', label: 'tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval', group: 'tool-call', metadata: { tool_name: 'workspace.glob', loop_index: 1, loop_max: 3 } },
          { id: 'evt-continuation', sequence: 2, type: 'message.model_output_delta', label: 'message', detail: 'Continuation after workspace.glob', time: 'Now', status: 'running', group: 'model-stream', metadata: { model_phase: 'continuation' } },
          { id: 'evt-limit', sequence: 3, type: 'run.failed', label: 'run', detail: 'Tool loop limit reached', time: 'Now', status: 'failed', group: 'error', severity: 'error', metadata: { error_code: 'tool_loop_limit_reached', loop_count: 3, max_tool_calls: 3 } },
          { id: 'evt-stopped', sequence: 4, type: 'run.stopped', label: 'run', detail: 'Run stopped', time: 'Now', status: 'stopped', group: 'run-lifecycle' },
        ],
      },
      open: true,
      onOpenArtifact: () => {},
      onStopRun: () => {},
    }))

    expect(html).toContain('Loop 1/3')
    expect(html).toContain('Continuation model phase')
    expect(html).toContain('Tool loop limit reached')
    expect(html).toContain('Run stopped')
    expect(html).toContain('Stop run')
  })
})
