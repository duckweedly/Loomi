import { ArrowLeft, Check, ChevronDown, ChevronRight, Minus, Plus, RefreshCw, Search, Settings as SettingsIcon, X } from 'lucide-react'
import { useState, type ReactNode } from 'react'
import type { BackendCapabilityState, InstalledSkill, LocalProviderDetection, MCPServerConfigInput, MCPServerStatus, MemoryAuditItem, MemoryEntry, MemoryErrorEvent, MemoryFilters, MemoryImpressionSnapshot, MemoryOverviewSnapshot, MemoryProviderStatus, MemoryProviderUpdate, MemoryWriteProposal, Persona, ProviderCapability, Thread, ToolCatalogItem, WebSearchConfig, WorkspaceRootConfig } from '../domain'
import type { ProviderCheckResult, ProviderSaveResult } from '../state'
import type { ProviderDraftSettings } from '../useWorkspaceShellState'
import type { Locale } from '../i18n'
import { getDictionary } from '../i18n'
import { MemoryPanel } from './MemoryPanel'
import { getSettingsCategoriesByGroup, getSettingsCategory, settingsCategoryGroups, type SettingsCategory, type SettingsCategoryGroup, type SettingsCategoryId, type SettingRowStatus } from './settingsCatalog'

type Props = {
  locale: Locale
  selectedCategoryId: SettingsCategoryId
  defaultWorkspaceMode: Thread['mode']
  theme: 'dark' | 'light'
  backendCapability: BackendCapabilityState
  providerCapabilities: ProviderCapability[]
  workspaceRootConfig?: WorkspaceRootConfig | null
  workspaceRootSaveResult?: ProviderSaveResult
  personas?: Persona[]
  installedSkills?: InstalledSkill[]
  skillsLoading?: boolean
  skillsError?: string | null
  toolCatalog: ToolCatalogItem[]
  webSearchConfig?: WebSearchConfig | null
  webSearchSaveResult?: ProviderSaveResult
  mcpServers?: MCPServerStatus[]
  mcpActionResult?: ProviderSaveResult
  localProviderDetections: LocalProviderDetection[]
  localProviderDetectionError?: string | null
  memoryEntries: MemoryEntry[]
  memoryQuery: string
  memoryFilters: MemoryFilters
  memoryLoading: boolean
  memoryError?: string | null
  memoryDetail?: MemoryEntry | null
  memoryDetailLoading?: boolean
  memoryDetailError?: string | null
  memoryAuditItems: MemoryAuditItem[]
  memoryAuditLoading: boolean
  memoryAuditError?: string | null
  memoryWriteProposals?: MemoryWriteProposal[]
  memoryProposalsLoading?: boolean
  memoryProposalsError?: string | null
  memoryProviderStatus?: MemoryProviderStatus | null
  memoryErrors?: MemoryErrorEvent[]
  memoryProviderSaveResult?: ProviderSaveResult
  memoryOverviewSnapshot?: MemoryOverviewSnapshot | null
  memoryImpressionSnapshot?: MemoryImpressionSnapshot | null
  memorySnapshotLoading?: boolean
  pendingDeleteMemoryEntry?: MemoryEntry | null
  providerCheckResults: Record<string, ProviderCheckResult>
  providerSaveResult: ProviderSaveResult
  providerDraftSettings: ProviderDraftSettings
  onSelectLocale: (locale: Locale) => void
  onSelectCategory: (categoryId: SettingsCategoryId) => void
  onSelectDefaultWorkspaceMode: (mode: Thread['mode']) => void
  onSelectTheme: (theme: 'dark' | 'light') => void
  onProviderDraftSettingsChange: (settings: ProviderDraftSettings) => void
  onSaveProvider: (settings: ProviderDraftSettings) => void
  onSaveWebSearchKeys?: (input: { tavilyApiKey?: string; braveApiKey?: string }) => void
  onRefreshMemoryProviderStatus?: () => void
  onUpdateMemoryProvider?: (input: MemoryProviderUpdate) => void
  onDetectNowledgeMemoryProvider?: () => Promise<{ detected: boolean; baseUrl?: string; message: string }>
  onDetectOpenVikingMemoryProvider?: () => Promise<{ detected: boolean; baseUrl?: string; message: string }>
  onRebuildMemoryOverviewSnapshot?: () => void
  onRebuildMemoryImpressionSnapshot?: () => void
  onGetMemoryContent?: (uri: string, layer?: 'overview' | 'read') => Promise<string>
  onSaveMCPServer?: (input: MCPServerConfigInput) => void
  onDeleteMCPServer?: (slug: string) => void
  onDiscoverMCPServer?: (slug: string) => void
  onCheckProvider: (providerId: string) => void
  onDetectLocalProviders: () => void
  onEnableLocalProvider: (providerId: string) => void
  onDisableLocalProvider: (providerId: string) => void
  onMemoryQueryChange: (query: string) => void
  onMemoryFiltersChange: (filters: MemoryFilters) => void
  onOpenMemoryDetail: (entry: MemoryEntry) => void
  onCloseMemoryDetail: () => void
  onRequestDeleteMemoryEntry: (entry: MemoryEntry) => void
  onCancelDeleteMemoryEntry: () => void
  onCreateMemoryEntry?: (input: { title: string; content: string; scopeType?: 'user' | 'thread'; scopeId?: string }) => void
  onConfirmDeleteMemoryEntry: (entry: MemoryEntry) => void
  onApproveMemoryProposal?: (proposal: MemoryWriteProposal) => void
  onUpdateMemoryProposal?: (proposal: MemoryWriteProposal, input: { title: string; summary: string }) => void
  onDenyMemoryProposal?: (proposal: MemoryWriteProposal) => void
  onBack: () => void
}

const categoryGroups = Object.keys(settingsCategoryGroups) as SettingsCategoryGroup[]

function statusLabel(status: SettingRowStatus, t: ReturnType<typeof getDictionary>['settings']) {
  if (status === 'working') return t.working
  if (status === 'read_only') return t.readOnly
  if (status === 'disabled') return t.disabled
  return t.previewOnly
}

function capabilityLabel(status: BackendCapabilityState, t: ReturnType<typeof getDictionary>['settings']) {
  if (status === 'available') return t.available
  if (status === 'configured') return 'Configured'
  if (status === 'reachable') return 'Reachable'
  if (status === 'completion-ok') return 'Completion ok'
  if (status === 'completion-failed') return 'Completion failed'
  if (status === 'misconfigured') return t.misconfigured
  return t.unavailable
}

function localProviderStatusLabel(status: LocalProviderDetection['status'], t: ReturnType<typeof getDictionary>['settings']) {
  if (status === 'available') return t.available
  if (status === 'disabled') return t.disabled
  return t.unavailable
}

function categoryStatusLabel(category: SettingsCategory, t: ReturnType<typeof getDictionary>['settings']) {
  if (category.status === 'working') return t.working
  if (category.status === 'read_only') return t.readOnly
  if (category.status === 'mixed') return t.mixed
  return t.previewOnly
}

function SettingRow({ label, helperText, status, control, t }: { label: string; helperText: string; status: SettingRowStatus; control: ReactNode; t: ReturnType<typeof getDictionary>['settings'] }) {
  return (
    <div className="setting-row">
      <div className="setting-row-copy">
        <div className="setting-row-title">
          <span>{label}</span>
          <span className={`setting-status-badge ${status}`}>{statusLabel(status, t)}</span>
        </div>
        <p>{helperText}</p>
      </div>
      <div className="setting-row-control">{control}</div>
    </div>
  )
}

function memoryProviderErrorSummary(item: MemoryErrorEvent, locale: Locale) {
  const details = [item.provider, item.state, item.message]
  if (item.eventType) details.push(item.eventType)
  if (item.runId) details.push(`${locale === 'zh' ? '运行' : 'run'} ${item.runId}`)
  return details.filter(Boolean).join(' · ')
}

function MemoryProviderFoundationPanel({ status, errors = [], saveResult, locale, onRefresh, onUpdate, onDetectNowledge, onDetectOpenViking }: { status?: MemoryProviderStatus | null; errors?: MemoryErrorEvent[]; saveResult?: ProviderSaveResult; locale: Locale; onRefresh?: () => void; onUpdate?: (input: MemoryProviderUpdate) => void; onDetectNowledge?: () => Promise<{ detected: boolean; baseUrl?: string; message: string }>; onDetectOpenViking?: () => Promise<{ detected: boolean; baseUrl?: string; message: string }> }) {
  const [detectMessage, setDetectMessage] = useState('')
  const [configOpen, setConfigOpen] = useState(false)
  const copy = locale === 'zh'
    ? {
        title: '记忆服务',
        subtitle: '运行前读取安全记忆；每轮后整理会生成待审批提案。',
        enabled: '启用记忆',
        commit: '每轮后自动整理',
        local: '本地',
        nowledge: 'Nowledge',
        openviking: 'OpenViking',
        localDescription: '使用 Loomi 本地已批准记忆。',
        nowledgeDescription: '语义召回与本地 Nowledge 服务。',
        openvikingDescription: '自动整理的记忆，由召回服务按相关性提供。',
        refresh: '刷新',
        state: '状态',
        configured: '配置',
        configuredYes: '已配置',
        configuredNo: '未配置',
        baseUrl: '服务地址',
        rootKey: 'Root API Key',
        apiKey: 'API Key',
        timeout: '超时 ms',
        embedding: 'Embedding 模型',
        vlm: 'VLM 模型',
        rerank: 'Rerank 模型',
        selector: '选择器',
        providerName: 'Provider',
        model: '模型',
        apiBase: 'API Base',
        dimension: '维度',
        stored: '已保存',
        recentErrors: '近期异常',
        detect: '检测本地实例',
        configure: '配置',
        configTitle: '记忆服务配置',
        close: '关闭',
      }
    : {
        title: 'Memory Service',
        subtitle: 'Reads safe memory before runs; post-run organization creates approval-gated proposals.',
        enabled: 'Enable memory',
        commit: 'Organize after each run',
        local: 'Local',
        nowledge: 'Nowledge',
        openviking: 'OpenViking',
        localDescription: 'Use Loomi approved local memory.',
        nowledgeDescription: 'Semantic recall through the local Nowledge service.',
        openvikingDescription: 'Organized memory provided by relevance from the recall service.',
        refresh: 'Refresh',
        state: 'State',
        configured: 'Config',
        configuredYes: 'Configured',
        configuredNo: 'Unconfigured',
        baseUrl: 'Base URL',
        rootKey: 'Root API Key',
        apiKey: 'API Key',
        timeout: 'Timeout ms',
        embedding: 'Embedding model',
        vlm: 'VLM model',
        rerank: 'Rerank model',
        selector: 'Selector',
        providerName: 'Provider',
        model: 'Model',
        apiBase: 'API Base',
        dimension: 'Dimension',
        stored: 'Stored',
        recentErrors: 'Recent Errors',
        detect: 'Detect local instance',
        configure: 'Configure',
        configTitle: 'Memory Service Configuration',
        close: 'Close',
      }
  const current = status ?? {
    enabled: true,
    provider: 'local',
    label: copy.local,
    state: 'available',
    configured: true,
    commitAfterRun: false,
    diagnostic: { code: 'loading', message: locale === 'zh' ? '正在读取记忆服务状态。' : 'Loading memory provider status.' },
  } satisfies MemoryProviderStatus
  const update = (patch: Partial<MemoryProviderUpdate>) => onUpdate?.({
    enabled: current.enabled,
    provider: current.provider,
    commitAfterRun: current.commitAfterRun,
    openviking: current.openviking,
    nowledge: current.nowledge,
    ...patch,
  })
  const openviking = current.openviking ?? {}
  const nowledge = current.nowledge ?? {}
  const selectedProvider = current.provider === 'openviking' || current.provider === 'nowledge' ? current.provider : 'local'
  const showProviderConfig = selectedProvider === 'nowledge' || selectedProvider === 'openviking'
  const updateOpenViking = (patch: NonNullable<MemoryProviderUpdate['openviking']>) => update({ openviking: { ...openviking, ...patch } })
  const updateNowledge = (patch: NonNullable<MemoryProviderUpdate['nowledge']>) => update({ nowledge: { ...nowledge, ...patch } })
  const providerOptions = [
    { value: 'local', label: copy.local, description: copy.localDescription },
    { value: 'nowledge', label: copy.nowledge, description: copy.nowledgeDescription },
    { value: 'openviking', label: copy.openviking, description: copy.openvikingDescription },
  ] as const
  const detectNowledge = async () => {
    if (!onDetectNowledge) return
    try {
      const result = await onDetectNowledge()
      setDetectMessage(result.message)
      if (result.detected && result.baseUrl) updateNowledge({ baseUrl: result.baseUrl })
    } catch (err) {
      setDetectMessage(err instanceof Error ? err.message : 'Nowledge detect failed')
    }
  }
  const detectOpenViking = async () => {
    if (!onDetectOpenViking) return
    try {
      const result = await onDetectOpenViking()
      setDetectMessage(result.message)
      if (result.detected && result.baseUrl) updateOpenViking({ baseUrl: result.baseUrl })
    } catch (err) {
      setDetectMessage(err instanceof Error ? err.message : 'OpenViking detect failed')
    }
  }

  return (
    <section className="memory-provider-panel">
      <div className="memory-provider-header">
        <div>
          <h3>{copy.title}</h3>
          <p>{copy.subtitle}</p>
        </div>
        <button className="settings-icon-button" type="button" aria-label={copy.refresh} onClick={onRefresh}>
          <RefreshCw size={16} />
        </button>
      </div>
      <div className="memory-provider-controls">
        <label>
          <input type="checkbox" checked={current.enabled} onChange={(event) => update({ enabled: event.currentTarget.checked })} />
          <span>{copy.enabled}</span>
        </label>
        <label>
          <input type="checkbox" checked={current.commitAfterRun} onChange={(event) => update({ commitAfterRun: event.currentTarget.checked })} />
          <span>{copy.commit}</span>
        </label>
      </div>
      <div className="memory-provider-choice-grid">
        {providerOptions.map((provider) => (
          <button key={provider.value} className={provider.value === selectedProvider ? 'selected' : undefined} type="button" onClick={() => update({ provider: provider.value })}>
            <span className="memory-provider-choice-dot" />
            <strong>{provider.label}</strong>
            <small>{provider.description}</small>
          </button>
        ))}
      </div>
      <div className="memory-provider-status-row">
        <div className="memory-provider-status-grid">
          <span>{copy.state}</span>
          <strong>{current.state}</strong>
          <span>{copy.configured}</span>
          <strong>{current.configured ? copy.configuredYes : copy.configuredNo}</strong>
        </div>
        <button className="settings-secondary-button" type="button" onClick={() => setConfigOpen(true)} disabled={!showProviderConfig}>
          <SettingsIcon size={14} />
          <span>{copy.configure}</span>
        </button>
      </div>
      <p className="memory-provider-diagnostic">{current.diagnostic.message}</p>
      {errors.length > 0 && (
        <div className="memory-provider-errors">
          <strong>{copy.recentErrors}</strong>
          {errors.slice(0, 3).map((item) => <span className="memory-provider-error-item" key={`${item.code}-${item.checkedAt ?? ''}-${item.runId ?? ''}`}>{memoryProviderErrorSummary(item, locale)}</span>)}
        </div>
      )}
      {saveResult?.status === 'saving' && <p className="memory-provider-diagnostic">Saving</p>}
      {saveResult?.message && saveResult.status !== 'saving' && <p className="memory-provider-diagnostic">{saveResult.message}</p>}
      {configOpen && showProviderConfig && (
        <div className="memory-provider-modal-backdrop" role="presentation" onClick={() => setConfigOpen(false)}>
          <section className="memory-provider-modal" role="dialog" aria-modal="true" aria-label={copy.configTitle} onClick={(event) => event.stopPropagation()}>
            <div className="memory-provider-modal-header">
              <div>
                <h3>{copy.configTitle}</h3>
                <p>{selectedProvider === 'nowledge' ? copy.nowledge : copy.openviking}</p>
              </div>
              <button className="settings-icon-button" type="button" aria-label={copy.close} onClick={() => setConfigOpen(false)}>
                <X size={16} />
              </button>
            </div>
            {selectedProvider === 'nowledge' && (
              <div className="memory-provider-config-grid">
                <label>
                  <span>{copy.baseUrl}</span>
                  <input value={nowledge.baseUrl ?? ''} onChange={(event) => updateNowledge({ baseUrl: event.currentTarget.value })} placeholder="http://127.0.0.1:7727" />
                </label>
                <label>
                  <span>{copy.apiKey}{nowledge.apiKeySet ? ` · ${copy.stored}` : ''}</span>
                  <input type="password" value={nowledge.apiKey ?? ''} onChange={(event) => updateNowledge({ apiKey: event.currentTarget.value })} placeholder={nowledge.apiKeySet ? copy.stored : ''} />
                </label>
                <label>
                  <span>{copy.timeout}</span>
                  <input type="number" min={0} value={nowledge.requestTimeoutMs ?? 0} onChange={(event) => updateNowledge({ requestTimeoutMs: Number(event.currentTarget.value) || 0 })} />
                </label>
                <button className="settings-secondary-button" type="button" onClick={() => void detectNowledge()} disabled={!onDetectNowledge}>
                  <RefreshCw size={14} />
                  <span>{copy.detect}</span>
                </button>
                {detectMessage && <p className="memory-provider-diagnostic">{detectMessage}</p>}
              </div>
            )}
            {selectedProvider === 'openviking' && (
              <div className="memory-provider-config-grid">
                <label>
                  <span>{copy.baseUrl}</span>
                  <input value={openviking.baseUrl ?? ''} onChange={(event) => updateOpenViking({ baseUrl: event.currentTarget.value })} placeholder="http://127.0.0.1:8282" />
                </label>
                <label>
                  <span>{copy.rootKey}{openviking.rootApiKeySet ? ` · ${copy.stored}` : ''}</span>
                  <input type="password" value={openviking.rootApiKey ?? ''} onChange={(event) => updateOpenViking({ rootApiKey: event.currentTarget.value })} placeholder={openviking.rootApiKeySet ? copy.stored : ''} />
                </label>
                <button className="settings-secondary-button" type="button" onClick={() => void detectOpenViking()} disabled={!onDetectOpenViking}>
                  <RefreshCw size={14} />
                  <span>{copy.detect}</span>
                </button>
                {detectMessage && <p className="memory-provider-diagnostic">{detectMessage}</p>}
                {[
                  ['embedding', copy.embedding],
                  ['vlm', copy.vlm],
                  ['rerank', copy.rerank],
                ].map(([prefix, label]) => (
                  <div className="memory-provider-config-group" key={prefix}>
                    <strong>{label}</strong>
                    <label>
                      <span>{copy.selector}</span>
                      <input value={String(openviking[`${prefix}Selector` as keyof typeof openviking] ?? '')} onChange={(event) => updateOpenViking({ [`${prefix}Selector`]: event.currentTarget.value })} />
                    </label>
                    <label>
                      <span>{copy.providerName}</span>
                      <input value={String(openviking[`${prefix}Provider` as keyof typeof openviking] ?? '')} onChange={(event) => updateOpenViking({ [`${prefix}Provider`]: event.currentTarget.value })} />
                    </label>
                    <label>
                      <span>{copy.model}</span>
                      <input value={String(openviking[`${prefix}Model` as keyof typeof openviking] ?? '')} onChange={(event) => updateOpenViking({ [`${prefix}Model`]: event.currentTarget.value })} />
                    </label>
                    <label>
                      <span>{copy.apiKey}{openviking[`${prefix}ApiKeySet` as keyof typeof openviking] ? ` · ${copy.stored}` : ''}</span>
                      <input type="password" value={String(openviking[`${prefix}ApiKey` as keyof typeof openviking] ?? '')} onChange={(event) => updateOpenViking({ [`${prefix}ApiKey`]: event.currentTarget.value })} placeholder={openviking[`${prefix}ApiKeySet` as keyof typeof openviking] ? copy.stored : ''} />
                    </label>
                    <label>
                      <span>{copy.apiBase}</span>
                      <input value={String(openviking[`${prefix}ApiBase` as keyof typeof openviking] ?? '')} onChange={(event) => updateOpenViking({ [`${prefix}ApiBase`]: event.currentTarget.value })} />
                    </label>
                    {prefix === 'embedding' && (
                      <label>
                        <span>{copy.dimension}</span>
                        <input type="number" min={0} value={openviking.embeddingDimension ?? 0} onChange={(event) => updateOpenViking({ embeddingDimension: Number(event.currentTarget.value) || 0 })} />
                      </label>
                    )}
                  </div>
                ))}
              </div>
            )}
          </section>
        </div>
      )}
    </section>
  )
}

function MemorySnapshotPanel({ overview, impression, loading, locale, onRebuildOverview, onRebuildImpression, onGetContent }: { overview?: MemoryOverviewSnapshot | null; impression?: MemoryImpressionSnapshot | null; loading?: boolean; locale: Locale; onRebuildOverview?: () => void; onRebuildImpression?: () => void; onGetContent?: (uri: string, layer?: 'overview' | 'read') => Promise<string> }) {
  const [contentOpen, setContentOpen] = useState(false)
  const [contentTitle, setContentTitle] = useState('')
  const [contentText, setContentText] = useState('')
  const [contentLoading, setContentLoading] = useState(false)
  const copy = locale === 'zh'
    ? {
        title: '运行记忆',
        subtitle: '查看本地已批准记忆生成的快照和画像；原始对话与工具输出不会出现在这里。',
        overview: '记忆快照',
        impression: '记忆画像',
        rebuild: '重建',
        loading: '更新中',
        updated: '更新时间',
        hits: '命中',
        emptyOverview: '暂无已保存记忆。',
        emptyImpression: '暂无记忆画像。',
        view: '查看记忆',
        close: '关闭',
      }
    : {
        title: 'Runtime Memory',
        subtitle: 'Inspect the local approved-memory snapshot and impression without raw conversations or tool output.',
        overview: 'Memory Snapshot',
        impression: 'Memory Impression',
        rebuild: 'Rebuild',
        loading: 'Updating',
        updated: 'Updated',
        hits: 'Hits',
        emptyOverview: 'No approved memories yet.',
        emptyImpression: 'No memory impression yet.',
        view: 'View Memory',
        close: 'Close',
      }
  const overviewText = overview?.memoryBlock || copy.emptyOverview
  const impressionText = impression?.impression || copy.emptyImpression
  const openHit = async (uri: string, title: string) => {
    if (!onGetContent) return
    setContentOpen(true)
    setContentTitle(title)
    setContentText('')
    setContentLoading(true)
    try {
      setContentText(await onGetContent(uri, 'read'))
    } finally {
      setContentLoading(false)
    }
  }

  return (
    <section className="memory-snapshot-panel">
      <div className="memory-provider-header">
        <div>
          <h3>{copy.title}</h3>
          <p>{copy.subtitle}</p>
        </div>
        {loading && <span className="memory-snapshot-loading">{copy.loading}</span>}
      </div>
      <div className="memory-snapshot-grid">
        <article className="memory-snapshot-card">
          <div className="memory-snapshot-card-header">
            <div>
              <h4>{copy.impression}</h4>
              <p>{copy.updated}: {impression?.updatedAt ? new Date(impression.updatedAt).toLocaleString() : '-'}</p>
            </div>
            <button className="settings-secondary-button" type="button" onClick={onRebuildImpression} disabled={!onRebuildImpression || loading}>
              <RefreshCw size={14} />
              <span>{copy.rebuild}</span>
            </button>
          </div>
          <p className="memory-snapshot-text">{impressionText}</p>
        </article>
        <article className="memory-snapshot-card">
          <div className="memory-snapshot-card-header">
            <div>
              <h4>{copy.overview}</h4>
              <p>{copy.hits}: {overview?.hits.length ?? 0} · {copy.updated}: {overview?.updatedAt ? new Date(overview.updatedAt).toLocaleString() : '-'}</p>
            </div>
            <button className="settings-secondary-button" type="button" onClick={onRebuildOverview} disabled={!onRebuildOverview || loading}>
              <RefreshCw size={14} />
              <span>{copy.rebuild}</span>
            </button>
          </div>
          <p className="memory-snapshot-text">{overviewText}</p>
          {overview?.hits?.length ? (
            <div className="memory-snapshot-hit-list">
              {overview.hits.slice(0, 4).map((hit) => (
                <button key={hit.uri} type="button" onClick={() => void openHit(hit.uri, hit.title || hit.uri)} disabled={!onGetContent}>{hit.title || hit.uri}</button>
              ))}
            </div>
          ) : null}
        </article>
      </div>
      {contentOpen && (
        <div className="memory-content-modal-backdrop" role="presentation" onClick={() => setContentOpen(false)}>
          <div className="memory-content-modal" role="dialog" aria-modal="true" aria-label={copy.view} onClick={(event) => event.stopPropagation()}>
            <div className="memory-content-modal-header">
              <div>
                <h4>{contentTitle || copy.view}</h4>
                <p>{copy.view}</p>
              </div>
              <button className="settings-icon-button" type="button" aria-label={copy.close} onClick={() => setContentOpen(false)}>
                <X size={16} />
              </button>
            </div>
            <pre>{contentLoading ? copy.loading : contentText || copy.emptyOverview}</pre>
          </div>
        </div>
      )}
    </section>
  )
}

function SegmentedControl<T extends string>({ value, options, onChange }: { value: T; options: { value: T; label: string }[]; onChange: (value: T) => void }) {
  return (
    <div className="settings-segmented-control">
      {options.map((option) => (
        <button key={option.value} className={option.value === value ? 'selected' : undefined} onClick={() => onChange(option.value)}>
          {option.label}
        </button>
      ))}
    </div>
  )
}

function StatusValue({ children }: { children: ReactNode }) {
  return <span className="settings-status-value">{children}</span>
}

type ProviderKind = 'openai_responses' | 'openai_chat_completions' | 'anthropic' | 'google_gemini'
type ProviderFilter = 'all' | 'enabled' | 'local' | 'cloud'

const providerKindOptions: { value: ProviderKind; label: string }[] = [
  { value: 'openai_responses', label: 'OpenAI (Responses)' },
  { value: 'openai_chat_completions', label: 'OpenAI (Chat Completions)' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'google_gemini', label: 'Google Gemini' },
]

function providerDisplayName(providerId: string) {
  if (providerId === 'local_codex') return 'Codex (Local)'
  if (providerId === 'local_claude_code' || providerId === 'claude_code') return 'Claude Code (Local)'
  return providerId
    .split(/[_-]/)
    .filter(Boolean)
    .map((part) => `${part.charAt(0).toUpperCase()}${part.slice(1)}`)
    .join(' ')
}

function providerKindLabel(kind: ProviderKind) {
  return providerKindOptions.find((option) => option.value === kind)?.label ?? providerKindOptions[0].label
}

function ProviderCapabilityList({ providerCapabilities, t }: { providerCapabilities: ProviderCapability[]; t: ReturnType<typeof getDictionary>['settings'] }) {
  if (!providerCapabilities.length) return <StatusValue>{t.notConnected}</StatusValue>
  return (
    <div className="provider-capability-list">
      {providerCapabilities.map((provider) => (
        <span key={provider.id}>
          {provider.id} · {provider.family} · {provider.model} · {capabilityLabel(provider.status, t)}
          {provider.localProvider && ` · ${t.localProviderSessionLocal} · ${t.localProviderCredentialRedacted}`}
          {provider.executionState === 'unsupported' && ` · ${t.localProviderExecutionUnsupported}`}
          {provider.executionState === 'supported' && ` · ${t.localProviderExecutionSupported}`}
        </span>
      ))}
    </div>
  )
}

function ProviderManagementPanel({
  providerCapabilities,
  providerCheckResults,
  providerSaveResult,
  providerDraftSettings,
  localProviderDetections,
  localProviderDetectionError,
  onProviderDraftSettingsChange,
  onSaveProvider,
  onCheckProvider,
  onDetectLocalProviders,
  onEnableLocalProvider,
  onDisableLocalProvider,
  isAddOpen,
  onCloseAdd,
  t,
}: Pick<Props, 'providerCapabilities' | 'providerCheckResults' | 'providerSaveResult' | 'providerDraftSettings' | 'localProviderDetections' | 'localProviderDetectionError' | 'onProviderDraftSettingsChange' | 'onSaveProvider' | 'onCheckProvider' | 'onDetectLocalProviders' | 'onEnableLocalProvider' | 'onDisableLocalProvider'> & { isAddOpen: boolean; onCloseAdd: () => void; t: ReturnType<typeof getDictionary>['settings'] }) {
  const [filter, setFilter] = useState<ProviderFilter>('all')
  const [query, setQuery] = useState('')
  const [providerName, setProviderName] = useState('My Provider')
  const [providerKind, setProviderKind] = useState<ProviderKind>('openai_responses')
  const [isKindOpen, setIsKindOpen] = useState(false)
  const [headerName, setHeaderName] = useState('')
  const [headerValue, setHeaderValue] = useState('')
  const capabilityIds = new Set(providerCapabilities.map((provider) => provider.id))
  const detectedOnlyProviders = localProviderDetections
    .filter((provider) => !capabilityIds.has(provider.providerId))
    .map((provider) => ({
      id: provider.providerId,
      name: provider.displayName,
      route: `local://${provider.providerId.replace(/^local_/, '').replaceAll('_', '-')}`,
      status: provider.status,
      statusText: localProviderStatusLabel(provider.status, t),
      checkCode: null,
      family: provider.providerKind,
      model: provider.modelCandidates[0] ?? provider.authMode,
      localProvider: true,
      readOnly: true,
      enabled: false,
      detectedStatus: provider.status,
    }))
  const providerCards = [
    ...providerCapabilities.map((provider) => ({
      id: provider.id,
      name: providerDisplayName(provider.id),
      route: provider.localProvider ? `local://${provider.id.replace(/^local_/, '').replaceAll('_', '-')}` : provider.baseUrl ?? provider.family,
      status: provider.status,
      statusText: provider.checkCode ?? capabilityLabel(provider.status, t),
      checkCode: provider.checkCode,
      family: provider.family,
      model: provider.model,
      localProvider: Boolean(provider.localProvider),
      readOnly: Boolean(provider.localProvider),
      enabled: ['available', 'configured', 'reachable', 'completion-ok'].includes(provider.status),
      detectedStatus: provider.status,
    })),
    ...detectedOnlyProviders,
  ]
  const visibleProviders = providerCards.filter((provider) => {
    const matchesQuery = `${provider.name} ${provider.route}`.toLowerCase().includes(query.trim().toLowerCase())
    const matchesFilter = filter === 'all' || (filter === 'enabled' && provider.enabled) || (filter === 'local' && provider.localProvider) || (filter === 'cloud' && !provider.localProvider)
    return matchesQuery && matchesFilter
  })
  const canSaveProvider = providerName.trim().length > 0 && providerDraftSettings.baseUrl.trim().length > 0 && providerDraftSettings.apiKey.trim().length > 0

  return (
    <>
      <div className="provider-management">
        <div className="provider-management-toolbar">
          <label className="provider-search">
            <Search size={17} />
            <input value={query} placeholder={t.providerSearchPlaceholder} onChange={(event) => setQuery(event.target.value)} />
          </label>
          <div className="provider-filter-tabs" role="tablist" aria-label={t.providerFilterLabel}>
            {[
              { value: 'all' as const, label: t.providerFilterAll },
              { value: 'enabled' as const, label: t.providerFilterEnabled },
              { value: 'local' as const, label: t.providerFilterLocal },
              { value: 'cloud' as const, label: t.providerFilterCloud },
            ].map((item) => (
              <button key={item.value} className={filter === item.value ? 'selected' : undefined} onClick={() => setFilter(item.value)}>
                {item.label}
              </button>
            ))}
          </div>
        </div>
        {localProviderDetectionError && <p className="provider-inline-error">{localProviderDetectionError}</p>}
        {!providerCards.length && (
          <div className="provider-empty-state">
            <span>{t.providerConsoleEmpty}</span>
            <button className="provider-secondary-action" onClick={onDetectLocalProviders}>{t.localProviderDetectAction}</button>
          </div>
        )}
        <div className="provider-card-grid">
          {visibleProviders.map((provider) => {
            const result = providerCheckResults[provider.id]
            return (
              <article className={`provider-management-card ${provider.enabled ? 'enabled' : 'idle'}`} key={provider.id}>
                <div className="provider-card-header">
                  <div className="provider-card-title">
                    <strong>{provider.name}</strong>
                    <code>{provider.route}</code>
                  </div>
                  <div className="provider-card-badges">
                    {provider.localProvider && <span>{t.providerFilterLocal}</span>}
                    {provider.readOnly && <span>{t.readOnly}</span>}
                  </div>
                </div>
                <div className="provider-card-body">
                  <span className={`provider-status-line ${provider.status}`}>
                    <span aria-hidden="true" />
                    {provider.statusText}
                  </span>
                  {provider.checkCode === 'completion-failed-503' && <small>completion-failed-503</small>}
                  <p>{provider.family} · {provider.model}</p>
                  {result?.message && <small>{t.providerCheckResult(result.status, result.message)}</small>}
                </div>
                <div className="provider-card-actions">
                  {provider.localProvider && !provider.enabled && provider.detectedStatus === 'available' && (
                    <button className="provider-card-test" onClick={() => onEnableLocalProvider(provider.id)}>
                      {t.localProviderEnableForSession}
                    </button>
                  )}
                  {provider.localProvider && provider.enabled && (
                    <button className="provider-card-test" onClick={() => onDisableLocalProvider(provider.id)}>
                      {t.localProviderDisableForSession}
                    </button>
                  )}
                  {capabilityIds.has(provider.id) && (
                    <button className="provider-card-test" onClick={() => onCheckProvider(provider.id)} disabled={result?.status === 'checking'}>
                      {result?.status === 'checking' ? t.providerChecking : t.providerTestConnection}
                    </button>
                  )}
                </div>
              </article>
            )
          })}
        </div>
      </div>

      {isAddOpen && (
        <div className="provider-modal-backdrop" role="presentation">
          <section className="provider-modal" role="dialog" aria-modal="true" aria-label={t.providerAdd}>
            <header className="provider-modal-header">
              <h2>{t.providerAdd}</h2>
              <button aria-label={t.close} onClick={onCloseAdd}>
                <X size={22} />
              </button>
            </header>
            <div className="provider-modal-grid">
              <label>
                <span>{t.providerName}</span>
                <input value={providerName} placeholder="My Provider" onChange={(event) => setProviderName(event.target.value)} />
              </label>
              <label>
                <span>{t.providerType}</span>
                <button className="provider-kind-select" type="button" onClick={() => setIsKindOpen(!isKindOpen)}>
                  {providerKindLabel(providerKind)}
                  <ChevronDown size={18} />
                </button>
                {isKindOpen && (
                  <div className="provider-kind-menu">
                    {providerKindOptions.map((option) => (
                      <button key={option.value} type="button" className={option.value === providerKind ? 'selected' : undefined} onClick={() => {
                        setProviderKind(option.value)
                        setIsKindOpen(false)
                      }}>
                        {option.label}
                        {option.value === providerKind && <Check size={17} />}
                      </button>
                    ))}
                  </div>
                )}
              </label>
              <label className="provider-modal-wide">
                <span>{t.providerApiKey}</span>
                <input type="password" value={providerDraftSettings.apiKey} placeholder="sk-..." onChange={(event) => onProviderDraftSettingsChange({ ...providerDraftSettings, apiKey: event.target.value, apiKeySet: event.target.value.length > 0 })} />
              </label>
              <label className="provider-modal-wide">
                <span>{t.providerBaseUrl}</span>
                <input value={providerDraftSettings.baseUrl} placeholder="https://api.openai.com/v1" onChange={(event) => onProviderDraftSettingsChange({ ...providerDraftSettings, baseUrl: event.target.value })} />
              </label>
            </div>
            <div className="provider-advanced">
              <h3>{t.providerAdvancedOptions}</h3>
              <span>{t.providerHeaders}</span>
              <div className="provider-header-row">
                <input value={headerName} placeholder={t.providerHeaderName} onChange={(event) => setHeaderName(event.target.value)} />
                <input value={headerValue} placeholder={t.providerHeaderValue} onChange={(event) => setHeaderValue(event.target.value)} />
                <button type="button" aria-label={t.remove}>
                  <Minus size={18} />
                </button>
              </div>
              <button className="provider-secondary-action" type="button">
                <Plus size={18} /> {t.providerAddHeader}
              </button>
            </div>
            <footer className="provider-modal-actions">
              <button className="provider-cancel-button" onClick={onCloseAdd}>{t.cancel}</button>
              <button
                className="provider-save-modal-button"
                disabled={!canSaveProvider || providerSaveResult.status === 'saving'}
                onClick={() => {
                  onSaveProvider({ ...providerDraftSettings, model: providerDraftSettings.model.trim() || 'gpt-5.5' })
                  onCloseAdd()
                }}
              >
                {providerSaveResult.status === 'saving' ? t.providerSaving : t.providerSave}
              </button>
            </footer>
          </section>
        </div>
      )}
    </>
  )
}

const toolDisplayCopy: Record<string, { zh: { name: string; description: string }; en: { name: string; description: string } }> = {
  'runtime.get_current_time': { zh: { name: '当前时间', description: '返回当前 UTC 时间。' }, en: { name: 'Current time', description: 'Returns the current UTC time.' } },
  'workspace.glob': { zh: { name: '查找工作区文件', description: '在已配置的工作区根目录下查找文件。' }, en: { name: 'Workspace glob', description: 'Find files under the configured workspace root.' } },
  'workspace.grep': { zh: { name: '搜索工作区文本', description: '在工作区文本文件中搜索匹配内容。' }, en: { name: 'Workspace grep', description: 'Search text files under the configured workspace root.' } },
  'workspace.read': { zh: { name: '读取工作区文件', description: '从单个工作区文件读取有边界的 UTF-8 文本片段。' }, en: { name: 'Workspace read', description: 'Read a bounded UTF-8 text slice from one workspace file.' } },
  'workspace.write_file': { zh: { name: '写入工作区文件', description: '在工作区根目录下创建有边界的 UTF-8 文本文件。' }, en: { name: 'Workspace write file', description: 'Create a bounded UTF-8 text file under the configured workspace root.' } },
  'workspace.edit': { zh: { name: '编辑工作区文件', description: '在工作区文件内执行一次有边界的精确文本替换。' }, en: { name: 'Workspace edit', description: 'Apply one bounded exact text replacement inside a workspace file.' } },
  'sandbox.exec_command': { zh: { name: '有界命令', description: '在工作区根目录下运行一次已批准的 argv 形式读码或验证命令；它不是隔离沙箱。' }, en: { name: 'Bounded command', description: 'Run one approved argv-form read or validation command under the configured workspace root. This is not an isolated sandbox.' } },
  'lsp.diagnostics': { zh: { name: 'LSP 诊断', description: '读取工作区源码文件的有边界诊断信息。' }, en: { name: 'LSP diagnostics', description: 'Read bounded diagnostics for a workspace source file.' } },
  'lsp.symbols': { zh: { name: 'LSP 符号', description: '读取工作区源码文件的有边界符号摘要。' }, en: { name: 'LSP symbols', description: 'Read bounded symbol summaries for a workspace source file.' } },
  'lsp.references': { zh: { name: 'LSP 引用', description: '读取某个源码位置的有边界工作区引用。' }, en: { name: 'LSP references', description: 'Read bounded workspace references for a source position.' } },
  'web.fetch': { zh: { name: '网页读取', description: '读取一个有边界的公开 HTTP(S) 地址，并返回安全文本摘要。' }, en: { name: 'Web fetch', description: 'Fetch one bounded public HTTP(S) URL and return a safe text summary.' } },
  'web.search': { zh: { name: '网页搜索', description: '通过已配置的 Brave 或 Tavily provider 搜索公开网页，并返回有边界的安全结果。' }, en: { name: 'Web search', description: 'Search the public web through a configured Brave or Tavily provider and return bounded safe results.' } },
  'browser.open': { zh: { name: '打开浏览器页面', description: '在 run 作用域浏览器会话中打开一个有边界的公开 HTTP(S) 页面。' }, en: { name: 'Browser open', description: 'Open one bounded public HTTP(S) page in a run-scoped browser session.' } },
  'browser.snapshot': { zh: { name: '浏览器快照', description: '返回 run 作用域浏览器会话的当前安全快照。' }, en: { name: 'Browser snapshot', description: 'Return the current safe snapshot for a run-scoped browser session.' } },
  'browser.click_link': { zh: { name: '点击浏览器链接', description: '从 run 作用域浏览器会话中导航到一个安全链接。' }, en: { name: 'Browser click link', description: 'Navigate one safe link from a run-scoped browser session.' } },
  'artifact.create_text': { zh: { name: '创建文本 artifact', description: '创建一个有边界、不可执行的文本 artifact。' }, en: { name: 'Artifact create text', description: 'Create one bounded non-executable text artifact.' } },
  'artifact.read': { zh: { name: '读取 artifact', description: '读取一个有边界的文本 artifact 摘录。' }, en: { name: 'Artifact read', description: 'Read one bounded text artifact excerpt.' } },
  'artifact.list': { zh: { name: '列出 artifacts', description: '列出有边界的安全 artifact 摘要。' }, en: { name: 'Artifact list', description: 'List bounded safe artifact summaries.' } },
  'agent.spawn': { zh: { name: '创建子任务', description: '创建一个有边界的子协调任务。' }, en: { name: 'Agent spawn', description: 'Create one bounded child coordination task.' } },
  'agent.list': { zh: { name: '列出子任务', description: '列出有边界的子协调任务摘要。' }, en: { name: 'Agent list', description: 'List bounded child coordination task summaries.' } },
  'agent.complete': { zh: { name: '完成子任务', description: '用有边界的结果摘要完成一个子协调任务。' }, en: { name: 'Agent complete', description: 'Complete one child coordination task with a bounded result summary.' } },
  'memory.search': { zh: { name: '搜索记忆', description: '在当前安全范围内搜索已批准的记忆摘要。' }, en: { name: 'Memory search', description: 'Search approved memory summaries in the current safe scope.' } },
  'memory.list': { zh: { name: '列出记忆', description: '列出当前安全范围内已批准的记忆摘要。' }, en: { name: 'Memory list', description: 'List approved memory summaries in the current safe scope.' } },
  'memory.read': { zh: { name: '读取记忆', description: '读取一个已批准的记忆摘要，不返回原始内容。' }, en: { name: 'Memory read', description: 'Read one approved memory summary without raw content.' } },
  'memory.write': { zh: { name: '写入记忆提案', description: '创建一个需要审批的记忆写入提案。' }, en: { name: 'Memory write', description: 'Create one approval-gated memory write proposal.' } },
  'memory.edit': { zh: { name: '编辑记忆提案', description: '编辑待审批提案，或创建替换记忆提案。' }, en: { name: 'Memory edit', description: 'Edit a pending proposal or create a replacement proposal.' } },
  'memory.forget': { zh: { name: '遗忘记忆', description: '通过审计边界删除一个已批准的记忆条目。' }, en: { name: 'Memory forget', description: 'Tombstone one approved memory entry through the audited memory boundary.' } },
  'memory.context': { zh: { name: '记忆上下文', description: '返回 provider 状态和相关记忆摘要。' }, en: { name: 'Memory context', description: 'Return provider status and relevant memory summaries.' } },
  'memory.timeline': { zh: { name: '记忆时间线', description: '列出安全的记忆审计时间线。' }, en: { name: 'Memory timeline', description: 'List safe memory audit timeline items.' } },
  'memory.connections': { zh: { name: '记忆关联', description: '返回与条目或查询相关的记忆摘要。' }, en: { name: 'Memory connections', description: 'Return related memory summaries for an entry or query.' } },
  'memory.thread_search': { zh: { name: '线程记忆搜索', description: '搜索本地线程和消息历史的安全摘录。' }, en: { name: 'Memory thread search', description: 'Search local thread and message history with safe excerpts.' } },
  'memory.thread_fetch': { zh: { name: '线程记忆读取', description: '读取本地线程消息的安全摘录。' }, en: { name: 'Memory thread fetch', description: 'Fetch safe local thread message excerpts.' } },
  'memory.status': { zh: { name: '记忆状态', description: '返回记忆 provider 的配置和可用状态。' }, en: { name: 'Memory status', description: 'Return memory provider readiness and configuration state.' } },
  'notebook.read': { zh: { name: 'Notebook 读取', description: '读取一个已批准的结构化 notebook 条目。' }, en: { name: 'Notebook read', description: 'Read one approved structured notebook entry.' } },
  'notebook.write': { zh: { name: 'Notebook 写入', description: '通过审批和审计边界写入结构化 notebook 条目。' }, en: { name: 'Notebook write', description: 'Write one approval-gated structured notebook entry.' } },
  'notebook.edit': { zh: { name: 'Notebook 编辑', description: '替换一个结构化 notebook 条目并保留审计记录。' }, en: { name: 'Notebook edit', description: 'Replace one structured notebook entry with audit history.' } },
  'notebook.forget': { zh: { name: 'Notebook 遗忘', description: '通过审计边界删除一个结构化 notebook 条目。' }, en: { name: 'Notebook forget', description: 'Tombstone one structured notebook entry through the audited memory boundary.' } },
}

function toolCopy(tool: ToolCatalogItem, locale: Locale) {
  const copy = toolDisplayCopy[tool.name]?.[locale]
  return {
    name: copy?.name ?? (tool.displayName || tool.name),
    description: copy?.description ?? tool.description,
  }
}

function toolBadgeLabel(value: string, locale: Locale) {
  if (locale === 'en') return value
  const labels: Record<string, string> = {
    builtin: '内置',
    mcp: 'MCP',
    runtime: '运行时',
    workspace: '工作区',
    artifact: 'Artifact',
    sandbox: '沙箱',
    lsp: 'LSP',
    web: '网页',
    browser: '浏览器',
    agent: 'Agent',
    memory: '记忆',
    'read-only': '只读',
    'write-capable': '可写',
    'exec-capable': '可执行命令',
    'non-executable': '不可执行',
    'coordination-only': '仅协调',
    'no autonomous execution': '不自主执行',
    'approval-gated': '需审批',
    'public HTTP only': '仅公开 HTTP',
    low: '低风险',
    medium: '中风险',
    high: '高风险',
    always_required: '始终需批准',
    enabled: '已启用',
    disabled: '已禁用',
    executable: '可执行',
    not_discovered: '未发现',
    not_allowed: '未允许',
    non_executable: '不可执行',
    read_only: '只读策略',
  }
  return labels[value] ?? value
}

function toolScopeLabel(tool: ToolCatalogItem, locale: Locale) {
  const scope = tool.safeMetadata?.scope
  const value = typeof scope === 'string' && scope.trim() ? scope.trim() : tool.group
  return locale === 'zh' ? `${toolBadgeLabel(value, locale)}范围` : `${value} scope`
}

function skillSourceLabel(skill: InstalledSkill, locale: Locale) {
  if (skill.sourceLabel) return skill.sourceLabel
  const labels: Record<string, string> = {
    project: locale === 'zh' ? '项目' : 'Project',
    codex: 'Codex',
    claude_code: 'Claude Code',
    agents: locale === 'zh' ? '用户技能' : 'User skills',
    plugin_cache: locale === 'zh' ? '插件缓存' : 'Plugin cache',
  }
  return labels[skill.source] ?? skill.source
}

function SkillPanel({ personas, skills, loading, error, locale }: { personas: Persona[]; skills: InstalledSkill[]; loading?: boolean; error?: string | null; locale: Locale }) {
  const copy = locale === 'zh'
    ? {
        noPersona: '暂无 persona',
        noSkills: loading ? '正在读取已安装技能' : '暂无已安装技能',
        error: '读取技能失败',
        default: '默认',
        version: '版本',
        installed: '已安装',
        personaSection: 'Loomi Persona',
        skillSection: '已安装 Skill',
      }
    : {
        noPersona: 'No personas',
        noSkills: loading ? 'Loading installed skills' : 'No installed skills',
        error: 'Failed to load skills',
        default: 'Default',
        version: 'Version',
        installed: 'Installed',
        personaSection: 'Loomi Personas',
        skillSection: 'Installed Skills',
      }

  return (
    <div className="skill-settings-surface" data-testid="skill-settings">
      <div className="skill-summary-strip">
        <span>{copy.personaSection}: <strong>{personas.length}</strong></span>
        <span>{copy.skillSection}: <strong>{skills.length}</strong></span>
      </div>

      <section className="skill-settings-section" aria-label={copy.personaSection}>
        <h2>{copy.personaSection}</h2>
        <div className="tools-catalog-list">
          {personas.length ? personas.map((persona) => (
            <article className="tools-catalog-row skill-catalog-row" key={persona.id}>
              <div className="tools-catalog-main">
                <div className="tools-catalog-heading">
                  <strong>{persona.name}</strong>
                  <code>{persona.slug}</code>
                </div>
                <p>{persona.description}</p>
              </div>
              <div className="tools-catalog-badges">
                {persona.isDefault && <span>{copy.default}</span>}
                <span>{copy.version}: {persona.activeVersion}</span>
              </div>
            </article>
          )) : <StatusValue>{copy.noPersona}</StatusValue>}
        </div>
      </section>

      <section className="skill-settings-section" aria-label={copy.skillSection}>
        <h2>{copy.skillSection}</h2>
        {error && <p className="skill-error">{copy.error}: {error}</p>}
        <div className="tools-catalog-list">
          {skills.length ? skills.map((skill) => (
            <article className="tools-catalog-row skill-catalog-row" key={skill.id}>
              <div className="tools-catalog-main">
                <div className="tools-catalog-heading">
                  <strong>{skill.name}</strong>
                  {skill.package && <code>{skill.package}</code>}
                </div>
                <p>{skill.description || skill.path}</p>
                <small>{skill.path}</small>
              </div>
              <div className="tools-catalog-badges">
                <span>{skillSourceLabel(skill, locale)}</span>
                <span>{copy.installed}</span>
              </div>
            </article>
          )) : <StatusValue>{copy.noSkills}</StatusValue>}
        </div>
      </section>
    </div>
  )
}

function ToolsPanel({ tools, locale }: { tools: ToolCatalogItem[]; locale: Locale }) {
  const visibleTools = tools.filter((tool) => tool.name !== 'web.search')
  if (!visibleTools.length) return <StatusValue>{locale === 'zh' ? '暂无工具目录' : 'No tools discovered'}</StatusValue>
  return (
    <div className="tools-catalog-list" data-testid="tools-catalog-list">
      {visibleTools.map((tool) => {
        const copy = toolCopy(tool, locale)
        return (
          <article className="tools-catalog-row" key={tool.name}>
            <div className="tools-catalog-main">
              <div className="tools-catalog-heading">
                <strong>{copy.name}</strong>
                <code>{tool.name}</code>
              </div>
              <p>{copy.description}</p>
              {tool.inputSchemaHash && <small>{tool.inputSchemaHash}</small>}
            </div>
            <div className="tools-catalog-badges">
              <span>{toolBadgeLabel(tool.source, locale)}</span>
              <span>{toolBadgeLabel(tool.group, locale)}</span>
              {tool.safeMetadata?.read_only === true && <span>{toolBadgeLabel('read-only', locale)}</span>}
              {tool.safeMetadata?.write_capable === true && <span>{toolBadgeLabel('write-capable', locale)}</span>}
              {tool.safeMetadata?.exec_capable === true && <span>{toolBadgeLabel('exec-capable', locale)}</span>}
              {tool.safeMetadata?.non_executable === true && <span>{toolBadgeLabel('non-executable', locale)}</span>}
              {tool.safeMetadata?.coordination_only === true && <span>{toolBadgeLabel('coordination-only', locale)}</span>}
              {tool.safeMetadata?.autonomous_execution === false && <span>{toolBadgeLabel('no autonomous execution', locale)}</span>}
              {tool.safeMetadata?.approval_gated === true && <span>{toolBadgeLabel('approval-gated', locale)}</span>}
              <span>{toolScopeLabel(tool, locale)}</span>
              {tool.safeMetadata?.network_access === 'public_http_only' && <span>{toolBadgeLabel('public HTTP only', locale)}</span>}
              <span>{toolBadgeLabel(tool.riskLevel, locale)}</span>
              <span>{toolBadgeLabel(tool.approvalPolicy, locale)}</span>
              <span>{toolBadgeLabel(tool.enabled ? 'enabled' : 'disabled', locale)}</span>
              <span>{toolBadgeLabel(tool.executionState, locale)}</span>
            </div>
          </article>
        )
      })}
    </div>
  )
}

function WebSearchPanel({ tools, locale, config, saveResult, onSave }: { tools: ToolCatalogItem[]; locale: Locale; config?: WebSearchConfig | null; saveResult?: ProviderSaveResult; onSave: (input: { tavilyApiKey?: string; braveApiKey?: string }) => void }) {
  const webSearch = tools.find((tool) => tool.name === 'web.search')
  const [tavilyApiKey, setTavilyApiKey] = useState('')
  const [braveApiKey, setBraveApiKey] = useState('')
  const canSave = tavilyApiKey.trim() !== '' || braveApiKey.trim() !== ''
  const copy = locale === 'zh'
    ? {
        title: '网页搜索',
        description: '填 Tavily 或 Brave Search key，一个就能用，两个都填更好。',
        tavily: 'Tavily Key',
        brave: 'Brave Search Key',
        save: '保存',
        saving: '保存中',
        saved: '已保存',
        ready: '已可用',
        missing: '未配置',
        toolMissing: 'API 还没返回 web.search，重启后再看。',
        hidden: '已保存的 key 不显示。',
      }
    : {
        title: 'Web Search',
        description: 'Add a Tavily or Brave Search key. One key is enough; two is better.',
        tavily: 'Tavily Key',
        brave: 'Brave Search Key',
        save: 'Save',
        saving: 'Saving',
        saved: 'Saved',
        ready: 'Ready',
        missing: 'Not configured',
        toolMissing: 'API has not returned web.search yet. Restart and check again.',
        hidden: 'Saved keys are hidden.',
      }
  const status = config?.enabled ? copy.ready : copy.missing

  return (
    <div className="settings-card-stack web-search-settings" data-testid="web-search-settings">
      <section className="settings-card">
        <div className="settings-card-head">
          <h2>{copy.title}</h2>
          <p>{copy.description}</p>
        </div>
        <div className="web-search-simple-form">
          <label>
            <span>{copy.tavily}</span>
            <input type="password" value={tavilyApiKey} placeholder={config?.hasTavilyKey ? copy.hidden : 'tvly-...'} onChange={(event) => setTavilyApiKey(event.target.value)} />
          </label>
          <label>
            <span>{copy.brave}</span>
            <input type="password" value={braveApiKey} placeholder={config?.hasBraveKey ? copy.hidden : 'BSA...'} onChange={(event) => setBraveApiKey(event.target.value)} />
          </label>
          <div className="web-search-actions">
            <button className="provider-add-button" disabled={!canSave || saveResult?.status === 'saving'} onClick={() => {
              onSave({ tavilyApiKey, braveApiKey })
              setTavilyApiKey('')
              setBraveApiKey('')
            }}>{saveResult?.status === 'saving' ? copy.saving : copy.save}</button>
            <span>{status}</span>
            {webSearch ? <code>web.search</code> : <span>{copy.toolMissing}</span>}
          </div>
          {saveResult?.status === 'success' && <p className="web-search-save-status">{copy.saved}</p>}
          {saveResult?.status === 'failed' && saveResult.message && <p className="web-search-save-status error">{saveResult.message}</p>}
        </div>
      </section>
    </div>
  )
}

function parseMCPArgs(value: string) {
  return value.split('\n').map((item) => item.trim()).filter(Boolean)
}

function parseMCPEnv(value: string) {
  const env: Record<string, string> = {}
  for (const line of value.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed || !trimmed.includes('=')) continue
    const [key, ...rest] = trimmed.split('=')
    if (key.trim()) env[key.trim()] = rest.join('=').trim()
  }
  return env
}

function MCPPanel({ servers, locale, actionResult, onSave, onDelete, onDiscover }: { servers: MCPServerStatus[]; locale: Locale; actionResult?: ProviderSaveResult; onSave?: (input: MCPServerConfigInput) => void; onDelete?: (slug: string) => void; onDiscover?: (slug: string) => void }) {
  const copy = locale === 'zh'
    ? { title: '本地 stdio MCP', name: '名称', slug: 'Slug', command: 'Command', args: 'Args', env: 'Env', timeout: 'Timeout', enabled: '启用', save: '保存配置', saving: '处理中', discover: '连接测试', remove: '删除', empty: '暂无本地 MCP server 配置', commandHidden: '命令和环境变量保存后不会回显。', argsHint: '每行一个参数', envHint: '每行 KEY=VALUE' }
    : { title: 'Local stdio MCP', name: 'Name', slug: 'Slug', command: 'Command', args: 'Args', env: 'Env', timeout: 'Timeout', enabled: 'Enabled', save: 'Save config', saving: 'Working', discover: 'Test connection', remove: 'Delete', empty: 'No local MCP servers configured', commandHidden: 'Command and env values are not echoed after save.', argsHint: 'One argument per line', envHint: 'One KEY=VALUE per line' }
  const [slug, setSlug] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [command, setCommand] = useState('')
  const [argsText, setArgsText] = useState('')
  const [envText, setEnvText] = useState('')
  const [timeoutMs, setTimeoutMs] = useState(5000)
  const [enabled, setEnabled] = useState(true)
  const canSave = slug.trim() !== '' && displayName.trim() !== '' && (!enabled || command.trim() !== '')
  const isBusy = actionResult?.status === 'saving'

  return (
    <div className="mcp-settings-surface" data-testid="mcp-settings">
      <section className="mcp-config-form" aria-label={copy.title}>
        <h2>{copy.title}</h2>
        <div className="mcp-form-grid">
          <label><span>{copy.name}</span><input value={displayName} placeholder="Local Search" onChange={(event) => setDisplayName(event.target.value)} /></label>
          <label><span>{copy.slug}</span><input value={slug} placeholder="local-search" onChange={(event) => setSlug(event.target.value)} /></label>
          <label className="mcp-form-wide"><span>{copy.command}</span><input value={command} placeholder="/path/to/mcp-server" onChange={(event) => setCommand(event.target.value)} /></label>
          <label><span>{copy.args}</span><textarea value={argsText} placeholder={copy.argsHint} onChange={(event) => setArgsText(event.target.value)} /></label>
          <label><span>{copy.env}</span><textarea value={envText} placeholder={copy.envHint} onChange={(event) => setEnvText(event.target.value)} /></label>
          <label><span>{copy.timeout}</span><input type="number" min={100} max={60000} value={timeoutMs} onChange={(event) => setTimeoutMs(Number(event.target.value) || 5000)} /></label>
          <label className="mcp-checkbox"><input type="checkbox" checked={enabled} onChange={(event) => setEnabled(event.target.checked)} /><span>{copy.enabled}</span></label>
        </div>
        <div className="mcp-form-actions">
          <button className="provider-add-button" disabled={!canSave || isBusy} onClick={() => onSave?.({ slug: slug.trim(), displayName: displayName.trim(), enabled, transport: 'stdio', command: command.trim(), args: parseMCPArgs(argsText), env: parseMCPEnv(envText), timeoutMs })}>{isBusy ? copy.saving : copy.save}</button>
          <span>{copy.commandHidden}</span>
          {actionResult?.message && <span className={actionResult.status === 'failed' ? 'skill-error' : undefined}>{actionResult.message}</span>}
        </div>
      </section>

      {!servers.length ? <StatusValue>{copy.empty}</StatusValue> : (
        <div className="tools-catalog-list" data-testid="mcp-server-list">
          {servers.map((server) => (
            <article className="tools-catalog-row" key={server.serverSafeId}>
              <div className="tools-catalog-main">
                <div className="tools-catalog-heading">
                  <strong>{server.displayName || server.serverSlug}</strong>
                  <code>{server.serverSlug}</code>
                </div>
                <p>{server.candidateNames.join(', ') || (locale === 'zh' ? '未发现工具' : 'No discovered tools')}</p>
                {server.redactedErrorCode && <small>{server.redactedErrorCode}</small>}
              </div>
              <div className="tools-catalog-badges">
                <span>{server.transport}</span>
                <span>{server.configSource}</span>
                <span>{toolBadgeLabel(server.enabled ? 'enabled' : 'disabled', locale)}</span>
                <span>{toolBadgeLabel(server.discoveryStatus, locale)}</span>
                <span>{server.executionMode}</span>
                <span>{locale === 'zh' ? `${server.candidateCount} 个工具` : `${server.candidateCount} tools`}</span>
                <button onClick={() => onDiscover?.(server.serverSlug)} disabled={isBusy}>{copy.discover}</button>
                <button onClick={() => onDelete?.(server.serverSlug)} disabled={isBusy}>{copy.remove}</button>
              </div>
            </article>
          ))}
        </div>
      )}
    </div>
  )
}

export function SettingsView({
  locale,
  selectedCategoryId,
  defaultWorkspaceMode,
  theme,
  backendCapability,
  providerCapabilities,
  personas = [],
  installedSkills = [],
  skillsLoading,
  skillsError,
  toolCatalog,
  webSearchConfig,
  webSearchSaveResult,
  mcpServers = [],
  mcpActionResult,
  localProviderDetections,
  localProviderDetectionError,
  memoryEntries,
  memoryQuery,
  memoryFilters,
  memoryLoading,
  memoryError,
  memoryDetail,
  memoryDetailLoading,
  memoryDetailError,
  memoryAuditItems,
  memoryAuditLoading,
  memoryAuditError,
  memoryWriteProposals = [],
  memoryProposalsLoading,
  memoryProposalsError,
  memoryProviderStatus,
  memoryErrors = [],
  memoryProviderSaveResult,
  memoryOverviewSnapshot,
  memoryImpressionSnapshot,
  memorySnapshotLoading,
  pendingDeleteMemoryEntry,
  providerCheckResults,
  providerSaveResult,
  providerDraftSettings,
  onSelectLocale,
  onSelectCategory,
  onSelectDefaultWorkspaceMode,
  onSelectTheme,
  onProviderDraftSettingsChange,
  onSaveProvider,
  onSaveWebSearchKeys,
  onRefreshMemoryProviderStatus,
  onUpdateMemoryProvider,
  onDetectNowledgeMemoryProvider,
  onDetectOpenVikingMemoryProvider,
  onRebuildMemoryOverviewSnapshot,
  onRebuildMemoryImpressionSnapshot,
  onGetMemoryContent,
  onSaveMCPServer,
  onDeleteMCPServer,
  onDiscoverMCPServer,
  onCheckProvider,
  onDetectLocalProviders,
  onEnableLocalProvider,
  onDisableLocalProvider,
  onMemoryQueryChange,
  onMemoryFiltersChange,
  onOpenMemoryDetail,
  onCloseMemoryDetail,
  onRequestDeleteMemoryEntry,
  onCancelDeleteMemoryEntry,
  onCreateMemoryEntry,
  onConfirmDeleteMemoryEntry,
  onApproveMemoryProposal,
  onUpdateMemoryProposal,
  onDenyMemoryProposal,
  onBack,
}: Props) {
  const dictionary = getDictionary(locale)
  const t = dictionary.settings
  const selectedCategory = getSettingsCategory(selectedCategoryId, locale)
  const isGeneral = selectedCategory.id === 'general'
  const isProviders = selectedCategory.id === 'providers'
  const isWebSearch = selectedCategory.id === 'web-search'
  const isSkill = selectedCategory.id === 'skill'
  const isMemory = selectedCategory.id === 'memory'
  const isMCP = selectedCategory.id === 'mcp'
  const isTools = selectedCategory.id === 'tools'
  const isAbout = selectedCategory.id === 'about'
  const [isProviderAddOpen, setIsProviderAddOpen] = useState(false)

  return (
    <div className="settings-shell" aria-label={t.title}>
      <nav className="settings-sidebar" aria-label="Settings categories">
        <button className="settings-back-button" onClick={onBack}>
          <ArrowLeft size={15} /> {t.back}
        </button>
        {categoryGroups.map((group) => (
          <div className="settings-nav-group" key={group}>
            <span>{settingsCategoryGroups[group][locale]}</span>
            {getSettingsCategoriesByGroup(group, locale).map((category) => (
              <button key={category.id} className={category.id === selectedCategoryId ? 'selected' : undefined} onClick={() => onSelectCategory(category.id)}>
                <span>{category.label}</span>
                <small>{categoryStatusLabel(category, t)}</small>
                <ChevronRight size={13} />
              </button>
            ))}
          </div>
        ))}
      </nav>

      <section className="settings-content">
        <header className="settings-content-header">
          <div>
            <span className="settings-eyebrow">{t.title}</span>
            <h1>{selectedCategory.label}</h1>
            <p>{selectedCategory.description}</p>
          </div>
          {isProviders ? (
            <button className="provider-add-button" onClick={() => setIsProviderAddOpen(true)}>
              <Plus size={18} /> {t.providerAdd}
            </button>
          ) : (
            <span className={`settings-category-pill ${selectedCategory.status}`}>{categoryStatusLabel(selectedCategory, t)}</span>
          )}
        </header>

        {isGeneral && (
          <div className="settings-card-stack">
            <section className="settings-card">
              <div className="settings-card-head">
                <h2>{t.workspaceDefaults}</h2>
                <p>{t.workspaceDefaultsDescription}</p>
              </div>
              <SettingRow
                label={t.language}
                helperText={t.languageHelper}
                status="working"
                t={t}
                control={(
                  <SegmentedControl
                    value={locale}
                    options={[{ value: 'zh', label: t.chinese }, { value: 'en', label: t.english }]}
                    onChange={onSelectLocale}
                  />
                )}
              />
              <SettingRow
                label={t.defaultWorkspaceMode}
                helperText={t.defaultWorkspaceModeHelper}
                status="working"
                t={t}
                control={(
                  <SegmentedControl
                    value={defaultWorkspaceMode}
                    options={[{ value: 'chat', label: dictionary.app.chat }, { value: 'work', label: dictionary.app.work }]}
                    onChange={onSelectDefaultWorkspaceMode}
                  />
                )}
              />
              <SettingRow
                label={t.theme}
                helperText={t.themeHelper}
                status="working"
                t={t}
                control={(
                  <SegmentedControl
                    value={theme}
                    options={[{ value: 'light', label: t.light }, { value: 'dark', label: t.dark }]}
                    onChange={onSelectTheme}
                  />
                )}
              />
            </section>
          </div>
        )}

        {isProviders && (
          <div className="settings-card-stack">
            <ProviderManagementPanel
              providerCapabilities={providerCapabilities}
              providerCheckResults={providerCheckResults}
              providerSaveResult={providerSaveResult}
              providerDraftSettings={providerDraftSettings}
              localProviderDetections={localProviderDetections}
              localProviderDetectionError={localProviderDetectionError}
              onProviderDraftSettingsChange={onProviderDraftSettingsChange}
              onSaveProvider={onSaveProvider}
              onCheckProvider={onCheckProvider}
              onDetectLocalProviders={onDetectLocalProviders}
              onEnableLocalProvider={onEnableLocalProvider}
              onDisableLocalProvider={onDisableLocalProvider}
              isAddOpen={isProviderAddOpen}
              onCloseAdd={() => setIsProviderAddOpen(false)}
              t={t}
            />
          </div>
        )}

        {isWebSearch && <WebSearchPanel tools={toolCatalog} locale={locale} config={webSearchConfig} saveResult={webSearchSaveResult} onSave={onSaveWebSearchKeys ?? (() => {})} />}

        {isSkill && <SkillPanel personas={personas} skills={installedSkills} loading={skillsLoading} error={skillsError} locale={locale} />}

        {isMemory && (
          <div className="memory-settings-surface">
            <MemoryPanel
              entries={memoryEntries}
              query={memoryQuery}
              locale={locale}
              filters={memoryFilters}
              loading={memoryLoading}
              error={memoryError}
              detailEntry={memoryDetail}
              detailLoading={memoryDetailLoading}
              detailError={memoryDetailError}
              auditItems={memoryAuditItems}
              auditLoading={memoryAuditLoading}
              auditError={memoryAuditError}
              writeProposals={memoryWriteProposals}
              proposalsLoading={memoryProposalsLoading}
              proposalsError={memoryProposalsError}
              pendingDeleteEntry={pendingDeleteMemoryEntry}
              onQueryChange={onMemoryQueryChange}
              onFiltersChange={onMemoryFiltersChange}
              onOpenDetail={onOpenMemoryDetail}
              onCloseDetail={onCloseMemoryDetail}
              onRequestDelete={onRequestDeleteMemoryEntry}
              onCancelDelete={onCancelDeleteMemoryEntry}
              onCreateMemory={onCreateMemoryEntry}
              onConfirmDelete={onConfirmDeleteMemoryEntry}
              onApproveProposal={onApproveMemoryProposal}
              onUpdateProposal={onUpdateMemoryProposal}
              onDenyProposal={onDenyMemoryProposal}
            />
            <MemorySnapshotPanel
              overview={memoryOverviewSnapshot}
              impression={memoryImpressionSnapshot}
              loading={memorySnapshotLoading}
              locale={locale}
              onRebuildOverview={onRebuildMemoryOverviewSnapshot}
              onRebuildImpression={onRebuildMemoryImpressionSnapshot}
              onGetContent={onGetMemoryContent}
            />
            <MemoryProviderFoundationPanel
              status={memoryProviderStatus}
              errors={memoryErrors}
              saveResult={memoryProviderSaveResult}
              locale={locale}
              onRefresh={onRefreshMemoryProviderStatus}
              onUpdate={onUpdateMemoryProvider}
              onDetectNowledge={onDetectNowledgeMemoryProvider}
              onDetectOpenViking={onDetectOpenVikingMemoryProvider}
            />
          </div>
        )}

        {isMCP && (
          <MCPPanel
            servers={mcpServers}
            locale={locale}
            actionResult={mcpActionResult}
            onSave={onSaveMCPServer}
            onDelete={onDeleteMCPServer}
            onDiscover={onDiscoverMCPServer}
          />
        )}

        {isTools && (
          <div className="tools-settings-surface">
            <section aria-label={selectedCategory.label}>
              <ToolsPanel tools={toolCatalog} locale={locale} />
            </section>
          </div>
        )}

        {isAbout && (
          <div className="settings-card-stack">
            <section className="settings-card">
              <div className="settings-card-head">
                <h2>{t.aboutLocalApp}</h2>
                <p>{t.aboutLocalAppDescription}</p>
              </div>
              <SettingRow label={t.appName} helperText={t.appNameHelper} status="read_only" t={t} control={<StatusValue>Loomi</StatusValue>} />
              <SettingRow label={t.appVersion} helperText={t.appVersionHelper} status="mock" t={t} control={<StatusValue>{t.previewOnly}</StatusValue>} />
              <SettingRow label={t.appStatus} helperText={t.appStatusHelper} status="read_only" t={t} control={<StatusValue>Real API · {capabilityLabel(backendCapability, t)}</StatusValue>} />
            </section>
          </div>
        )}
      </section>
    </div>
  )
}
