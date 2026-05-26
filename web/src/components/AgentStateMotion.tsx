import { loomiHedgehogImage } from '../assets/loomiHedgehog'
import type { Run } from '../domain'
import type { Locale } from '../i18n'

export type AgentMotionState = 'idle' | 'thinking' | 'tool' | 'speaking' | 'confirm' | 'done' | 'error'

const stateLabels: Record<Locale, Record<AgentMotionState, string>> = {
  zh: {
    idle: '空闲',
    thinking: '思考中',
    tool: '使用工具',
    speaking: '回复中',
    confirm: '等待确认',
    done: '完成',
    error: '出错',
  },
  en: {
    idle: 'Idle',
    thinking: 'Thinking',
    tool: 'Tool',
    speaking: 'Speaking',
    confirm: 'Confirm',
    done: 'Done',
    error: 'Error',
  },
}

export function deriveAgentMotionState(run: Run | null): AgentMotionState {
  if (!run) return 'idle'
  if (run.status === 'completed' && run.events.length === 0) return 'idle'
  if (run.status === 'completed') return 'done'
  if (run.status === 'failed' || run.status === 'stopped') return 'error'
  if (run.status === 'blocked_on_tool_approval') return 'confirm'
  if (run.status === 'queued' || run.status === 'recovering' || run.status === 'stopping') return 'thinking'
  if (run.assistantDraft?.status === 'streaming') return 'speaking'
  if (run.assistantDraft?.status === 'failed' || run.assistantDraft?.status === 'stopped') return 'error'
  if (run.assistantDraft?.status === 'queued' || run.assistantDraft?.status === 'recovering' || run.assistantDraft?.status === 'stopping') return 'thinking'

  let currentEvent = run.events.at(-1)
  for (let index = run.events.length - 1; index >= 0; index -= 1) {
    if (run.events[index].status === 'running') {
      currentEvent = run.events[index]
      break
    }
  }
  const eventText = `${currentEvent?.type ?? ''} ${currentEvent?.label ?? ''}`.toLowerCase()
  if (eventText.includes('tool')) return 'tool'
  if (eventText.includes('model.delta') || eventText.includes('message') || eventText.includes('draft') || eventText.includes('reply')) return 'speaking'
  if (eventText.includes('confirm')) return 'confirm'
  return 'thinking'
}

type Props = {
  run: Run | null
  compact?: boolean
  locale?: Locale
}

export function AgentStateMotion({ run, compact = false, locale = 'en' }: Props) {
  const state = deriveAgentMotionState(run)
  const labels = stateLabels[locale]

  return (
    <div className={compact ? 'agent-motion-card compact' : 'agent-motion-card'} data-state={state} aria-label={`${locale === 'zh' ? 'Agent 状态' : 'Agent state'}: ${labels[state]}`}>
      <div className="agent-motion-stage">
        <div className="agent-motion-ring" />
        <div className="agent-motion-shadow" />
        <div className="agent-motion-actor">
          <img className="agent-motion-hog" src={loomiHedgehogImage} alt="" aria-hidden="true" />
        </div>
        <div className="agent-motion-blink" />
        <div className="agent-motion-overlay">
          <div className="agent-motion-dots"><b /><b /><b /></div>
          <div className="agent-motion-wave"><i /><i /><i /><i /></div>
          <div className="agent-motion-question">?</div>
        </div>
      </div>
      <div className="agent-motion-meta">
        <span className="agent-motion-kicker">{locale === 'zh' ? 'AGENT' : 'Agent'}</span>
        <strong>{labels[state]}</strong>
      </div>
    </div>
  )
}
