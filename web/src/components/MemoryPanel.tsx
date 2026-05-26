import { ChevronDown, Eye, Filter, Trash2, X } from 'lucide-react'
import { useState } from 'react'
import type { MemoryAuditItem, MemoryEntry, MemoryFilters } from '../domain'
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
  onConfirmDelete?: (entry: MemoryEntry) => void
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
        title: '已保存记忆',
        subtitle: '只展示经过批准的本地记忆，原始对话和工具输出不会出现在这里。',
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
      }
    : {
        title: 'Saved Memory',
        subtitle: 'Only approved local memories are shown. Raw chat and tool output stay out of this view.',
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
  onConfirmDelete,
}: Props) {
  const copy = copyFor(locale)
  const [filtersOpen, setFiltersOpen] = useState(isActiveFilter(filters))
  const updateFilters = (next: MemoryFilters) => onFiltersChange(next)
  const clearFilters = () => onFiltersChange({ limit: 20 })
  const visibleAuditItems = auditItems.filter((item) => auditEventTypes.includes(item.eventType))
  const latestSnapshot = auditItems.find((item) => item.eventType === 'memory_snapshot_loaded')

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
          <select value={filters.scopeType ?? ''} onChange={(event) => updateFilters({ ...filters, scopeType: event.target.value as MemoryFilters['scopeType'] })}>
            <option value="">{copy.any}</option>
            <option value="user">{copy.user}</option>
            <option value="thread">{copy.thread}</option>
          </select>
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
          <select value={filters.sourceType ?? 'any'} onChange={(event) => updateFilters({ ...filters, sourceType: event.target.value as MemoryFilters['sourceType'] })}>
            <option value="any">{copy.any}</option>
            <option value="manual">{copy.manual}</option>
            <option value="thread">{copy.thread}</option>
            <option value="run">{copy.run}</option>
          </select>
        </label>
        <label>
          <span>{copy.limit}</span>
          <input type="number" min="1" max="100" value={filters.limit ?? 20} onChange={(event) => updateFilters({ ...filters, limit: Number(event.target.value) || undefined })} />
        </label>
        <label className="memory-checkbox-filter">
          <input type="checkbox" checked={Boolean(filters.includeTombstoned)} onChange={(event) => updateFilters({ ...filters, includeTombstoned: event.target.checked })} />
          <span>{copy.includeDeleted}</span>
        </label>
        {isActiveFilter(filters) && <button className="memory-clear-filters" type="button" onClick={clearFilters}>{copy.clearFilters}</button>}
      </div>}
      {!filtersOpen && isActiveFilter(filters) && <p className="memory-filter-note">{copy.activeFilters}</p>}

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
