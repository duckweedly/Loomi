import { FormEvent, ClipboardEvent, useMemo, useRef, useState } from 'react'
import { Button } from '@lobehub/ui'
import { ArrowUp, FileText, FolderOpen, Image, Paperclip, X } from 'lucide-react'
import type { Message, Run } from '../domain'
import { deriveComposerActions } from '../runtime/composerActions'

export type ComposerModelOption = {
  key: string
  label: string
  providerId: string
  model: string
}

export type ComposerAttachment = {
  id: string
  name: string
  type: string
  size: number
  kind: 'file' | 'image'
}

type Props = {
  disabled?: boolean
  providerUnavailable?: boolean
  placeholder?: string
  mode?: 'chat' | 'work'
  dataSourceMode?: 'mock' | 'real_api'
  threadSelected: boolean
  run: Run | null
  messages: Message[]
  modelOptions?: ComposerModelOption[]
  onSubmit: (content: string, options?: { providerId?: string; model?: string; attachments?: ComposerAttachment[] }) => void
  onStop?: () => void
  onRetry?: () => void
  onRegenerate?: () => void
  onChooseWorkspaceFolder?: () => void
  stopLabel?: string
  retryLabel?: string
  regenerateLabel?: string
  workspaceFolderLabel?: string
  workspaceFolderStatus?: string
  modelLabel?: string
  attachLabel?: string
  pasteImageLabel?: string
  attachmentPendingLabel?: string
  modelUnavailableLabel?: string
}

function createAttachmentId() {
  return `attachment-${Date.now()}-${Math.random().toString(36).slice(2)}`
}

function formatAttachmentSize(size: number) {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`
  return `${(size / 1024 / 1024).toFixed(1)} MB`
}

export function Composer({ disabled, providerUnavailable = false, placeholder, mode = 'chat', threadSelected, run, messages, modelOptions = [], onSubmit, onStop, onChooseWorkspaceFolder, stopLabel = 'Stop', workspaceFolderLabel = '选择目录', workspaceFolderStatus, modelLabel = 'Model', attachLabel = 'Attach', pasteImageLabel = 'Paste image', attachmentPendingLabel = 'queued for this message', modelUnavailableLabel = 'No model' }: Props) {
  const [value, setValue] = useState('')
  const [attachments, setAttachments] = useState<ComposerAttachment[]>([])
  const [selectedModelKey, setSelectedModelKey] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const actions = deriveComposerActions({ threadSelected, text: value, run, messages, providerUnavailable })
  const composerDisabled = Boolean(disabled || providerUnavailable)
  const hasAttachments = attachments.length > 0
  const canSubmit = actions.canSend || actions.canContinue || (threadSelected && hasAttachments && !actions.canStop && !providerUnavailable)
  const inputPlaceholder = placeholder ?? (mode === 'work' ? 'Describe the task for Loomi' : 'Message Loomi')
  const selectedModel = useMemo(
    () => modelOptions.find((option) => option.key === selectedModelKey) ?? modelOptions[0],
    [modelOptions, selectedModelKey],
  )

  function addFiles(files: FileList | File[]) {
    const next = Array.from(files).map((file) => ({
      id: createAttachmentId(),
      name: file.name || (file.type.startsWith('image/') ? pasteImageLabel : attachLabel),
      type: file.type || 'application/octet-stream',
      size: file.size,
      kind: file.type.startsWith('image/') ? 'image' as const : 'file' as const,
    }))
    if (next.length) setAttachments((current) => [...current, ...next])
  }

  function handlePaste(event: ClipboardEvent<HTMLTextAreaElement>) {
    const files = Array.from(event.clipboardData.files).filter((file) => file.type.startsWith('image/'))
    if (files.length) addFiles(files)
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const content = value.trim()
    if (composerDisabled || !canSubmit || (!content && !hasAttachments)) return
    onSubmit(content || attachments.map((attachment) => attachment.name).join('\n'), selectedModel ? { providerId: selectedModel.providerId, model: selectedModel.model, attachments } : { attachments })
    setValue('')
    setAttachments([])
  }

  return (
    <form className="composer glass-panel" onSubmit={handleSubmit}>
      <textarea
        className="composer-input"
        disabled={composerDisabled}
        onChange={(event) => setValue(event.target.value)}
        onPaste={handlePaste}
        onKeyDown={(event) => {
          if (event.key === 'Enter' && !event.shiftKey) handleSubmit(event)
        }}
        placeholder={inputPlaceholder}
        rows={2}
        value={value}
      />
      {attachments.length > 0 && (
        <div className="composer-attachments" aria-label={attachLabel}>
          {attachments.map((attachment) => (
            <span className="composer-attachment-chip" key={attachment.id}>
              {attachment.kind === 'image' ? <Image size={14} /> : <FileText size={14} />}
              <span>{attachment.name}</span>
              <small>{formatAttachmentSize(attachment.size)} · {attachmentPendingLabel}</small>
              <button type="button" aria-label="Remove attachment" onClick={() => setAttachments((current) => current.filter((item) => item.id !== attachment.id))}>
                <X size={12} />
              </button>
            </span>
          ))}
        </div>
      )}
      <div className="composer-toolbar">
        <div className="composer-toolbar-left">
          <input
            ref={fileInputRef}
            className="composer-file-input"
            type="file"
            multiple
            accept="image/*,.pdf,.txt,.md,.csv,.json,.doc,.docx,.xls,.xlsx,.ppt,.pptx"
            onChange={(event) => {
              if (event.currentTarget.files) addFiles(event.currentTarget.files)
              event.currentTarget.value = ''
            }}
          />
          <button type="button" className="composer-icon-action" aria-label={attachLabel} onClick={() => fileInputRef.current?.click()}>
            <Paperclip size={16} />
          </button>
          {mode === 'work' && onChooseWorkspaceFolder && (
            <button type="button" className="composer-action composer-folder" onClick={onChooseWorkspaceFolder}>
              <FolderOpen size={15} />
              <span>{workspaceFolderLabel}</span>
            </button>
          )}
          {mode === 'work' && workspaceFolderStatus && <span className="composer-folder-status">{workspaceFolderStatus}</span>}
          <div className="composer-actions">
            {actions.canStop && onStop && <button type="button" className="composer-action" onClick={onStop}>{stopLabel}</button>}
          </div>
        </div>
        <div className="composer-toolbar-right">
          <label className={`composer-model-select${modelOptions.length === 0 ? ' disabled' : ''}`}>
            <span>{modelLabel}</span>
            {modelOptions.length > 0 ? (
              <select value={selectedModel?.key ?? ''} onChange={(event) => setSelectedModelKey(event.target.value)}>
                {modelOptions.map((option) => <option key={option.key} value={option.key}>{option.label}</option>)}
              </select>
            ) : (
              <select disabled value="">
                <option>{modelUnavailableLabel}</option>
              </select>
            )}
          </label>
          <Button aria-label="Send message" disabled={composerDisabled || !canSubmit || (value.trim().length === 0 && !hasAttachments)} htmlType="submit" icon={<ArrowUp size={15} />} type="primary" />
        </div>
      </div>
    </form>
  )
}
