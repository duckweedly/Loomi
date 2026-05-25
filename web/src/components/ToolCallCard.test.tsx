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

  test('enables approval controls when real action handlers are available', () => {
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
      onApprove: () => undefined,
      onDeny: () => undefined,
    }))

    expect(html).toContain('Approve')
    expect(html).not.toContain('disabled=""')
  })

  test('renders succeeded result summary without approval actions', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_1',
        toolCallId: 'tc_1',
        name: 'runtime.get_current_time',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { timezone: 'UTC' },
        resultSummary: { iso_time: '2026-05-25T10:00:00Z', timezone: 'UTC' },
        errorCode: null,
        errorMessage: null,
      },
      onApprove: () => undefined,
      onDeny: () => undefined,
    }))

    expect(html).toContain('succeeded')
    expect(html).toContain('Result')
    expect(html).toContain('2026-05-25T10:00:00Z')
    expect(html).not.toContain('Approving')
    expect(html).not.toContain('Denying')
  })
})
