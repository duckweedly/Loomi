import { FormEvent, useState } from 'react'
import { Button } from '@lobehub/ui'
import { ArrowUp, Paperclip } from 'lucide-react'
import type { Message, Run } from '../domain'
import { deriveComposerActions } from '../runtime/composerActions'

type Props = {
  disabled?: boolean
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

export function Composer({ disabled, placeholder = 'Message Loomi', threadSelected, run, messages, onSubmit, onStop, onRetry, onRegenerate, attachLabel = 'Attach', stopLabel = 'Stop', retryLabel = 'Retry', regenerateLabel = 'Regenerate' }: Props) {
  const [value, setValue] = useState('')
  const actions = deriveComposerActions({ threadSelected, text: value, run, messages })
  const canSubmit = actions.canSend || actions.canContinue

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const content = value.trim()
    if (disabled || !canSubmit || !content) return
    onSubmit(content)
    setValue('')
  }

  return (
    <form className="composer glass-panel" onSubmit={handleSubmit}>
      <div className="composer-actions">
        {actions.canStop && onStop && <button type="button" className="composer-action" onClick={onStop}>{stopLabel}</button>}
        {actions.canRetry && onRetry && <button type="button" className="composer-action" onClick={onRetry}>{retryLabel}</button>}
        {actions.canRegenerate && onRegenerate && <button type="button" className="composer-action" onClick={onRegenerate}>{regenerateLabel}</button>}
      </div>
      <Button aria-label={attachLabel} icon={<Paperclip size={15} />} size="small" />
      <textarea
        className="composer-input"
        disabled={disabled}
        onChange={(event) => setValue(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === 'Enter' && !event.shiftKey) handleSubmit(event)
        }}
        placeholder={placeholder}
        rows={1}
        value={value}
      />
      <Button disabled={disabled || !canSubmit || value.trim().length === 0} htmlType="submit" icon={<ArrowUp size={15} />} type="primary" />
    </form>
  )
}
