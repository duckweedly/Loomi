import type { RuntimeEvent, RuntimeScript, RuntimeScriptId, RuntimeScriptStep } from '../domain'

export const runtimeScripts: Record<RuntimeScriptId, RuntimeScript> = {
  success: {
    id: 'success',
    name: '成功剧本',
    terminalStatus: 'completed',
    finalAssistantMessage: '已完成一次模拟执行。',
    steps: [
      { type: 'run.created', label: 'Run', detail: '已创建', status: 'running' },
      { type: 'run.queued', label: 'Queue', detail: '已排队', status: 'queued', group: 'run-lifecycle', severity: 'info' },
      { type: 'job.claimed', label: 'Worker', detail: 'Worker 已领取', status: 'running', group: 'worker-job', severity: 'info' },
      { type: 'pipeline.step.started', label: 'Pipeline', detail: '开始执行 runtime', status: 'running', group: 'worker-job', severity: 'progress' },
      { type: 'pipeline.step.completed', label: 'Pipeline', detail: '完成执行 runtime', status: 'running', group: 'worker-job', severity: 'info' },
      { type: 'context.loading', label: 'Context', detail: '加载上下文', status: 'running' },
      { type: 'assistant.thinking', label: 'Thinking', detail: '思考中', status: 'running' },
      { type: 'assistant.drafting', label: 'Drafting', detail: '草拟回复', status: 'running', assistantDelta: '正在整理答案。' },
      { type: 'assistant.message.completed', label: 'Message', detail: '回复完成', status: 'running' },
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
  'model-stream': {
    id: 'model-stream',
    name: '模型流剧本',
    terminalStatus: 'completed',
    finalAssistantMessage: '模型流式回复完成。',
    steps: [
      { type: 'run.created', label: 'Run', detail: '已创建', status: 'running', group: 'run-lifecycle', severity: 'info' },
      { type: 'job.queued', label: 'Queue', detail: '已排队', status: 'running', group: 'worker-job', severity: 'info' },
      { type: 'worker.claimed', label: 'Worker', detail: 'Worker 已领取', status: 'running', group: 'worker-job', severity: 'info' },
      { type: 'job.retrying', label: 'Retry', detail: '重试中', status: 'retrying', group: 'worker-job', severity: 'warning' },
      { type: 'model.delta', label: 'Model', detail: '模型输出片段', status: 'running', group: 'model-stream', severity: 'progress', assistantDelta: '模型' },
      { type: 'model.delta', label: 'Model', detail: '模型继续输出', status: 'running', group: 'model-stream', severity: 'progress', assistantDelta: '回复' },
      { type: 'model.final', label: 'Model', detail: '模型输出完成', status: 'running', group: 'model-stream', severity: 'info', usage: { inputTokens: 11, outputTokens: 22 } },
      { type: 'run.completed', label: 'Done', detail: '执行完成', status: 'completed', group: 'run-lifecycle', severity: 'info' },
    ],
  },
  'model-error': {
    id: 'model-error',
    name: '模型错误剧本',
    terminalStatus: 'failed',
    steps: [
      { type: 'run.created', label: 'Run', detail: '已创建', status: 'running', group: 'run-lifecycle', severity: 'info' },
      { type: 'model.delta', label: 'Model', detail: '模型输出片段', status: 'running', group: 'model-stream', severity: 'progress', assistantDelta: '部分回复' },
      { type: 'provider.error', label: 'Provider error', detail: 'Provider 不可用', status: 'running', group: 'error', severity: 'error' },
      { type: 'model.error', label: 'Model error', detail: '模型输出失败', status: 'running', group: 'error', severity: 'error' },
      { type: 'run.failed', label: 'Failed', detail: '执行失败', status: 'failed', group: 'error', severity: 'error' },
    ],
  },
  stopped: {
    id: 'stopped',
    name: '停止剧本',
    terminalStatus: 'stopped',
    steps: [
      { type: 'run.created', label: 'Run', detail: '已创建', status: 'running', group: 'run-lifecycle', severity: 'info' },
      { type: 'model.delta', label: 'Model', detail: '模型输出片段', status: 'running', group: 'model-stream', severity: 'progress', assistantDelta: '停止前片段' },
      { type: 'run.cancelled', label: 'Cancelled', detail: '用户取消', status: 'cancelled', group: 'run-lifecycle', severity: 'warning' },
      { type: 'run.stopped', label: 'Stopped', detail: '用户已停止', status: 'stopped', group: 'run-lifecycle', severity: 'warning' },
    ],
  },
  replayed: {
    id: 'replayed',
    name: '重放剧本',
    terminalStatus: 'completed',
    finalAssistantMessage: '重放回复完成。',
    steps: [
      { type: 'run.created', label: 'Run', detail: '已创建', status: 'running', group: 'run-lifecycle', severity: 'info' },
      { type: 'model.delta', label: 'Model', detail: '重放片段', status: 'running', group: 'model-stream', severity: 'progress', assistantDelta: '重放' },
      { type: 'model.delta', label: 'Model', detail: '重复重放片段', status: 'running', group: 'model-stream', severity: 'progress', assistantDelta: '重放' },
      { type: 'model.final', label: 'Model', detail: '重放完成', status: 'running', group: 'model-stream', severity: 'info' },
      { type: 'run.completed', label: 'Done', detail: '执行完成', status: 'completed', group: 'run-lifecycle', severity: 'info' },
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
    group: step.group,
    severity: step.severity,
    usage: step.usage,
    assistantDelta: step.assistantDelta,
  }
}
