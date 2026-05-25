import { Trash2 } from 'lucide-react'
import type { MemoryEntry } from '../domain'

type Props = {
  entries: MemoryEntry[]
  query: string
  loading?: boolean
  error?: string | null
  onQueryChange: (query: string) => void
  onDelete: (entryId: string) => void
}

export function MemoryPanel({ entries, query, loading = false, error = null, onQueryChange, onDelete }: Props) {
  return (
    <section className="memory-panel" aria-label="Memory">
      <div className="memory-panel-toolbar">
        <input value={query} placeholder="Search memory" aria-label="Search memory" onChange={(event) => onQueryChange(event.target.value)} />
        {loading && <span>Loading</span>}
      </div>
      {error && <p className="memory-error" role="alert">{error}</p>}
      <div className="memory-entry-list">
        {error ? (
          <p className="memory-empty">Memory could not be loaded</p>
        ) : entries.length === 0 ? (
          <p className="memory-empty">No memory entries</p>
        ) : entries.map((entry) => (
          <article className="memory-entry" key={entry.id}>
            <div>
              <strong>{entry.title}</strong>
              <p>{entry.summary}</p>
              <small>{entry.scopeType} · {entry.updatedAt}</small>
            </div>
            <button aria-label={`Delete ${entry.title}`} onClick={() => onDelete(entry.id)}>
              <Trash2 size={15} />
            </button>
          </article>
        ))}
      </div>
    </section>
  )
}
