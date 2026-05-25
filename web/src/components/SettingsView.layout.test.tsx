import { describe, expect, test } from 'bun:test'

const source = Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()

describe('SettingsView layout contract', () => {
  test('renders the required desktop-style landmarks', async () => {
    const text = await source

    expect(text).toContain('className="settings-shell"')
    expect(text).toContain('className="settings-sidebar"')
    expect(text).toContain('className="settings-content"')
    expect(text).toContain('className="settings-card"')
    expect(text).toContain('t.back')
  })

  test('distinguishes rows with status badges and right-aligned controls', async () => {
    const text = await source

    expect(text).toContain('setting-status-badge')
    expect(text).toContain('setting-row-control')
    expect(text).toContain('t.working')
    expect(text).toContain('t.previewOnly')
  })

  test('providers settings expose a saved local provider action without draft-only copy', async () => {
    const text = await source

    expect(text).toContain('provider-save-button')
    expect(text).toContain('onSaveProvider')
    expect(text).not.toContain('不会保存')
    expect(text).not.toContain('not saved')
  })

  test('memory settings wires grounded filters detail and confirmed delete actions', async () => {
    const text = await source

    expect(text).toContain('memoryFilters')
    expect(text).toContain('onMemoryFiltersChange')
    expect(text).toContain('memoryDetail')
    expect(text).toContain('onOpenMemoryDetail')
    expect(text).toContain('onConfirmDeleteMemoryEntry')
    expect(text).not.toContain('onDeleteMemoryEntry={(entryId) => void deleteMemoryEntry(entryId)}')
  })

  test('memory settings wires real audit history state', async () => {
    const text = await source

    expect(text).toContain('memoryAuditItems')
    expect(text).toContain('memoryAuditError')
    expect(text).toContain('memoryAuditLoading')
    expect(text).toContain('auditItems={memoryAuditItems}')
  })
})
