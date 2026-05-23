import { loomiHedgehogImage } from '../assets/loomiHedgehog'
import type { Run } from '../domain'

export type AgentMotionState = 'idle' | 'thinking' | 'tool' | 'speaking' | 'confirm' | 'done' | 'error'

const stateLabels: Record<AgentMotionState, string> = {
  idle: 'Idle',
  thinking: 'Thinking',
  tool: 'Tool',
  speaking: 'Speaking',
  confirm: 'Confirm',
  done: 'Done',
  error: 'Error',
}

export function deriveAgentMotionState(run: Run | null): AgentMotionState {
  if (!run) return 'idle'
  if (run.status === 'completed' && run.events.length === 0) return 'idle'
  if (run.status === 'completed') return 'done'
  if (run.status === 'failed' || run.status === 'stopped') return 'error'

  const currentEvent = [...run.events].reverse().find((event) => event.status === 'running') ?? run.events.at(-1)
  const eventText = `${currentEvent?.type ?? ''} ${currentEvent?.label ?? ''}`.toLowerCase()
  if (eventText.includes('tool')) return 'tool'
  if (eventText.includes('message') || eventText.includes('draft') || eventText.includes('reply')) return 'speaking'
  if (eventText.includes('confirm')) return 'confirm'
  return 'thinking'
}

type Props = {
  run: Run | null
  compact?: boolean
}

export function AgentStateMotion({ run, compact = false }: Props) {
  const state = deriveAgentMotionState(run)

  return (
    <div className={compact ? 'agent-motion-card compact' : 'agent-motion-card'} data-state={state} aria-label={`Agent state: ${stateLabels[state]}`}>
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
        <span className="agent-motion-kicker">Agent</span>
        <strong>{stateLabels[state]}</strong>
      </div>
    </div>
  )
}
