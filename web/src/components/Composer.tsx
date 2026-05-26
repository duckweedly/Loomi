import { FormEvent, useState } from 'react'
import { Button } from '@lobehub/ui'
import { ArrowUp, FolderOpen } from 'lucide-react'
import type { Message, Run } from '../domain'
import { deriveComposerActions } from '../runtime/composerActions'

type Props = {
  disabled?: boolean
  providerUnavailable?: boolean
  placeholder?: string
  mode?: 'chat' | 'work'
  dataSourceMode?: 'mock' | 'real_api'
  threadSelected: boolean
  run: Run | null
  messages: Message[]
  onSubmit: (content: string) => void
  onStop?: () => void
  onRetry?: () => void
  onRegenerate?: () => void
  onChooseWorkspaceFolder?: () => void
  stopLabel?: string
  retryLabel?: string
  regenerateLabel?: string
  workspaceFolderLabel?: string
  workspaceFolderStatus?: string
}

export function Composer({ disabled, providerUnavailable = false, placeholder, mode = 'chat', threadSelected, run, messages, onSubmit, onStop, onRetry, onRegenerate, onChooseWorkspaceFolder, stopLabel = 'Stop', retryLabel = 'Retry', regenerateLabel = 'Regenerate', workspaceFolderLabel = '选择目录', workspaceFolderStatus }: Props) {
  const [value, setValue] = useState('')
  const actions = deriveComposerActions({ threadSelected, text: value, run, messages, providerUnavailable })
  const composerDisabled = Boolean(disabled || providerUnavailable)
  const canSubmit = actions.canSend || actions.canContinue
  const inputPlaceholder = placeholder ?? (mode === 'work' ? 'Describe the task for Loomi' : 'Message Loomi')

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
        placeholder={inputPlaceholder}
        rows={2}
        value={value}
      />
      <div className="composer-toolbar">
        <div className="composer-toolbar-left">
          {mode === 'work' && onChooseWorkspaceFolder && (
            <button type="button" className="composer-action composer-folder" onClick={onChooseWorkspaceFolder}>
              <FolderOpen size={15} />
              <span>{workspaceFolderLabel}</span>
            </button>
          )}
          {mode === 'work' && workspaceFolderStatus && <span className="composer-folder-status">{workspaceFolderStatus}</span>}
          <div className="composer-actions">
            {actions.canStop && onStop && <button type="button" className="composer-action" onClick={onStop}>{stopLabel}</button>}
            {actions.canRetry && !composerDisabled && onRetry && <button type="button" className="composer-action" onClick={onRetry}>{retryLabel}</button>}
            {actions.canRegenerate && !composerDisabled && onRegenerate && <button type="button" className="composer-action" onClick={onRegenerate}>{regenerateLabel}</button>}
          </div>
        </div>
        <div className="composer-toolbar-right">
          <Button disabled={composerDisabled || !canSubmit || value.trim().length === 0} htmlType="submit" icon={<ArrowUp size={15} />} type="primary" />
        </div>
      </div>
    </form>
  )
}
