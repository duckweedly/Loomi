import { FormEvent, useState } from 'react'
import { Button } from '@lobehub/ui'
import { ArrowUp, ChevronDown, Folder, Mic, Plus } from 'lucide-react'
import type { Message, Run } from '../domain'
import { deriveComposerActions } from '../runtime/composerActions'

type Props = {
  disabled?: boolean
  providerUnavailable?: boolean
  placeholder?: string
  threadSelected: boolean
  run: Run | null
  messages: Message[]
  onSubmit: (content: string) => void
  onStop?: () => void
  onRetry?: () => void
  onRegenerate?: () => void
  attachLabel?: string
  stopLabel?: string
  retryLabel?: string
  regenerateLabel?: string
}

export function Composer({ disabled, providerUnavailable = false, placeholder = 'Message Loomi', threadSelected, run, messages, onSubmit, onStop, onRetry, onRegenerate, attachLabel = 'Attach', stopLabel = 'Stop', retryLabel = 'Retry', regenerateLabel = 'Regenerate' }: Props) {
  const [value, setValue] = useState('')
  const actions = deriveComposerActions({ threadSelected, text: value, run, messages, providerUnavailable })
  const composerDisabled = Boolean(disabled || providerUnavailable)
  const canSubmit = actions.canSend || actions.canContinue

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const content = value.trim()
    if (composerDisabled || !canSubmit || !content) return
    onSubmit(content)
    setValue('')
  }

  return (
    <form className="composer glass-panel" onSubmit={handleSubmit}>
      <textarea
        className="composer-input"
        disabled={composerDisabled}
        onChange={(event) => setValue(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' && !event.shiftKey) handleSubmit(event)
        }}
        placeholder={placeholder}
        rows={2}
        value={value}
      />
      <div className="composer-toolbar">
        <div className="composer-toolbar-left">
          <button className="composer-folder" disabled type="button">
            <Folder size={17} />
            <span>Work in a folder</span>
          </button>
          <button aria-label={attachLabel} className="composer-tool" disabled type="button">
            <Plus size={21} />
          </button>
          <div className="composer-actions">
            {actions.canStop && onStop && <button type="button" className="composer-action" onClick={onStop}>{stopLabel}</button>}
            {actions.canRetry && !composerDisabled && onRetry && <button type="button" className="composer-action" onClick={onRetry}>{retryLabel}</button>}
            {actions.canRegenerate && !composerDisabled && onRegenerate && <button type="button" className="composer-action" onClick={onRegenerate}>{regenerateLabel}</button>}
          </div>
        </div>
        <div className="composer-toolbar-right">
          <button className="composer-model" disabled type="button">
            <span>{run?.model ?? 'gpt-5.5'}</span>
            <ChevronDown size={16} />
          </button>
          <button className="composer-tool" disabled type="button">
            <Mic size={20} />
          </button>
          <Button disabled={composerDisabled || !canSubmit || value.trim().length === 0} htmlType="submit" icon={<ArrowUp size={15} />} type="primary" />
        </div>
      </div>
    </form>
  )
}
