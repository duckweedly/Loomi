import { describe, expect, test } from 'bun:test'
import type { Message, Run, Thread } from './domain'
import { deriveWorkPlanProjection, safeWorkMetadataPreview } from './workModeProjection'

const workThread: Thread = { id: 'thread-work', title: 'Ship M16', project: 'Loomi', mode: 'work', updatedAt: 'Now', lifecycleStatus: 'active', runStatus: 'running' }
const chatThread: Thread = { ...workThread, id: 'thread-chat', mode: 'chat' }
const messages: Message[] = [{ id: 'msg-1', threadId: workThread.id, role: 'user', content: 'Ship the Work mode foundation', createdAt: 'Now' }]

describe('deriveWorkPlanProjection', () => {
  test('projects a seeded work thread plan from safe event metadata', () => {
    const run: Run = {
      id: 'run-work',
      threadId: workThread.id,
      status: 'running',
      model: 'Model gateway',
      context: 'model_gateway',
      events: [{
        id: 'evt-plan',
        runId: 'run-work',
        threadId: workThread.id,
        sequence: 1,
        type: 'work.plan.updated',
        label: 'Work',
        detail: 'Plan updated',
        time: 'Now',
        status: 'running',
        metadata: {
          work_goal: 'Deliver Work Plan View',
          work_steps: [{ id: 'step-1', title: 'Project plan', status: 'completed' }, { id: 'step-2', title: 'Render view', status: 'running' }],
          work_artifacts: [{ id: 'artifact-1', title: 'M16 plan', type: 'markdown', source_run_id: 'run-work', summary: 'Safe preview', created_at: '2026-05-25' }],
        },
      }],
    }

    const projection = deriveWorkPlanProjection(workThread, messages, run)

    expect(projection?.goal).toBe('Deliver Work Plan View')
    expect(projection?.steps.map((step) => `${step.title}:${step.status}`)).toEqual(['Project plan:completed', 'Render view:running'])
    expect(projection?.artifacts[0]).toMatchObject({ title: 'M16 plan', type: 'markdown', sourceRunId: 'run-work', summary: 'Safe preview' })
    expect(projection?.recentEvents[0].detail).toBe('Plan updated')
  })

  test('does not project chat mode even when events contain work metadata', () => {
    const run: Run = { id: 'run-chat', threadId: chatThread.id, status: 'running', model: 'Mock', context: 'local_simulated', events: [{ id: 'evt-chat', type: 'work.plan.updated', label: 'Work', detail: 'Should not render', time: 'Now', status: 'running' }] }

    expect(deriveWorkPlanProjection(chatThread, [], run)).toBeNull()
  })

  test('event replay drives current progress without a separate queue', () => {
    const initialRun: Run = { id: 'run-work', threadId: workThread.id, status: 'running', model: 'Mock', context: 'local_simulated', events: [] }
    const replayedRun: Run = {
      ...initialRun,
      status: 'completed',
      events: [
        { id: 'evt-1', type: 'work.plan.updated', label: 'Work', detail: 'Step started', time: '10:00', status: 'running', metadata: { work_steps: [{ title: 'Replay event', status: 'running' }] } },
        { id: 'evt-2', type: 'run.completed', label: 'Run', detail: 'Work complete', time: '10:01', status: 'completed' },
      ],
    }

    expect(deriveWorkPlanProjection(workThread, messages, initialRun)?.status).toBe('running')
    const projection = deriveWorkPlanProjection(workThread, messages, replayedRun)
    expect(projection?.status).toBe('completed')
    expect(projection?.statusDetail).toBe('Step started')
    expect(projection?.steps[0].status).toBe('running')
  })

  test('redacts secret-looking metadata and executable hints in allowed artifact values', () => {
    const run: Run = {
      id: 'run-work',
      threadId: workThread.id,
      status: 'completed',
      model: 'Mock',
      context: 'local_simulated',
      events: [{
        id: 'evt-secret',
        type: 'work.artifact.linked',
        label: 'Artifact',
        detail: 'Linked sk-super-secret-token',
        time: 'Now',
        status: 'completed',
        metadata: {
          work_artifacts: [{
            title: 'Report sk-super-secret-token',
            type: 'open /tmp/report.md',
            summary: 'Open browser at https://example.test/private',
            created_at: '/Users/xuean/private.md',
            updated_at: '2026-05-25',
            command: 'open /tmp/private',
            file_path: '/Users/xuean/private.md',
          }],
        },
      }],
    }

    const projection = deriveWorkPlanProjection(workThread, messages, run)
    const serialized = JSON.stringify(projection)

    expect(serialized).not.toContain('sk-super-secret-token')
    expect(serialized).not.toContain('https://example.test/private')
    expect(serialized).not.toContain('open /tmp/private')
    expect(serialized).not.toContain('/Users/xuean/private.md')
    expect(projection?.artifacts[0].title).toBe('[redacted]')
    expect(projection?.artifacts[0].type).toBe('[redacted]')
    expect(projection?.artifacts[0].summary).toBe('[redacted]')
    expect(projection?.artifacts[0].createdAt).toBe('[redacted]')
    expect(projection?.artifacts[0].updatedAt).toBe('2026-05-25')
    expect(safeWorkMetadataPreview({ token: 'sk-super-secret-token', title: 'Safe note', source: 'https://example.test/private', command: 'run this' })).toBe('token: [redacted] · title: Safe note · source: [redacted]')
  })
})
