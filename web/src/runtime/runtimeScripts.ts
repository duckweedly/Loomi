import type { RuntimeEvent, RuntimeScript, RuntimeScriptId, RuntimeScriptStep } from '../domain'

export const runtimeScripts: Record<RuntimeScriptId, RuntimeScript> = {
  success: {
    id: 'success',
    name: '成功剧本',
    terminalStatus: 'completed',
    finalAssistantMessage: '已完成一次模拟执行。',
    steps: [
      { type: 'run.created', label: 'Run', detail: '已创建', status: 'running' },
      { type: 'context.loading', label: 'Context', detail: '加载上下文', status: 'running' },
      { type: 'assistant.thinking', label: 'Thinking', detail: '思考中', status: 'running' },
      { type: 'assistant.drafting', label: 'Drafting', detail: '草拟回复', status: 'running', assistantDelta: '正在整理答案。' },
      { type: 'assistant.message.completed', label: 'Message', detail: '回复完成', status: 'completed' },
      { type: 'run.completed', label: 'Done', detail: '执行完成', status: 'completed' },
    ],
  },
  failure: {
    id: 'failure',
    name: '失败剧本',
    terminalStatus: 'failed',
    steps: [
      { type: 'run.created', label: 'Run', detail: '已创建', status: 'running' },
      { type: 'context.loading', label: 'Context', detail: '加载上下文', status: 'running' },
      { type: 'assistant.thinking', label: 'Thinking', detail: '思考中', status: 'running' },
      { type: 'run.failed', label: 'Failed', detail: '执行失败', status: 'failed' },
    ],
  },
}

export function getRuntimeScript(scriptId: RuntimeScriptId = 'success') {
  return runtimeScripts[scriptId]
}

export function getRuntimeScriptSteps(scriptId: RuntimeScriptId = 'success') {
  return getRuntimeScript(scriptId).steps
}

export function createRuntimeEvent({ threadId, runId, sequence, step, time = 'Now' }: { threadId: string; runId: string; sequence: number; step: RuntimeScriptStep; time?: string }): RuntimeEvent {
  return {
    id: `${runId}-evt-${sequence}`,
    threadId,
    runId,
    type: step.type,
    label: step.label,
    detail: step.detail,
    time,
    status: step.status,
    assistantDelta: step.assistantDelta,
  }
}
