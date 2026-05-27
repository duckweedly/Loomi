import { Select, Switch } from 'animal-island-ui'
import { Check, ChevronDown, Eye, Filter, Pencil, Trash2, X } from 'lucide-react'
import { useState } from 'react'
import type { MemoryAuditItem, MemoryEntry, MemoryFilters, MemoryWriteProposal } from '../domain'
import type { Locale } from '../i18n'

type Props = {
  entries: MemoryEntry[]
  query: string
  locale?: Locale
  filters?: MemoryFilters
  loading?: boolean
  error?: string | null
  auditItems?: MemoryAuditItem[]
  auditLoading?: boolean
  auditError?: string | null
  writeProposals?: MemoryWriteProposal[]
  proposalsLoading?: boolean
  proposalsError?: string | null
  detailEntry?: MemoryEntry | null
  detailLoading?: boolean
  detailError?: string | null
  pendingDeleteEntry?: MemoryEntry | null
  onQueryChange: (query: string) => void
  onFiltersChange: (filters: MemoryFilters) => void
  onOpenDetail: (entry: MemoryEntry) => void
  onCloseDetail?: () => void
  onRequestDelete: (entry: MemoryEntry) => void
  onCancelDelete?: () => void
  onCreateMemory?: (input: { title: string; content: string; scopeType?: 'user' | 'thread'; scopeId?: string }) => void
  onConfirmDelete?: (entry: MemoryEntry) => void
  onApproveProposal?: (proposal: MemoryWriteProposal) => void
  onUpdateProposal?: (proposal: MemoryWriteProposal, input: { title: string; summary: string }) => void
  onDenyProposal?: (proposal: MemoryWriteProposal) => void
}

const auditEventTypes: MemoryAuditItem['eventType'][] = [
  'memory_write_approved',
  'memory_write_denied',
  'memory_deleted',
  'memory_write_proposed',
]

function metadataValue(value?: string | null) {
  return value && value.trim() ? value : 'none'
}

function isActiveFilter(filters: MemoryFilters) {
  return Boolean(filters.scopeType || filters.scopeId || filters.sourceThreadId || filters.sourceRunId || (filters.sourceType && filters.sourceType !== 'any') || filters.includeTombstoned || (filters.limit && filters.limit !== 20))
}

function formatDate(value: string) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function redactionLabel(redactionApplied: boolean, locale: Locale) {
  return redactionApplied ? (locale === 'zh' ? '已脱敏' : 'Redacted') : (locale === 'zh' ? '未脱敏' : 'Not redacted')
}

function scopeLabel(entry: MemoryEntry, locale: Locale) {
  if (entry.scopeType === 'thread') return locale === 'zh' ? '会话记忆' : 'Thread memory'
  return locale === 'zh' ? '全局记忆' : 'User memory'
}

function statusLabel(status: MemoryEntry['status'], locale: Locale) {
  if (locale === 'en') return status === 'tombstoned' ? 'Deleted' : status
  const labels: Record<MemoryEntry['status'], string> = {
    approved: '已保存',
    tombstoned: '已删除',
    disabled: '已停用',
  }
  return labels[status]
}

function copyFor(locale: Locale) {
  return locale === 'zh'
    ? {
        title: '记忆控制台',
        subtitle: '审批 AI 提出的记忆、检索已保存内容，并确认运行时只读取安全摘要。',
        search: '搜索记忆',
        filters: '筛选',
        hideFilters: '收起筛选',
        loading: '读取中',
        scopeType: '范围',
        scopeId: '范围 ID',
        sourceThread: '来源会话',
        sourceRun: '来源运行',
        sourceType: '来源类型',
        limit: '数量',
        any: '全部',
        user: '全局',
        thread: '会话',
        manual: '手动',
        run: '运行',
        includeDeleted: '包含已删除',
        activeFilters: '已启用筛选',
        clearFilters: '清除',
        emptyTitle: '暂无已保存记忆',
        emptyBody: '记忆需要由运行过程提出并通过批准后，才会出现在这里。',
        errorTitle: '记忆读取失败',
        detail: '记忆详情',
        closeDetail: '关闭记忆详情',
        scope: '范围',
        source: '来源',
        status: '状态',
        redaction: '脱敏',
        created: '创建',
        updated: '更新',
        deletedAt: '删除时间',
        confirmDelete: '确认删除',
        cancel: '取消',
        deleteMemory: '删除记忆',
        history: '记忆历史',
        historyEmpty: '暂无记忆历史',
        historyError: '记忆历史读取失败',
        systemSnapshot: '系统快照',
        systemSnapshotLoaded: '运行时读取了记忆快照',
        hiddenSystemEvents: '系统快照事件已折叠',
        pendingTitle: '待审批提案',
        pendingEmpty: '暂无待审批记忆',
        pendingError: '待审批记忆读取失败',
        addTitle: '新增记忆',
        addTitlePlaceholder: '标题',
        addContentPlaceholder: '输入希望 AI 记住的内容...',
        addButton: '添加',
        edit: '编辑',
        saveEdit: '保存修改',
        approve: '保存',
        deny: '拒绝',
      }
    : {
        title: 'Memory Console',
        subtitle: 'Review AI-proposed memories, search saved content, and confirm runs only read safe summaries.',
        search: 'Search memory',
        filters: 'Filters',
        hideFilters: 'Hide filters',
        loading: 'Loading',
        scopeType: 'Scope',
        scopeId: 'Scope ID',
        sourceThread: 'Source thread',
        sourceRun: 'Source run',
        sourceType: 'Source type',
        limit: 'Limit',
        any: 'Any',
        user: 'User',
        thread: 'Thread',
        manual: 'Manual',
        run: 'Run',
        includeDeleted: 'Include deleted',
        activeFilters: 'Filters active',
        clearFilters: 'Clear',
        emptyTitle: 'No saved memories',
        emptyBody: 'Memories appear here only after a run proposes them and they are approved.',
        errorTitle: 'Memory could not be loaded',
        detail: 'Memory detail',
        closeDetail: 'Close memory detail',
        scope: 'Scope',
        source: 'Source',
        status: 'Status',
        redaction: 'Redaction',
        created: 'Created',
        updated: 'Updated',
        deletedAt: 'Deleted at',
        confirmDelete: 'Confirm deletion',
        cancel: 'Cancel',
        deleteMemory: 'Delete memory',
        history: 'Memory history',
        historyEmpty: 'No memory history',
        historyError: 'Memory history could not be loaded',
        systemSnapshot: 'System snapshot',
        systemSnapshotLoaded: 'Runtime loaded a memory snapshot',
        hiddenSystemEvents: 'System snapshot events are folded',
        pendingTitle: 'Review Queue',
        pendingEmpty: 'No pending memories',
        pendingError: 'Pending memory could not be loaded',
        addTitle: 'Add memory',
        addTitlePlaceholder: 'Title',
        addContentPlaceholder: 'Enter something useful for AI to remember...',
        addButton: 'Add',
        edit: 'Edit',
        saveEdit: 'Save changes',
        approve: 'Save',
        deny: 'Decline',
      }
}

function auditTitle(item: MemoryAuditItem, locale: Locale) {
  if (locale === 'en') {
    const labels: Record<MemoryAuditItem['eventType'], string> = {
      memory_write_approved: 'Memory saved',
      memory_write_denied: 'Memory declined',
      memory_deleted: 'Memory deleted',
      memory_write_proposed: 'Memory proposed',
      memory_snapshot_loaded: 'Runtime loaded memory',
    }
    return labels[item.eventType]
  }
  const labels: Record<MemoryAuditItem['eventType'], string> = {
    memory_write_approved: '记忆已保存',
    memory_write_denied: '记忆已拒绝',
    memory_deleted: '记忆已删除',
    memory_write_proposed: '记忆待审批',
    memory_snapshot_loaded: '运行时读取记忆',
  }
  return labels[item.eventType]
}

function auditMetadata(item: MemoryAuditItem, locale: Locale) {
  const chunks = [
    item.scopeType ? (locale === 'zh' && item.scopeType === 'thread' ? '会话范围' : item.scopeType) : '',
    item.sourceType ? (locale === 'zh' ? `来源：${item.sourceType}` : `source: ${item.sourceType}`) : '',
    item.status ? (locale === 'zh' ? `状态：${item.status}` : `status: ${item.status}`) : '',
    item.redactionApplied ? redactionLabel(true, locale) : '',
    formatDate(item.occurredAt),
  ].filter(Boolean)
  return chunks.join(' · ')
}

export function MemoryPanel({
  entries,
  query,
  locale = 'zh',
  filters = {},
  loading = false,
  error = null,
  auditItems = [],
  auditLoading = false,
  auditError = null,
  writeProposals = [],
  proposalsLoading = false,
  proposalsError = null,
  detailEntry = null,
  detailLoading = false,
  detailError = null,
  pendingDeleteEntry = null,
  onQueryChange,
  onFiltersChange,
  onOpenDetail,
  onCloseDetail,
  onRequestDelete,
  onCancelDelete,
  onCreateMemory,
  onConfirmDelete,
  onApproveProposal,
  onUpdateProposal,
  onDenyProposal,
}: Props) {
  const copy = copyFor(locale)
  const scopeTypeOptions = [
    { key: '', label: copy.any },
    { key: 'user', label: copy.user },
    { key: 'thread', label: copy.thread },
  ]
  const sourceTypeOptions = [
    { key: 'any', label: copy.any },
    { key: 'manual', label: copy.manual },
    { key: 'thread', label: copy.thread },
    { key: 'run', label: copy.run },
  ]
  const [filtersOpen, setFiltersOpen] = useState(isActiveFilter(filters))
  const [editingProposalID, setEditingProposalID] = useState<string | null>(null)
  const [proposalDrafts, setProposalDrafts] = useState<Record<string, { title: string; summary: string }>>({})
  const [manualTitle, setManualTitle] = useState('')
  const [manualContent, setManualContent] = useState('')
  const updateFilters = (next: MemoryFilters) => onFiltersChange(next)
  const clearFilters = () => onFiltersChange({ limit: 20 })
  const visibleAuditItems = auditItems.filter((item) => auditEventTypes.includes(item.eventType))
  const latestSnapshot = auditItems.find((item) => item.eventType === 'memory_snapshot_loaded')
  const startProposalEdit = (proposal: MemoryWriteProposal) => {
    setEditingProposalID(proposal.id)
    setProposalDrafts((current) => ({ ...current, [proposal.id]: { title: proposal.title, summary: proposal.summary } }))
  }
  const updateProposalDraft = (proposal: MemoryWriteProposal, patch: Partial<{ title: string; summary: string }>) => {
    setProposalDrafts((current) => ({ ...current, [proposal.id]: { title: current[proposal.id]?.title ?? proposal.title, summary: current[proposal.id]?.summary ?? proposal.summary, ...patch } }))
  }
  const saveProposalDraft = (proposal: MemoryWriteProposal) => {
    const draft = proposalDrafts[proposal.id] ?? { title: proposal.title, summary: proposal.summary }
    onUpdateProposal?.(proposal, draft)
    setEditingProposalID(null)
  }
  const submitManualMemory = () => {
    const content = manualContent.trim()
    if (!content || !onCreateMemory) return
    onCreateMemory({ title: manualTitle.trim() || copy.addTitle, content, scopeType: 'user' })
    setManualTitle('')
    setManualContent('')
  }

  return (
    <section className="memory-panel" aria-label="Memory">
      <div className="memory-overview">
        <div>
          <h2>{copy.title}</h2>
          <p>{copy.subtitle}</p>
        </div>
        <div className="memory-stat-strip" aria-label="Memory status">
          <span><strong>{entries.length}</strong>{copy.title}</span>
          <span><strong>{visibleAuditItems.length}</strong>{copy.history}</span>
        </div>
      </div>

      <div className="memory-panel-toolbar">
        <input value={query} placeholder={copy.search} aria-label={copy.search} onChange={(event) => onQueryChange(event.target.value)} />
        <button className={filtersOpen ? 'selected' : undefined} type="button" onClick={() => setFiltersOpen(!filtersOpen)}>
          <Filter size={15} />
          {filtersOpen ? copy.hideFilters : copy.filters}
          <ChevronDown size={15} />
        </button>
        {loading && <span>{copy.loading}</span>}
      </div>

      {filtersOpen && <div className="memory-filter-grid" aria-label="Memory filters">
        <label>
          <span>{copy.scopeType}</span>
          <Select options={scopeTypeOptions} value={filters.scopeType ?? ''} onChange={(key) => updateFilters({ ...filters, scopeType: key as MemoryFilters['scopeType'] })} />
        </label>
        <label>
          <span>{copy.scopeId}</span>
          <input value={filters.scopeId ?? ''} onChange={(event) => updateFilters({ ...filters, scopeId: event.target.value })} />
        </label>
        <label>
          <span>{copy.sourceThread}</span>
          <input value={filters.sourceThreadId ?? ''} onChange={(event) => updateFilters({ ...filters, sourceThreadId: event.target.value })} />
        </label>
        <label>
          <span>{copy.sourceRun}</span>
          <input value={filters.sourceRunId ?? ''} onChange={(event) => updateFilters({ ...filters, sourceRunId: event.target.value })} />
        </label>
        <label>
          <span>{copy.sourceType}</span>
          <Select options={sourceTypeOptions} value={filters.sourceType ?? 'any'} onChange={(key) => updateFilters({ ...filters, sourceType: key as MemoryFilters['sourceType'] })} />
        </label>
        <label>
          <span>{copy.limit}</span>
          <input type="number" min="1" max="100" value={filters.limit ?? 20} onChange={(event) => updateFilters({ ...filters, limit: Number(event.target.value) || undefined })} />
        </label>
        <label className="memory-checkbox-filter">
          <Switch checked={Boolean(filters.includeTombstoned)} onChange={(checked) => updateFilters({ ...filters, includeTombstoned: checked })} />
          <span>{copy.includeDeleted}</span>
        </label>
        {isActiveFilter(filters) && <button className="memory-clear-filters" type="button" onClick={clearFilters}>{copy.clearFilters}</button>}
      </div>}
      {!filtersOpen && isActiveFilter(filters) && <p className="memory-filter-note">{copy.activeFilters}</p>}

      <section className="memory-manual-add" aria-label={copy.addTitle}>
        <div className="memory-history-head">
          <h3>{copy.addTitle}</h3>
        </div>
        <div className="memory-manual-add-form">
          <input value={manualTitle} placeholder={copy.addTitlePlaceholder} onChange={(event) => setManualTitle(event.target.value)} />
          <textarea value={manualContent} placeholder={copy.addContentPlaceholder} onChange={(event) => setManualContent(event.target.value)} />
          <button type="button" disabled={!manualContent.trim() || !onCreateMemory} onClick={submitManualMemory}>{copy.addButton}</button>
        </div>
      </section>

      <section className="memory-proposal-review" aria-label="Pending memory">
        <div className="memory-history-head">
          <h3>{copy.pendingTitle}</h3>
          {proposalsLoading && <span>{copy.loading}</span>}
        </div>
        {proposalsError ? (
          <div className="memory-empty-card"><strong>{copy.pendingError}</strong><p>{proposalsError}</p></div>
        ) : writeProposals.length === 0 ? (
          <p className="memory-empty">{copy.pendingEmpty}</p>
        ) : (
          <div className="memory-entry-list">
            {writeProposals.map((proposal) => (
              <article className="memory-entry pending" key={proposal.id}>
                {editingProposalID === proposal.id ? (
                  <div className="memory-proposal-editor">
                    <label>
                      <span>{copy.pendingTitle}</span>
                      <input value={proposalDrafts[proposal.id]?.title ?? proposal.title} onChange={(event) => updateProposalDraft(proposal, { title: event.target.value })} />
                    </label>
                    <label>
                      <span>{copy.subtitle}</span>
                      <textarea value={proposalDrafts[proposal.id]?.summary ?? proposal.summary} onChange={(event) => updateProposalDraft(proposal, { summary: event.target.value })} />
                    </label>
                    <small>{proposal.scopeType} · {metadataValue(proposal.scopeId)} · {redactionLabel(proposal.redactionApplied || proposal.safetyState === 'redacted', locale)} · {formatDate(proposal.createdAt)}</small>
                  </div>
                ) : (
                  <div>
                    <div className="memory-entry-head">
                      <strong>{proposal.title}</strong>
                      <span>{proposal.status}</span>
                    </div>
                    <p>{proposal.summary}</p>
                    <small>{proposal.scopeType} · {metadataValue(proposal.scopeId)} · {redactionLabel(proposal.redactionApplied || proposal.safetyState === 'redacted', locale)} · {formatDate(proposal.createdAt)}</small>
                  </div>
                )}
                <div className="memory-entry-actions">
                  {editingProposalID === proposal.id ? (
                    <>
                      <button type="button" aria-label={`${copy.saveEdit} ${proposal.title}`} onClick={() => saveProposalDraft(proposal)}>
                        <Check size={15} />
                      </button>
                      <button type="button" aria-label={`${copy.cancel} ${proposal.title}`} onClick={() => setEditingProposalID(null)}>
                        <X size={15} />
                      </button>
                    </>
                  ) : (
                    <button type="button" aria-label={`${copy.edit} ${proposal.title}`} onClick={() => startProposalEdit(proposal)}>
                      <Pencil size={15} />
                    </button>
                  )}
                  <button type="button" aria-label={`${copy.approve} ${proposal.title}`} onClick={() => onApproveProposal?.(proposal)}>
                    <Check size={15} />
                  </button>
                  <button type="button" aria-label={`${copy.deny} ${proposal.title}`} onClick={() => onDenyProposal?.(proposal)}>
                    <X size={15} />
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>

      {error && <p className="memory-error" role="alert">{error}</p>}
      <div className="memory-entry-list">
        {error ? (
          <div className="memory-empty-card"><strong>{copy.errorTitle}</strong><p>{error}</p></div>
        ) : entries.length === 0 ? (
          <div className="memory-empty-card"><strong>{copy.emptyTitle}</strong><p>{copy.emptyBody}</p></div>
        ) : entries.map((entry) => (
          <article className={`memory-entry ${entry.status}`} key={entry.id}>
            <div>
              <div className="memory-entry-head">
                <strong>{entry.title}</strong>
                <span>{statusLabel(entry.status, locale)}</span>
              </div>
              <p>{entry.summary}</p>
              <small>{scopeLabel(entry, locale)} · {entry.sourceType ?? 'manual'} · {redactionLabel(entry.redactionApplied || entry.safetyState === 'redacted', locale)} · {formatDate(entry.updatedAt)}</small>
            </div>
            <div className="memory-entry-actions">
              <button aria-label={`View ${entry.title}`} onClick={() => onOpenDetail(entry)}>
                <Eye size={15} />
              </button>
              <button aria-label={`Delete ${entry.title}`} onClick={() => onRequestDelete(entry)}>
                <Trash2 size={15} />
              </button>
            </div>
          </article>
        ))}
      </div>

      {(detailEntry || detailLoading || detailError) && (
        <aside className="memory-detail-panel" aria-label="Memory detail">
          <div className="memory-detail-head">
            <h3>{copy.detail}</h3>
            <button aria-label={copy.closeDetail} onClick={onCloseDetail}>
              <X size={15} />
            </button>
          </div>
          {detailLoading && <p className="memory-empty">{copy.loading}</p>}
          {detailError && <p className="memory-error" role="alert">{detailError}</p>}
          {detailEntry && (
            <>
              <strong>{detailEntry.title}</strong>
              <p>{detailEntry.summary}</p>
              <dl>
                <dt>{copy.scope}</dt><dd>{detailEntry.scopeType} · {metadataValue(detailEntry.scopeId)}</dd>
                <dt>{copy.source}</dt><dd>{metadataValue(detailEntry.sourceType)} · {metadataValue(detailEntry.sourceThreadId)} · {metadataValue(detailEntry.sourceRunId)}</dd>
                <dt>{copy.status}</dt><dd>{statusLabel(detailEntry.status, locale)}</dd>
                <dt>{copy.redaction}</dt><dd>{redactionLabel(detailEntry.redactionApplied || detailEntry.safetyState === 'redacted', locale)}</dd>
                <dt>{copy.created}</dt><dd>{formatDate(detailEntry.createdAt)}</dd>
                <dt>{copy.updated}</dt><dd>{formatDate(detailEntry.updatedAt)}</dd>
                {detailEntry.deletedAt && <><dt>{copy.deletedAt}</dt><dd>{formatDate(detailEntry.deletedAt)}</dd></>}
              </dl>
            </>
          )}
        </aside>
      )}

      {pendingDeleteEntry && (
        <div className="memory-delete-confirmation" role="alertdialog" aria-label="Confirm deletion">
          <div>
            <strong>{copy.confirmDelete}</strong>
            <p>{pendingDeleteEntry.title}</p>
            <small>{pendingDeleteEntry.scopeType} · {metadataValue(pendingDeleteEntry.scopeId)}</small>
            <small>{metadataValue(pendingDeleteEntry.sourceType)} · {metadataValue(pendingDeleteEntry.sourceThreadId)} · {metadataValue(pendingDeleteEntry.sourceRunId)}</small>
          </div>
          <button onClick={onCancelDelete}>{copy.cancel}</button>
          <button className="danger" onClick={() => onConfirmDelete?.(pendingDeleteEntry)}>{copy.deleteMemory}</button>
        </div>
      )}

      <section className="memory-history" aria-label="Memory history">
        <div className="memory-history-head">
          <div>
            <h3>{copy.history}</h3>
            {latestSnapshot && <small>{copy.systemSnapshotLoaded} · {formatDate(latestSnapshot.occurredAt)}</small>}
          </div>
          {auditLoading && <span>{copy.loading}</span>}
        </div>
        {auditError ? (
          <>
            <p className="memory-error" role="alert">{auditError}</p>
            <p className="memory-empty">{copy.historyError}</p>
          </>
        ) : visibleAuditItems.length === 0 ? (
          <p className="memory-empty">{copy.historyEmpty}</p>
        ) : (
          <div className="memory-audit-list">
            {visibleAuditItems.map((item) => (
              <article className="memory-audit-item" key={item.id}>
                <strong>{auditTitle(item, locale)}</strong>
                <small>{auditMetadata(item, locale)}</small>
              </article>
            ))}
          </div>
        )}
        {latestSnapshot && <p className="memory-filter-note">{copy.hiddenSystemEvents}</p>}
      </section>
    </section>
  )
}
