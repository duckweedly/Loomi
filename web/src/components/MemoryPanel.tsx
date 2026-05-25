import { Eye, Trash2, X } from 'lucide-react'
import type { MemoryAuditItem, MemoryEntry, MemoryFilters } from '../domain'

type Props = {
  entries: MemoryEntry[]
  query: string
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
  'memory_write_proposed',
  'memory_write_approved',
  'memory_write_denied',
  'memory_deleted',
  'memory_snapshot_loaded',
]

function metadataValue(value?: string | null) {
  return value && value.trim() ? value : 'none'
}

function redactionLabel(redactionApplied: boolean) {
  return redactionApplied ? 'Redacted' : 'Not redacted'
}

function entryFilterPayload(entry: MemoryEntry) {
  return [
    entry.scopeType,
    metadataValue(entry.scopeId),
    metadataValue(entry.sourceThreadId),
    metadataValue(entry.sourceRunId),
    metadataValue(entry.sourceType),
    entry.status,
    redactionLabel(entry.redactionApplied),
  ].join(' · ')
}

function auditMetadata(item: MemoryAuditItem) {
  return [
    item.eventType,
    metadataValue(item.memoryEntryId),
    metadataValue(item.memoryProposalId),
    metadataValue(item.threadId),
    metadataValue(item.runId),
    metadataValue(item.scopeType),
    metadataValue(item.sourceType),
    metadataValue(item.status),
    item.redactionApplied ? 'redacted' : 'not redacted',
    item.occurredAt,
  ].join(' · ')
}

export function MemoryPanel({
  entries,
  query,
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
  const updateFilters = (next: MemoryFilters) => onFiltersChange(next)

  return (
    <section className="memory-panel" aria-label="Memory">
      <div className="memory-panel-toolbar">
        <input value={query} placeholder="Search memory" aria-label="Search memory" onChange={(event) => onQueryChange(event.target.value)} />
        {loading && <span>Loading</span>}
      </div>

      <div className="memory-filter-grid" aria-label="Memory filters">
        <label>
          <span>Scope type</span>
          <select value={filters.scopeType ?? ''} onChange={(event) => updateFilters({ ...filters, scopeType: event.target.value as MemoryFilters['scopeType'] })}>
            <option value="">Any</option>
            <option value="user">user</option>
            <option value="thread">thread</option>
          </select>
        </label>
        <label>
          <span>Scope id</span>
          <input value={filters.scopeId ?? ''} onChange={(event) => updateFilters({ ...filters, scopeId: event.target.value })} />
        </label>
        <label>
          <span>Source thread</span>
          <input value={filters.sourceThreadId ?? ''} onChange={(event) => updateFilters({ ...filters, sourceThreadId: event.target.value })} />
        </label>
        <label>
          <span>Source run</span>
          <input value={filters.sourceRunId ?? ''} onChange={(event) => updateFilters({ ...filters, sourceRunId: event.target.value })} />
        </label>
        <label>
          <span>Source type</span>
          <select value={filters.sourceType ?? 'any'} onChange={(event) => updateFilters({ ...filters, sourceType: event.target.value as MemoryFilters['sourceType'] })}>
            <option value="any">any</option>
            <option value="manual">manual</option>
            <option value="thread">thread</option>
            <option value="run">run</option>
          </select>
        </label>
        <label>
          <span>Limit</span>
          <input type="number" min="1" max="100" value={filters.limit ?? 20} onChange={(event) => updateFilters({ ...filters, limit: Number(event.target.value) || undefined })} />
        </label>
        <label className="memory-checkbox-filter">
          <input type="checkbox" checked={Boolean(filters.includeTombstoned)} onChange={(event) => updateFilters({ ...filters, includeTombstoned: event.target.checked })} />
          <span>Include deleted</span>
        </label>
      </div>

      {error && <p className="memory-error" role="alert">{error}</p>}
      <div className="memory-entry-list">
        {error ? (
          <p className="memory-empty">Memory could not be loaded</p>
        ) : entries.length === 0 ? (
          <p className="memory-empty">No memory entries</p>
        ) : entries.map((entry) => (
          <article className={`memory-entry ${entry.status}`} key={entry.id}>
            <div>
              <strong>{entry.title}</strong>
              <p>{entry.summary}</p>
              <small>{entryFilterPayload(entry)}</small>
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
            <h3>Memory detail</h3>
            <button aria-label="Close memory detail" onClick={onCloseDetail}>
              <X size={15} />
            </button>
          </div>
          {detailLoading && <p className="memory-empty">Loading detail</p>}
          {detailError && <p className="memory-error" role="alert">{detailError}</p>}
          {detailEntry && (
            <>
              <strong>{detailEntry.title}</strong>
              <p>{detailEntry.summary}</p>
              <dl>
                <dt>Scope</dt><dd>{detailEntry.scopeType} · {metadataValue(detailEntry.scopeId)}</dd>
                <dt>Source</dt><dd>{metadataValue(detailEntry.sourceType)} · {metadataValue(detailEntry.sourceThreadId)} · {metadataValue(detailEntry.sourceRunId)}</dd>
                <dt>Status</dt><dd>{detailEntry.status === 'tombstoned' ? 'Deleted' : detailEntry.status}</dd>
            <dt>Redaction</dt><dd>{detailEntry.redactionApplied || detailEntry.safetyState === 'redacted' ? 'Redacted' : 'Not redacted'}</dd>
                <dt>Created</dt><dd>{detailEntry.createdAt}</dd>
                <dt>Updated</dt><dd>{detailEntry.updatedAt}</dd>
                {detailEntry.deletedAt && <><dt>Deleted at</dt><dd>{detailEntry.deletedAt}</dd></>}
              </dl>
            </>
          )}
        </aside>
      )}

      {pendingDeleteEntry && (
        <div className="memory-delete-confirmation" role="alertdialog" aria-label="Confirm deletion">
          <div>
            <strong>Confirm deletion</strong>
            <p>{pendingDeleteEntry.title}</p>
            <small>{pendingDeleteEntry.scopeType} · {metadataValue(pendingDeleteEntry.scopeId)}</small>
            <small>{metadataValue(pendingDeleteEntry.sourceType)} · {metadataValue(pendingDeleteEntry.sourceThreadId)} · {metadataValue(pendingDeleteEntry.sourceRunId)}</small>
          </div>
          <button onClick={onCancelDelete}>Cancel</button>
          <button className="danger" onClick={() => onConfirmDelete?.(pendingDeleteEntry)}>Delete memory</button>
        </div>
      )}

      <section className="memory-history" aria-label="Memory history">
        <div className="memory-history-head">
          <h3>Memory history</h3>
          {auditLoading && <span>Loading</span>}
        </div>
        {auditError ? (
          <>
            <p className="memory-error" role="alert">{auditError}</p>
            <p className="memory-empty">Memory history could not be loaded</p>
          </>
        ) : auditItems.length === 0 ? (
          <p className="memory-empty">No memory history</p>
        ) : (
          <div className="memory-audit-list">
            {auditItems.filter((item) => auditEventTypes.includes(item.eventType)).map((item) => (
              <article className="memory-audit-item" key={item.id}>
                <strong>{item.eventType}</strong>
                <small>{auditMetadata(item)}</small>
              </article>
            ))}
          </div>
        )}
      </section>
    </section>
  )
}
