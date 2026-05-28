import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { ToolCallCard } from './ToolCallCard'

describe('ToolCallCard M7 approval-required state', () => {
  test('uses animal-island-ui Button for approval actions and removes old lobe tags', () => {
    const source = readFileSync(resolve(import.meta.dir, 'ToolCallCard.tsx'), 'utf8')

    expect(source).toContain("import { Button } from 'animal-island-ui'")
    expect(source).not.toContain("from '@lobehub/ui'")
    expect(source).toContain('className="tool-status-pill"')
    expect(source).toContain('<Button className="primary"')
    expect(source).toContain('<Button disabled={actionsDisabled}')
  })

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

    expect(html).toContain('Awaiting approval')
    expect(html).toContain('Tool phases')
    expect(html).toContain('Request')
    expect(html).toContain('Approval')
    expect(html).toContain('Execution')
    expect(html).toContain('Result')
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

    expect(html).toContain('Completed')
    expect(html).toContain('timezone: UTC')
    expect(html).not.toContain('2026-05-25T10:00:00Z')
    expect(html).not.toContain('Tool phases')
    expect(html).not.toContain('Tool completed.')
    expect(html).not.toContain('tool-grid')
    expect(html).not.toContain('Approving')
    expect(html).not.toContain('Denying')
  })

  test('renders completed artifact tools as compact document resources', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      locale: 'zh',
      toolCall: {
        id: 'tool_artifact',
        toolCallId: 'tc_artifact',
        name: 'artifact.create_text',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { title: '三句话的 Markdown' },
        resultSummary: { operation: 'create_text', artifact_id: 'art_mock', title: '三句话的 Markdown', filename: '三句话.md', mime_type: 'text/markdown', text_excerpt: '# 三句话' },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('artifact-resource-card')
    expect(html).toContain('三句话的 Markdown')
    expect(html).toContain('Markdown 文档')
    expect(html).not.toContain('请求')
    expect(html).not.toContain('结果')
  })

  test('renders human-first tool labels without raw names in primary UI', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_1',
        toolCallId: 'tc_1',
        name: 'workspace.read',
        status: 'approval_required',
        approvalStatus: 'required',
        executionStatus: 'blocked',
        summary: 'Read web/src/App.tsx',
        input: 'path: web/src/App.tsx',
        output: '',
        argumentsSummary: { path: 'web/src/App.tsx' },
        resultSummary: null,
        errorCode: null,
        errorMessage: null,
      },
      onApprove: () => undefined,
      onDeny: () => undefined,
    }))

    expect(html).toContain('Read project files')
    expect(html).toContain('path')
    expect(html).not.toContain('web/src/App.tsx')
    expect(html).not.toContain('workspace.read')
  })

  test('renders code-agent patch preview and apply cards with safe diff metadata', () => {
    const previewHtml = renderToStaticMarkup(createElement(ToolCallCard, {
      toolCall: {
        id: 'tool_preview',
        toolCallId: 'tc_preview',
        name: 'workspace.patch_preview',
        status: 'approval_required',
        approvalStatus: 'required',
        executionStatus: 'blocked',
        summary: 'Tool approval required',
        input: '',
        output: '',
        argumentsSummary: { path: 'src/notes.txt', old_text: '[redacted]', new_text: '[redacted]' },
        resultSummary: null,
        errorCode: null,
        errorMessage: null,
      },
      onApprove: () => undefined,
      onDeny: () => undefined,
    }))

    const applyHtml = renderToStaticMarkup(createElement(ToolCallCard, {
      defaultExpanded: true,
      toolCall: {
        id: 'tool_apply',
        toolCallId: 'tc_apply',
        name: 'workspace.patch_apply',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { path: 'src/notes.txt' },
        resultSummary: { operation: 'patch_apply', path: 'src/notes.txt', changed: true, diff: '--- src/notes.txt\n+++ src/notes.txt\n-needle\n+daily loop\n', preview_id: 'patch_123' },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(previewHtml).toContain('Preview workspace patch')
    expect(previewHtml).toContain('Awaiting approval')
    expect(previewHtml).toContain('Approve')
    expect(previewHtml).not.toContain('workspace.patch_preview')
    expect(previewHtml).not.toContain('src/notes.txt')
    expect(applyHtml).toContain('Apply workspace patch')
    expect(applyHtml).toContain('changed: true')
    expect(applyHtml).toContain('diff: [redacted]')
    expect(applyHtml).not.toContain('workspace.patch_apply')
    expect(applyHtml).not.toContain('src/notes.txt')
    expect(applyHtml).not.toContain('needle')
    expect(applyHtml).not.toContain('patch_123')
  })

  test('renders directory browser tools with safe summaries and no absolute paths', () => {
    const listHtml = renderToStaticMarkup(createElement(ToolCallCard, {
      defaultExpanded: true,
      toolCall: {
        id: 'tool_list',
        toolCallId: 'tc_list',
        name: 'workspace.list_directory',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { path: '/Users/xuean/Downloads', max_entries: 200, depth: 2 },
        resultSummary: { operation: 'list_directory', path: '.', returned_entries: 12, total_entries_seen: 12, directories_count: 3, files_count: 9, truncated: false },
        errorCode: null,
        errorMessage: null,
      },
    }))
    const summaryHtml = renderToStaticMarkup(createElement(ToolCallCard, {
      defaultExpanded: true,
      toolCall: {
        id: 'tool_tree',
        toolCallId: 'tc_tree',
        name: 'workspace.tree_summary',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { path: '.', max_entries: 200, depth: 3 },
        resultSummary: { operation: 'tree_summary', by_kind: { code: 2, document: 4 }, by_extension: { '.go': 2 }, largest_files: [{ path: 'src/secret-token.txt', size: 100 }], recent_files: [{ path: '/Users/xuean/Downloads/report.pdf' }] },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(listHtml).toContain('Read directory')
    expect(listHtml).toContain('returned_entries: 12')
    expect(listHtml).toContain('directories_count: 3')
    expect(listHtml).not.toContain('/Users/')
    expect(summaryHtml).toContain('Summarize directory')
    expect(summaryHtml).toContain('by_kind')
    expect(summaryHtml).not.toContain('secret-token')
    expect(summaryHtml).not.toContain('/Users/')
  })

  test('shows safe workspace label for workspace tool cards', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      locale: 'zh',
      toolCall: {
        id: 'tool_workspace',
        toolCallId: 'tc_workspace',
        name: 'workspace.read',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { path: '/Users/xuean/Downloads/receipt.txt' },
        resultSummary: { workspace_label: 'Downloads', path: 'receipt.txt', bytes_read: 6, truncated: false },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('正在读取：Downloads')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('receipt.txt')
  })

  test('renders sandbox process lifecycle cards with bounded safe summaries', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      defaultExpanded: true,
      toolCall: {
        id: 'tool_process',
        toolCallId: 'tc_process',
        name: 'sandbox.terminate_process',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { process_id: 'sp_abc123' },
        resultSummary: { operation: 'terminate_process', process_id: 'sp_abc123', status: 'terminated', terminal_summary: 'terminated exit_code=-1', stdout: 'TOKEN=secret /Users/xuean/Loomi' },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('Terminate sandbox process')
    expect(html).toContain('process_id: sp_abc123')
    expect(html).toContain('terminal_summary: terminated exit_code=-1')
    expect(html).not.toContain('sandbox.terminate_process')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('TOKEN=secret')
  })

  test('renders sandbox process resume metadata without unsafe paths or raw output', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      defaultExpanded: true,
      toolCall: {
        id: 'tool_process_resume',
        toolCallId: 'tc_process_resume',
        name: 'sandbox.continue_process',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { process_id: 'sp_resume', cursor: 0 },
        resultSummary: { operation: 'continue_process', process_id: 'sp_resume', argv_summary: ['cat'], cwd_alias: '.', status: 'exited', terminal_summary: 'exited exit_code=0', next_cursor: 2048, stdout: 'token=secret /Users/xuean/Loomi/project', started_at: '2026-05-27T09:00:00Z', updated_at: '2026-05-27T09:00:01Z', ended_at: '2026-05-27T09:00:01Z' },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('Continue sandbox process')
    expect(html).toContain('argv_summary: cat')
    expect(html).toContain('cwd_alias: .')
    expect(html).toContain('status: exited')
    expect(html).toContain('next_cursor: 2048')
    expect(html).toContain('terminal_summary: exited exit_code=0')
    expect(html).not.toContain('/Users/')
    expect(html).not.toContain('token=secret')
  })

  test('hides unknown raw tool names behind a generic label', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      locale: 'zh',
      toolCall: {
        id: 'tool_legacy',
        toolCallId: 'tc_legacy',
        name: 'scan_reference',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Legacy fixture tool',
        input: 'Reference workspace files',
        output: 'Work mode timeline',
        argumentsSummary: null,
        resultSummary: null,
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('使用工具')
    expect(html).toContain('已完成')
    expect(html).not.toContain('scan_reference')
    expect(html).not.toContain('completed')
  })

  test('renders Chinese web search approval card without raw runtime fields', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      locale: 'zh',
      toolCall: {
        id: 'tool_1',
        toolCallId: 'tc_1',
        name: 'web.search',
        status: 'approval_required',
        approvalStatus: 'required',
        executionStatus: 'blocked',
        summary: 'Tool approval required · approval_status: required · tool_call_id: tc_1',
        input: '',
        output: '',
        argumentsSummary: { query: '今天最新 AI 新闻', provider: 'brave', limit: 5, api_key: 'BSA-secret' },
        resultSummary: null,
        errorCode: null,
        errorMessage: null,
      },
      onApprove: () => undefined,
      onDeny: () => undefined,
    }))

    expect(html).toContain('搜索网页')
    expect(html).toContain('等待确认')
    expect(html).toContain('搜索词: 今天最新 AI 新闻')
    expect(html).toContain('服务: brave')
    expect(html).toContain('数量: 5')
    expect(html).toContain('允许')
    expect(html).toContain('拒绝')
    expect(html).not.toContain('web.search')
    expect(html).not.toContain('approval_status')
    expect(html).not.toContain('tool_call_id')
    expect(html).not.toContain('BSA-secret')
  })

  test('renders Chinese web search result without raw backend summary fields', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      locale: 'zh',
      toolCall: {
        id: 'tool_1',
        toolCallId: 'tc_1',
        name: 'web.search',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded · approval_status: approved · execution_status: succeeded · tool_call_id: tc_1',
        input: '',
        output: '',
        argumentsSummary: { query: 'latest AI news', provider: 'brave', limit: 5 },
        resultSummary: { operation: 'search', provider: 'brave', query: 'latest AI news', result_count: 2, tool: 'web.search', items: [{ title: 'Reuters AI News', url: 'https://example.com/reuters', snippet: 'safe summary' }, { title: 'LLM News Today', url: 'https://example.com/llm', snippet: 'safe summary' }] },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('搜索网页')
    expect(html).toContain('搜索词: latest AI news')
    expect(html).not.toContain('Reuters AI News')
    expect(html).not.toContain('工具阶段')
    expect(html).not.toContain('工具已完成')
    expect(html).not.toContain('来源')
    expect(html).not.toContain('example.com')
    expect(html).not.toContain('LLM News Today')
    expect(html).not.toContain('结果: 2')
    expect(html).not.toContain('approval_status')
    expect(html).not.toContain('execution_status')
    expect(html).not.toContain('tool_call_id')
    expect(html).not.toContain('operation')
    expect(html).not.toContain('tool: web.search')
  })

  test('reveals completed tool request result and sources only when expanded', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      locale: 'zh',
      defaultExpanded: true,
      toolCall: {
        id: 'tool_search_expanded',
        toolCallId: 'tc_search_expanded',
        name: 'web.search',
        status: 'succeeded',
        approvalStatus: 'approved',
        executionStatus: 'succeeded',
        summary: 'Tool call succeeded',
        input: '',
        output: '',
        argumentsSummary: { query: 'latest AI news', provider: 'brave', limit: 5 },
        resultSummary: { operation: 'search', provider: 'brave', query: 'latest AI news', result_count: 2, tool: 'web.search', items: [{ title: 'Reuters AI News', url: 'https://example.com/reuters', snippet: 'safe summary' }, { title: 'LLM News Today', url: 'https://example.com/llm', snippet: 'safe summary' }] },
        errorCode: null,
        errorMessage: null,
      },
    }))

    expect(html).toContain('aria-expanded="true"')
    expect(html).toContain('请求')
    expect(html).toContain('结果')
    expect(html).toContain('来源')
    expect(html).toContain('Reuters AI News')
    expect(html).toContain('example.com')
    expect(html).toContain('LLM News Today')
    expect(html).not.toContain('tool: web.search')
  })

  test('redacts sensitive tool arguments results and raw text payloads', () => {
    const html = renderToStaticMarkup(createElement(ToolCallCard, {
      defaultExpanded: true,
      toolCall: {
        id: 'tool_1',
        toolCallId: 'tc_1',
        name: 'workspace.read',
        status: 'failed',
        approvalStatus: 'approved',
        executionStatus: 'failed',
        summary: 'Read failed',
        input: '{"path":"/home/xuean/project/.env","authorization":"Bearer sk-live-secret","cookie":"session=abc","token":"hidden"}',
        output: 'stdout raw payload stderr raw payload token=hidden api_key=secret secret=value password=pw credential=cred session=sid',
        argumentsSummary: { path: '/home/xuean/project/.env', authorization: 'Bearer sk-live-secret', cookie: 'session=abc' },
        resultSummary: { stdout: 'stdout raw payload', stderr: 'stderr raw payload', raw_body: 'raw body payload', status_code: 403 },
        errorCode: 'unsafe_path',
        errorMessage: 'Cannot read /home/xuean/project/.env with Authorization Bearer sk-live-secret token=hidden api_key=secret secret=value password=pw credential=cred session=sid',
      },
      onApprove: () => undefined,
      onDeny: () => undefined,
    }))

    expect(html).toContain('Read project files')
    expect(html).toContain('status_code: 403')
    expect(html).not.toContain('workspace.read')
    expect(html).not.toContain('/home/xuean')
    expect(html).not.toContain('.env')
    expect(html).not.toContain('Authorization')
    expect(html).not.toContain('sk-live-secret')
    expect(html).not.toContain('session=abc')
    expect(html).not.toContain('token=hidden')
    expect(html).not.toContain('api_key=secret')
    expect(html).not.toContain('secret=value')
    expect(html).not.toContain('password=pw')
    expect(html).not.toContain('credential=cred')
    expect(html).not.toContain('session=sid')
    expect(html).not.toContain('stdout raw payload')
    expect(html).not.toContain('stderr raw payload')
    expect(html).not.toContain('raw body payload')
  })
})
