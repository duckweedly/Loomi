import { describe, expect, test } from 'bun:test'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { createElement } from 'react'
import { renderToStaticMarkup } from 'react-dom/server'
import { RunRail } from './RunRail'

describe('RunRail restrained runtime polish', () => {
  test('uses compact Scenario controls instead of prominent script buttons', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('Scenario')
    expect(source).toContain('Success')
    expect(source).toContain('Fail')
    expect(source).not.toContain('成功剧本')
    expect(source).not.toContain('失败剧本')
  })

  test('keeps stop out of the runtime rail because the composer primary button owns it', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('Stop run')
    expect(source).not.toContain('runtime-stop-button ghost')
  })

  test('styles timeline with quiet dots and compact agent card', () => {
    const css = [
      readFileSync(resolve(import.meta.dir, '../styles.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/30-runtime-panels.css'), 'utf8'),
      readFileSync(resolve(import.meta.dir, '../styles/92-unified-workspace.css'), 'utf8'),
    ].join('\n')

    expect(css).toContain('.progress-row::before')
    expect(css).toContain('width: 7px')
    expect(css).toContain('.agent-motion-card.compact')
    expect(css).toContain('.runtime-script-switch.compact')
  })

  test('renders stable grouped timeline sections with error and usage details', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'failed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-run', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'run.created', label: 'Run', detail: 'created', time: 'Now', status: 'running' },
          { id: 'evt-model', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'model.usage', label: 'Usage', detail: 'usage', time: 'Now', status: 'running', usage: { inputTokens: 7, outputTokens: 11 } },
          { id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'worker.claimed', label: 'Worker', detail: 'claimed', time: 'Now', status: 'running' },
          { id: 'evt-error', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'provider.error', label: 'Provider', detail: 'provider failed', time: 'Now', status: 'failed' },
        ],
      },
    }))

    expect(html).toContain('Recent activity')
    expect(html).toContain('Run failed')
    expect(html).toContain('provider failed')
    expect(html).toContain('runtime-event-group recent')
    expect(html).not.toContain('Run created')
    expect(html).not.toContain('Model usage')
    expect(html).not.toContain('7 in / 11 out')
  })

  test('renders productized M6 worker event labels and unknown fallback', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-claim', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job_claimed', label: 'Job', detail: 'raw claim', time: 'Now', status: 'running' },
          { id: 'evt-lease', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'lease_renewed', label: 'Lease', detail: 'raw lease', time: 'Now', status: 'running' },
          { id: 'evt-recovering', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'job_recovering', label: 'Job', detail: 'raw recovery', time: 'Now', status: 'recovering' },
          { id: 'evt-unknown', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'future_worker_event', label: 'Future', detail: 'raw future', time: 'Now', status: 'running' },
        ],
      },
    }))

    expect(html).toContain('No activity yet')
    expect(html).not.toContain('Job claimed by worker')
    expect(html).not.toContain('Lease renewed')
    expect(html).not.toContain('Job recovering')
    expect(html).not.toContain('Unknown worker event')
    expect(html).not.toContain('future_worker_event')
  })

  test('renders productized worker event labels from Chinese i18n copy', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      locale: 'zh',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'recovering',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-claim', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'job_claimed', label: 'Job', detail: 'raw claim', time: 'Now', status: 'running' },
          { id: 'evt-unknown', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'future_worker_event', label: 'Future', detail: 'raw future', time: 'Now', status: 'running' },
        ],
      },
    }))

    expect(html).toContain('暂无活动')
    expect(html).not.toContain('Worker 已领取任务')
    expect(html).not.toContain('未知 Worker 事件')
  })

  test('renders provider errors and cancelled events with distinct row classes', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'cancelled',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-provider', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'provider.error', label: 'Provider', detail: 'provider failed', time: 'Now', status: 'running', group: 'error', severity: 'error' },
          { id: 'evt-cancelled', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'run.cancelled', label: 'Cancelled', detail: 'cancelled', time: 'Now', status: 'cancelled', group: 'run-lifecycle', severity: 'warning' },
        ],
      },
    }))

    expect(html).toContain('progress-row failed')
    expect(html).toContain('progress-row warning')
    expect(html).toContain('provider failed')
    expect(html).toContain('cancelled')
  })

  test('shows capability status detail in the rail', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      capabilityStatus: 'provider-unavailable',
      run: null,
    }))

    expect(html).toContain('Provider unavailable')
    expect(html).toContain('provider rejected')
  })

  test('hides positive runtime capability copy in the rail', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      capabilityStatus: 'model-gateway',
      run: null,
    }))

    expect(html).not.toContain('Model gateway')
    expect(html).not.toContain('Real provider execution is available')
    expect(html).not.toContain('capability-rail model-gateway')
  })

  test('renders run-level thinking hint with elapsed seconds while assistant draft is empty', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      locale: 'zh',
      run: {
        id: 'run-thinking',
        threadId: 'thread-a',
        status: 'running',
        model: 'Model gateway',
        context: 'model_gateway',
        createdAt: new Date().toISOString(),
        events: [],
        assistantDraft: { content: '', status: 'streaming' },
      },
    }))

    expect(html).toMatch(/组织回复 0s|梳理线索 0s|核对上下文 0s|提炼重点 0s|推敲答案 0s|收束思路 0s|准备回答 0s|再看一眼 0s/)
    expect(html).not.toContain('模型正在生成回复')
  })

  test('routes run thinking status through incremental typewriter helper', () => {
    const source = readFileSync(resolve(import.meta.dir, 'RunRail.tsx'), 'utf8')

    expect(source).toContain('nextTypewriterFrame')
    expect(source).toContain('prefers-reduced-motion')
  })

  test('renders safe completed thought summary without hidden reasoning content', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      locale: 'zh',
      run: {
        id: 'run-thinking-summary',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Model gateway',
        context: 'model_gateway',
        events: [],
        thinkingSummary: '检查输入并收束答案',
        thinkingDurationSeconds: 12,
      },
    }))

    expect(html).toContain('思考 12s')
    expect(html).toContain('检查输入并收束答案')
    expect(html).not.toContain('raw hidden reasoning')
  })

  test('renders human-first tool event labels without raw tool names or paths', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'running',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-read', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'tool.call.approval_required', label: 'Tool', detail: 'workspace.read waiting for approval', time: 'Now', status: 'running', metadata: { tool_name: 'workspace.read', arguments_summary: { path: 'web/src/App.tsx' } } },
          { id: 'evt-web', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'tool.call.succeeded', label: 'Tool', detail: 'web.fetch completed', time: 'Now', status: 'completed', metadata: { tool_name: 'web.fetch', result_summary: { status_code: 200 } } },
          { id: 'evt-lsp', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'tool.call.succeeded', label: 'Tool', detail: 'lsp.symbols completed', time: 'Now', status: 'completed', metadata: { tool_name: 'lsp.symbols' } },
          { id: 'evt-artifact', runId: 'run-a', threadId: 'thread-a', sequence: 4, type: 'tool.call.succeeded', label: 'Tool', detail: 'artifact.read completed', time: 'Now', status: 'completed', metadata: { tool_name: 'artifact.read' } },
          { id: 'evt-agent', runId: 'run-a', threadId: 'thread-a', sequence: 5, type: 'tool.call.succeeded', label: 'Tool', detail: 'agent.spawn completed', time: 'Now', status: 'completed', metadata: { tool_name: 'agent.spawn' } },
        ],
      },
    }))

    expect(html).toContain('Read project files')
    expect(html).toContain('Visit web page')
    expect(html).toContain('Analyze code')
    expect(html).toContain('Handle artifact')
    expect(html).toContain('Coordinate subtasks')
    expect(html).not.toContain('workspace.read')
    expect(html).not.toContain('web.fetch')
    expect(html).not.toContain('lsp.symbols')
    expect(html).not.toContain('artifact.read')
    expect(html).not.toContain('agent.spawn')
    expect(html).not.toContain('web/src/App.tsx')
    expect(html).not.toContain('Workspace tool · workspace.read waiting for approval')
    expect(html).not.toContain('Web fetch tool · medium risk · public HTTP only')
  })

  test('redacts sensitive tool metadata before rendering timeline rows', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'completed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          {
            id: 'evt-read',
            runId: 'run-a',
            threadId: 'thread-a',
            sequence: 1,
            type: 'tool.call.succeeded',
            label: 'Tool',
            detail: 'workspace.read /Users/xuean/project/.env Authorization Bearer sk-secret-token cookie=session token=hidden api_key=secret secret=value password=pw credential=cred session=sid',
            time: 'Now',
            status: 'completed',
            metadata: {
              tool_name: 'workspace.read',
              arguments_summary: { path: '/Users/xuean/project/.env', authorization: 'Bearer sk-secret-token', cookie: 'session=abc' },
              result_summary: { stdout: 'raw stdout payload', stderr: 'raw stderr payload', raw_body: '<html>secret</html>', status_code: 200 },
            },
          },
        ],
      },
    }))

    expect(html).toContain('Read project files completed')
    expect(html).not.toContain('workspace.read')
    expect(html).toContain('status_code: 200')
    expect(html).not.toContain('/Users/xuean/project')
    expect(html).not.toContain('.env')
    expect(html).not.toContain('Authorization')
    expect(html).not.toContain('sk-secret-token')
    expect(html).not.toContain('session=abc')
    expect(html).not.toContain('token=hidden')
    expect(html).not.toContain('api_key=secret')
    expect(html).not.toContain('secret=value')
    expect(html).not.toContain('password=pw')
    expect(html).not.toContain('credential=cred')
    expect(html).not.toContain('session=sid')
    expect(html).not.toContain('raw stdout payload')
    expect(html).not.toContain('raw stderr payload')
    expect(html).not.toContain('<html>secret</html>')
  })
})

describe('RunRail localized runtime copy', () => {
  test('renders Chinese runtime group and capability copy when locale is zh', () => {
    const html = renderToStaticMarkup(createElement(RunRail, {
      open: true,
      locale: 'zh',
      capabilityStatus: 'provider-unavailable',
      run: {
        id: 'run-a',
        threadId: 'thread-a',
        status: 'failed',
        model: 'Local simulated',
        context: 'local_simulated',
        events: [
          { id: 'evt-run', runId: 'run-a', threadId: 'thread-a', sequence: 1, type: 'run.created', label: 'Run', detail: 'created', time: 'Now', status: 'running' },
          { id: 'evt-worker', runId: 'run-a', threadId: 'thread-a', sequence: 2, type: 'worker.claimed', label: 'Worker', detail: 'claimed', time: 'Now', status: 'running' },
          { id: 'evt-error', runId: 'run-a', threadId: 'thread-a', sequence: 3, type: 'provider.error', label: 'Provider', detail: 'provider failed', time: 'Now', status: 'failed' },
        ],
      },
    }))

    expect(html).toContain('进度')
    expect(html).toContain('最近活动')
    expect(html).toContain('运行失败')
    expect(html).not.toContain('已创建运行')
    expect(html).not.toContain('Worker 已领取任务')
    expect(html).toContain('Provider 不可用')
    expect(html).toContain('Provider 拒绝或未能完成生成')
  })
})
