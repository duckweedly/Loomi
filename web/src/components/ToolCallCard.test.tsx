import { describe, expect, test } from 'bun:test'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { ToolCallCard } from './ToolCallCard'

describe('ToolCallCard M7 approval-required state', () => {
  test('renders tool name redacted arguments and disabled approval controls', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
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
      },
    }))

    expect(html).toContain('runtime.get_current_time')
    expect(html).toContain('Approval required')
    expect(html).toContain('timezone')
    expect(html).toContain('UTC')
    expect(html).toContain('Approve')
    expect(html).toContain('Deny')
    expect(html).toContain('disabled')
    expect(html).not.toContain('api_key')
    expect(html).not.toContain('secret')
  })

  test('renders active approval controls only when decision handlers exist', () => {
    const toolCall = {
      id: 'tool_1',
      toolCallId: 'tc_1',
      name: 'runtime.get_current_time',
      status: 'approval_required' as const,
      approvalStatus: 'required' as const,
      executionStatus: 'blocked' as const,
      summary: 'Approval required',
      input: '{"timezone":"UTC"}',
      output: '',
      argumentsSummary: { timezone: 'UTC' },
      resultSummary: null,
      errorCode: null,
      errorMessage: null,
    }

    const active = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall,
      onApprove: () => {},
      onDeny: () => {},
    }))
    const approved = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: { ...toolCall, status: 'approved', approvalStatus: 'approved', executionStatus: 'not_started' },
      onApprove: () => {},
      onDeny: () => {},
    }))

    expect(active).toContain('Approve')
    expect(active).toContain('Deny')
    expect(active).not.toContain('disabled')
    expect(approved).not.toContain('Approve')
    expect(approved).not.toContain('Deny')
  })

  test('renders executing succeeded failed denied and cancelled terminal states', () => {
    const base = {
      id: 'tool_1',
      toolCallId: 'tc_1',
      name: 'runtime.get_current_time',
      status: 'executing' as const,
      approvalStatus: 'approved' as const,
      executionStatus: 'executing' as const,
      summary: 'Tool call executing',
      input: '',
      output: '',
      argumentsSummary: { timezone: 'UTC' },
      resultSummary: null,
      errorCode: null,
      errorMessage: null,
    }

    const executing = renderToStaticMarkup(createElement(ToolCallCard, { toolCall: base }))
    const succeeded = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: { ...base, status: 'succeeded', executionStatus: 'succeeded', summary: 'Tool call succeeded', resultSummary: { timezone: 'UTC', local_time: '2026-05-26T10:00:00Z' } },
    }))
    const failed = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: { ...base, status: 'failed', executionStatus: 'failed', summary: 'Tool call failed', errorCode: 'validation_failed', errorMessage: 'Invalid timezone' },
    }))
    const denied = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: { ...base, status: 'denied', approvalStatus: 'denied', executionStatus: 'cancelled', summary: 'Tool call denied' },
    }))
    const cancelled = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: { ...base, status: 'cancelled', approvalStatus: 'cancelled', executionStatus: 'cancelled', summary: 'Tool call cancelled' },
    }))

    expect(executing).toContain('Executing')
    expect(succeeded).toContain('Result')
    expect(succeeded).toContain('local_time')
    expect(succeeded).toContain('2026-05-26T10:00:00Z')
    expect(failed).toContain('Redacted error')
    expect(failed).toContain('validation_failed')
    expect(failed).toContain('Invalid timezone')
    expect(denied).toContain('Denied')
    expect(cancelled).toContain('Cancelled')
    expect(failed).not.toContain('api_key')
    expect(failed).not.toContain('secret')
  })

  test('renders workspace grep results as bounded readable summaries', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_1',
        toolCallId: 'tc_grep',
        name: 'workspace.grep',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { query: 'workspace', path: '.', limit: 5 },
        resultSummary: { matches: [{ path: 'README.md', line: 2, preview: 'workspace read tools' }], match_count: 1, truncated: false },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('workspace.grep')
    expect(html).toContain('README.md:2')
    expect(html).toContain('workspace read tools')
    expect(html).not.toContain('[object Object]')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('sk-live')
  })

  test('renders workspace write and edit result summaries without raw object output', () => {
    const write = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_write',
        toolCallId: 'tc_write',
        name: 'workspace.write_file',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { path: 'internal/generated.txt' },
        resultSummary: { path: 'internal/generated.txt', bytes_written: 12, created: true, truncated: false },
        errorCode: null,
        errorMessage: null,
      },
    }))
    const edit = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_edit',
        toolCallId: 'tc_edit',
        name: 'workspace.edit',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { path: 'internal/generated.txt' },
        resultSummary: { path: 'internal/generated.txt', replacements: 1, bytes_before: 12, bytes_after: 13 },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(write).toContain('workspace.write_file')
    expect(write).toContain('bytes_written')
    expect(write).toContain('internal/generated.txt')
    expect(edit).toContain('workspace.edit')
    expect(edit).toContain('replacements')
    expect(edit).toContain('bytes_after')
    expect(`${write}${edit}`).not.toContain('[object Object]')
    expect(`${write}${edit}`).not.toContain('/Users/')
    expect(`${write}${edit}`).not.toContain('sk-live')
  })

  test('renders workspace exec command result summaries without raw object output', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_exec',
        toolCallId: 'tc_exec',
        name: 'workspace.exec_command',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { command: ['printf', 'hello'], cwd: '.' },
        resultSummary: { cwd: '.', exit_code: 0, stdout: 'hello', stderr: '', timed_out: false, stdout_truncated: false, stderr_truncated: false },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('workspace.exec_command')
    expect(html).toContain('exit_code')
    expect(html).toContain('stdout')
    expect(html).toContain('hello')
    expect(html).not.toContain('[object Object]')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('sk-live')
  })

  test('renders todo write planning summaries without raw object output', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_todo',
        toolCallId: 'tc_todo',
        name: 'runtime.todo_write',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { items: [{ title: 'Inspect tools', status: 'completed' }, { title: 'Implement todo', status: 'in_progress' }] },
        resultSummary: { total: 2, completed_count: 1, in_progress_count: 1, pending_count: 0, items: [{ title: 'Inspect tools', status: 'completed' }, { title: 'Implement todo', status: 'in_progress' }] },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('runtime.todo_write')
    expect(html).toContain('completed_count')
    expect(html).toContain('Inspect tools')
    expect(html).toContain('in_progress')
    expect(html).not.toContain('[object Object]')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('sk-live')
  })

  test('renders mcp call tool summaries without raw object output', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_mcp',
        toolCallId: 'tc_mcp',
        name: 'mcp.call_tool',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { server: 'local', tool: 'echo', arguments: { message: 'hello mcp' } },
        resultSummary: { server: 'local', tool: 'echo', message: 'hello mcp', side_effect: 'none' },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('mcp.call_tool')
    expect(html).toContain('local')
    expect(html).toContain('echo')
    expect(html).toContain('hello mcp')
    expect(html).toContain('side_effect')
    expect(html).not.toContain('[object Object]')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('sk-live')
  })
})
