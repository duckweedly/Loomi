import { FormEvent, ClipboardEvent, useCallback, useMemo, useRef, useState } from 'react'
import { Button, Select } from 'animal-island-ui'
import { ArrowUp, ChevronRight, FileText, FolderOpen, Image, ListChecks, PackageCheck, Paperclip, Plug, Plus, ShieldCheck, Square, X } from 'lucide-react'
import type { Message, Run } from '../domain'
import { deriveComposerActions } from '../runtime/composerActions'
import { LoomiFloatingMenu, LoomiMenuItem, LoomiMenuSeparator } from './LoomiMenu'

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
  onOpenSkills?: () => void
  onOpenConnectors?: () => void
  onOpenPlugins?: () => void
  stopLabel?: string
  retryLabel?: string
  regenerateLabel?: string
  workspaceFolderLabel?: string
  workspaceFolderStatus?: string
  modelLabel?: string
  attachLabel?: string
  addFilesAndPhotosLabel?: string
  addFolderLabel?: string
  skillsLabel?: string
  connectorsLabel?: string
  addPluginsLabel?: string
  contextMenuLabel?: string
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

function workspaceTitle(status?: string) {
  const label = (status ?? '')
    .replace(/^工作区\s*[·:：]\s*/, '')
    .replace(/^Workspace\s*[·:：]\s*/i, '')
    .trim()
  if (!label || label.includes('/')) return undefined
  return `当前工作区：${label}`
}

type ComposerMenuPosition = {
  top?: number
  bottom?: number
  left: number
  width: number
}

export function getComposerContextMenuPosition(
  rect: Pick<DOMRect, 'bottom' | 'left' | 'right' | 'top'>,
  viewportWidth: number,
  viewportHeight: number,
): ComposerMenuPosition {
  const gap = 8
  const margin = 12
  const width = 228
  const estimatedHeight = 248
  const maxLeft = Math.max(margin, viewportWidth - width - margin)
  const left = Math.min(Math.max(Math.round(rect.left), margin), maxLeft)

  if (rect.top - margin < estimatedHeight) {
    return {
      top: Math.min(Math.round(rect.bottom + gap), Math.max(margin, viewportHeight - estimatedHeight - margin)),
      left,
      width,
    }
  }

  return {
    bottom: Math.max(margin, Math.round(viewportHeight - rect.top + gap)),
    left,
    width,
  }
}

export function Composer({ disabled, providerUnavailable = false, placeholder, threadSelected, run, messages, modelOptions = [], onSubmit, onStop, onChooseWorkspaceFolder, onOpenSkills, onOpenConnectors, onOpenPlugins, stopLabel = 'Stop', workspaceFolderLabel = '选择工作区', workspaceFolderStatus, modelLabel = 'Model', attachLabel = 'Attach', addFilesAndPhotosLabel = 'Add files or photos', addFolderLabel = 'Add folder', skillsLabel = 'Skills', connectorsLabel = 'Connectors', addPluginsLabel = 'Add plugins...', contextMenuLabel = 'Add context', pasteImageLabel = 'Paste image', attachmentPendingLabel = 'queued for this message', modelUnavailableLabel = 'No model' }: Props) {
  const [value, setValue] = useState('')
  const [attachments, setAttachments] = useState<ComposerAttachment[]>([])
  const [selectedModelKey, setSelectedModelKey] = useState('')
  const [contextMenuOpen, setContextMenuOpen] = useState(false)
  const [contextMenuPosition, setContextMenuPosition] = useState<ComposerMenuPosition | null>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  const actions = deriveComposerActions({ threadSelected, text: value, run, messages, providerUnavailable })
  const composerDisabled = Boolean(disabled || providerUnavailable)
  const hasAttachments = attachments.length > 0
  const canSubmit = actions.canSend || actions.canContinue || (threadSelected && hasAttachments && !actions.canStop && !providerUnavailable)
  const showStopButton = Boolean(actions.canStop && onStop)
  const inputPlaceholder = placeholder ?? 'Message Loomi'
  const workspaceContextLabel = workspaceFolderStatus || workspaceFolderLabel
  const workspaceContextState = workspaceFolderStatus ? 'is-selected' : 'is-empty'
  const selectedModel = useMemo(
    () => modelOptions.find((option) => option.key === selectedModelKey) ?? modelOptions[0],
    [modelOptions, selectedModelKey],
  )
  const modelSelectOptions = modelOptions.length > 0
    ? modelOptions.map((option) => ({ key: option.key, label: option.label }))
    : [{ key: '', label: modelUnavailableLabel }]

  const closeContextMenu = useCallback(() => {
    setContextMenuOpen(false)
    setContextMenuPosition(null)
  }, [])

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
    requestAnimationFrame(() => inputRef.current?.focus())
  }

  return (
    <form className="composer glass-panel animal-command-bar" onSubmit={handleSubmit}>
      <textarea
        className="composer-input"
        disabled={composerDisabled}
        ref={inputRef}
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
            aria-label={attachLabel}
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
          <div className="composer-context-menu-wrap">
            <button
              type="button"
              className="composer-context-trigger"
              data-loomi-menu-trigger="composer-context"
              aria-expanded={contextMenuOpen}
              aria-label={contextMenuLabel}
              onClick={(event) => {
                const rect = event.currentTarget.getBoundingClientRect()
                setContextMenuOpen((open) => {
                  const nextOpen = !open
                  setContextMenuPosition(nextOpen ? getComposerContextMenuPosition(rect, window.innerWidth, window.innerHeight) : null)
                  return nextOpen
                })
              }}
            >
              <Plus size={19} />
            </button>
            {contextMenuOpen && (
              <LoomiFloatingMenu
                open
                className="composer-context-menu"
                ignoreSelector="[data-loomi-menu-trigger='composer-context']"
                onClose={closeContextMenu}
                style={contextMenuPosition ?? undefined}
              >
                <LoomiMenuItem onClick={() => {
                  closeContextMenu()
                  fileInputRef.current?.click()
                }}>
                  <Paperclip size={18} />
                  <span>{addFilesAndPhotosLabel}</span>
                </LoomiMenuItem>
                <LoomiMenuItem disabled={!onChooseWorkspaceFolder} onClick={() => {
                  closeContextMenu()
                  onChooseWorkspaceFolder?.()
                }}>
                  <FolderOpen size={18} />
                  <span>{addFolderLabel}</span>
                </LoomiMenuItem>
                <LoomiMenuSeparator />
                <LoomiMenuItem disabled={!onOpenSkills} onClick={() => {
                  closeContextMenu()
                  onOpenSkills?.()
                }}>
                  <ListChecks size={18} />
                  <span>{skillsLabel}</span>
                  <ChevronRight className="composer-context-chevron" size={17} />
                </LoomiMenuItem>
                <LoomiMenuItem disabled={!onOpenConnectors} onClick={() => {
                  closeContextMenu()
                  onOpenConnectors?.()
                }}>
                  <PackageCheck size={18} />
                  <span>{connectorsLabel}</span>
                  <ChevronRight className="composer-context-chevron" size={17} />
                </LoomiMenuItem>
                <LoomiMenuSeparator />
                <LoomiMenuItem disabled={!onOpenPlugins} onClick={() => {
                  closeContextMenu()
                  onOpenPlugins?.()
                }}>
                  <Plug size={18} />
                  <span>{addPluginsLabel}</span>
                </LoomiMenuItem>
              </LoomiFloatingMenu>
            )}
          </div>
          <button
            type="button"
            className={`composer-context-status ${workspaceContextState}`}
            aria-label={`业务上下文：${workspaceContextLabel}`}
            title={workspaceTitle(workspaceFolderStatus)}
            onClick={onChooseWorkspaceFolder}
          >
            <ShieldCheck size={14} />
            <span>{workspaceContextLabel}</span>
          </button>
        </div>
        <div className="composer-toolbar-right">
          <label className={`composer-model-select${modelOptions.length === 0 ? ' disabled' : ''}`}>
            <span>{modelLabel}</span>
            <Select disabled={modelOptions.length === 0} options={modelSelectOptions} placeholder={modelUnavailableLabel} value={selectedModel?.key ?? ''} onChange={setSelectedModelKey} />
          </label>
          <Button
            aria-label={showStopButton ? stopLabel : 'Send message'}
            className={`animal-send-button${showStopButton ? ' is-stopping' : ''}`}
            disabled={!showStopButton && (composerDisabled || !canSubmit || (value.trim().length === 0 && !hasAttachments))}
            htmlType={showStopButton ? 'button' : 'submit'}
            icon={showStopButton ? <Square size={12} fill="currentColor" /> : <ArrowUp size={15} />}
            onClick={showStopButton ? onStop : undefined}
            type="primary"
          />
        </div>
      </div>
    </form>
  )
}
