import { describe, expect, test } from 'bun:test'

const source = Bun.file(new URL('./SettingsView.tsx', import.meta.url)).text()
const themeSource = Bun.file(new URL('../styles/92-unified-workspace.css', import.meta.url)).text()

describe('SettingsView layout contract', () => {
  test('renders the required desktop-style landmarks', async () => {
    const text = await source

    expect(text).toContain('className="settings-shell"')
    expect(text).toContain('className="settings-sidebar"')
    expect(text).toContain('className="settings-content"')
    expect(text).toContain('settings-card routine-settings-card')
    expect(text).toContain('SettingsNavItem')
    expect(text).toContain('settings-nav-icon')
    expect(text).toContain('settings-nav-copy')
    expect(text).toContain('aria-label={`${category.label}: ${category.description}`}')
    expect(text).toContain('t.back')
  })

  test('settings navigator stays narrow and adapts without clipped side chrome', async () => {
    const css = await themeSource

    expect(css).toContain('grid-template-columns: minmax(172px, 190px) minmax(0, 1fr) !important;')
    expect(css).toContain('grid-template-columns: 22px minmax(0, 1fr) !important;')
    expect(css).toContain('.settings-back-button')
    expect(css).toContain('min-height: 34px !important;')
    expect(css).toContain('.settings-nav-group > span')
    expect(css).toContain('letter-spacing: 0 !important;')
    expect(css).toContain('@media (max-width: 720px)')
    expect(css).toContain('grid-template-columns: repeat(2, minmax(0, 1fr)) !important;')
    expect(css).not.toContain('.settings-nav-group button .settings-nav-copy > small')
  })

  test('keeps routine settings low-noise and reserves badges for non-working states', async () => {
    const text = await source

    expect(text).toContain('shouldShowSettingStatus')
    expect(text).toContain("status !== 'working'")
    expect(text).toContain('setting-status-badge')
    expect(text).toContain('setting-row-control')
    expect(text).toContain('t.working')
    expect(text).toContain('t.previewOnly')
  })

  test('keeps category status out of the settings navigation and page header chrome', async () => {
    const text = await source

    expect(text).not.toContain('settings-nav-state')
    expect(text).not.toContain('settings-category-pill')
    expect(text).not.toContain('shouldShowCategoryStatus')
    expect(text).not.toContain('categoryStatusLabel')
  })

  test('routine settings sections use compact rows instead of heavy cards', async () => {
    const css = await themeSource

    expect(css).toContain('.settings-card-stack')
    expect(css).toContain('width: min(100%, 580px) !important;')
    expect(css).toContain('max-width: 680px !important;')
    expect(css).toContain('.settings-card.routine-settings-card')
    expect(css).toContain('background: transparent !important;')
    expect(css).toContain('.settings-card-head')
    expect(css).toContain('border-bottom: 0 !important;')
    expect(css).toContain('padding: 0 0 10px !important;')
    expect(css).toContain('.settings-card .setting-row:first-of-type')
    expect(css).toContain('.setting-row + .setting-row')
    expect(css).toContain('grid-template-columns: minmax(0, 1fr) auto !important;')
    expect(css).toContain('gap: 12px !important;')
    expect(css).toContain('max-width: min(260px, 42vw) !important;')
    expect(css).toContain('.settings-segmented-control')
    expect(css).toContain('min-height: 32px !important;')
    expect(css).toContain('.settings-status-value')
    expect(css).toContain('background: transparent !important;')
  })

  test('providers settings expose a saved local provider action without draft-only copy', async () => {
    const text = await source

    expect(text).toContain("import { Button } from 'animal-island-ui'")
    expect(text).toContain('provider-save-modal-button')
    expect(text).toContain('provider-add-button')
    expect(text).toContain('provider-filter-tabs')
    expect(text).toContain('provider-card-test')
    expect(text).toContain('onSaveProvider')
    expect(text).not.toContain('不会保存')
    expect(text).not.toContain('not saved')
  })

  test('provider and web-search pages use compact settings rows instead of oversized cards', async () => {
    const text = await source
    const css = await themeSource

    expect(text).toContain('provider-row-list')
    expect(text).toContain('provider-row-main')
    expect(text).toContain('provider-row-actions')
    expect(text).toContain('web-search-provider-row')
    expect(text).toContain('web-search-footer-row')
    expect(text).toContain('web-search-save-button')
    expect(text).toContain('routine-settings-card web-search-card')
    expect(text).not.toContain('webSearch ? <code>web.search</code>')
    expect(css).toContain('.provider-row-list')
    expect(css).toContain('grid-template-columns: minmax(0, 1fr) !important;')
    expect(css).toContain('.provider-add-button')
    expect(css).toContain('border-radius: 0 !important;')
    expect(css).toContain('.web-search-provider-row')
    expect(css).toContain('.web-search-save-button')
    expect(css).toContain('grid-template-columns: minmax(96px, 132px) minmax(0, 1fr) !important;')
    expect(css).toContain('grid-template-columns: auto minmax(76px, auto) minmax(0, 1fr) !important;')
  })

  test('provider setup empty states and dialogs stay compact instead of card-like', async () => {
    const text = await source
    const css = await themeSource

    expect(text).toContain('className="provider-empty-state"')
    expect(text).toContain('onClick={onCloseAdd}')
    expect(text).toContain('onClick={(event) => event.stopPropagation()}')
    expect(css).toContain('.provider-management-toolbar')
    expect(css).toContain('.provider-empty-state')
    expect(css).toContain('grid-template-columns: minmax(0, 1fr) auto !important;')
    expect(css).toContain('.provider-capability-list')
    expect(css).toContain('.provider-modal-backdrop')
    expect(css).toContain('width: min(560px, calc(100vw - 48px)) !important;')
    expect(css).toContain('.provider-modal label')
    expect(css).toContain('grid-template-columns: 132px minmax(0, 1fr) !important;')
    expect(css).toContain('.provider-modal-actions')
    expect(css).toContain('.provider-console-card')
    expect(css).toContain('background: transparent !important;')
  })

  test('tool and skill catalog pages use quiet scan rows instead of badge-heavy cards', async () => {
    const text = await source
    const css = await themeSource

    expect(text).toContain('toolOperationalMeta')
    expect(text).toContain('toolSafetyChips')
    expect(text).toContain('tools-catalog-meta')
    expect(css).toContain('.tools-catalog-meta')
    expect(css).toContain('.tools-catalog-row')
    expect(css).toContain('grid-template-columns: minmax(0, 1fr) auto !important;')
    expect(css).toContain('.tools-catalog-badges span')
    expect(css).toContain('background: transparent !important;')
  })

  test('mcp settings use compact editable rows instead of a bulky form block', async () => {
    const text = await source
    const css = await themeSource

    expect(text).toContain('mcp-config-row-list')
    expect(text).toContain('mcp-config-row')
    expect(text).toContain('mcp-config-row-wide')
    expect(text).toContain('mcp-server-row')
    expect(text).toContain('mcp-server-actions')
    expect(text).toContain('mcp-save-button')
    expect(text).toContain('mcp-server-action-button')
    expect(css).toContain('.mcp-config-row-list')
    expect(css).toContain('.mcp-config-row')
    expect(css).toContain('.mcp-server-row')
    expect(css).toContain('.mcp-save-button')
    expect(css).toContain('.mcp-server-action-button.danger')
    expect(css).toContain('max-width: 580px !important;')
    expect(css).toContain('grid-template-columns: minmax(88px, 120px) minmax(0, 1fr) !important;')
    expect(css).toContain('grid-template-columns: auto minmax(0, 1fr) !important;')
    expect(css).toContain('@media (max-width: 900px)')
  })

  test('keeps settings content in normal document flow below the header', async () => {
    const css = await themeSource

    expect(css).toContain('position: static !important;')
    expect(css).toContain('margin: 0 0 30px !important;')
    expect(css).toContain('padding: 0 0 24px !important;')
    expect(css).toContain('.settings-content-header h1')
    expect(css).toContain('font-size: 26px !important;')
    expect(css).toContain('max-width: 560px !important;')
    expect(css).toContain('backdrop-filter: none !important;')
    expect(css).toContain('height: 100% !important;')
    expect(css).toContain('grid-template-columns: minmax(172px, 190px) minmax(0, 1fr) !important;')
    expect(css).toContain('overflow-x: hidden !important;')
  })

  test('splits appearance settings from general workspace defaults', async () => {
    const text = await source
    const catalog = await Bun.file(new URL('./settingsCatalog.ts', import.meta.url)).text()

    expect(catalog).toContain("| 'appearance'")
    expect(catalog).toContain("appearance: { zh: { label: '外观'")
    expect(catalog).not.toContain("{ id: 'theme', label: 'Theme'")
    expect(text).toContain('const isAppearance')
    expect(text).toContain('{isAppearance &&')
    expect(text).toContain('routine-settings-card')
    expect(text).toContain('t.displayPreferences')
    expect(text).toContain('t.themeResolved')
    expect(text).toContain('role="radiogroup"')
    expect(text).toContain('role="radio"')
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

  test('memory provider settings use row-style choices instead of provider cards', async () => {
    const css = await themeSource

    expect(css).toContain('.memory-provider-panel')
    expect(css).toContain('max-width: 680px !important;')
    expect(css).toContain('.memory-provider-choice-grid')
    expect(css).toContain('grid-template-columns: minmax(0, 1fr) !important;')
    expect(css).toContain('grid-template-columns: 18px minmax(92px, auto) minmax(0, 1fr) !important;')
    expect(css).toContain('.memory-provider-status-row')
    expect(css).toContain('grid-template-columns: minmax(0, 1fr) auto !important;')
    expect(css).toContain('.memory-provider-choice-grid button.selected .memory-provider-choice-dot')
  })

  test('runtime memory snapshot uses the same row grammar', async () => {
    const css = await themeSource

    expect(css).toContain('.memory-snapshot-panel')
    expect(css).toContain('.memory-snapshot-grid')
    expect(css).toContain('.memory-snapshot-card')
    expect(css).toContain('max-height: 88px !important;')
    expect(css).toContain('.memory-snapshot-hit-list button')
  })
})
