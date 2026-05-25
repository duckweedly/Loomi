import { describe, expect, test } from 'bun:test'
import type { Run } from '../domain'
import { deriveAgentMotionState } from './AgentStateMotion'

const run = (input: Partial<Run>): Run => ({
  id: 'run-test',
  threadId: 'thread-test',
  status: 'running',
  model: 'Claude Sonnet',
  context: '12k / 128k',
  events: [],
  ...input,
})

describe('deriveAgentMotionState', () => {
  test('does not copy and reverse the event list on render', () => {
    const source = Bun.file(new URL('./AgentStateMotion.tsx', import.meta.url)).text()

    return expect(source).resolves.not.toContain('[...run.events].reverse()')
  })

  test('uses idle when no run is selected', () => {
    expect(deriveAgentMotionState(null)).toBe('idle')
  })

  test('uses idle for a completed run that has not emitted events yet', () => {
    expect(deriveAgentMotionState(run({ status: 'completed', events: [] }))).toBe('idle')
  })

  test('uses done and error for terminal run states with execution history', () => {
    expect(deriveAgentMotionState(run({ status: 'completed', events: [{ id: 'evt-done', type: 'run.completed', label: 'Completed', detail: 'Finished', time: 'Now', status: 'completed' }] }))).toBe('done')
    expect(deriveAgentMotionState(run({ status: 'stopped', events: [{ id: 'evt-stop', type: 'run.stopped', label: 'Stopped', detail: 'Stopped', time: 'Now', status: 'stopped' }] }))).toBe('error')
  })

  test('uses the latest running event to choose tool, speaking, or thinking motion', () => {
    expect(deriveAgentMotionState(run({ events: [{ id: 'evt-1', type: 'tool.completed', label: 'Tool', detail: 'Tool finished', time: 'Now', status: 'running' }] }))).toBe('tool')
    expect(deriveAgentMotionState(run({ events: [{ id: 'evt-2', type: 'message.drafting', label: 'Drafting', detail: 'Writing reply', time: 'Now', status: 'running' }] }))).toBe('speaking')
    expect(deriveAgentMotionState(run({ events: [{ id: 'evt-3', type: 'context.loaded', label: 'Context', detail: 'Context ready', time: 'Now', status: 'running' }] }))).toBe('thinking')
  })

  test('uses confirm motion while waiting for tool approval', () => {
    expect(deriveAgentMotionState(run({ status: 'blocked_on_tool_approval', events: [{ id: 'evt-tool', type: 'tool.call.approval_required', label: 'Tool', detail: 'Tool approval required', time: 'Now', status: 'blocked_on_tool_approval' }] }))).toBe('confirm')
  })

  test('ignores completed events when a later active event is still running', () => {
    expect(deriveAgentMotionState(run({
      events: [
        { id: 'evt-1', type: 'tool.completed', label: 'Tool', detail: 'Tool finished', time: '10:25', status: 'completed' },
        { id: 'evt-2', type: 'message.drafting', label: 'Drafting', detail: 'Writing reply', time: 'Now', status: 'running' },
      ],
    }))).toBe('speaking')
  })
})
