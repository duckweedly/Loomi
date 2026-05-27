# Plan: M55 Memory Provider Config Modal

## Slice

This is a frontend-only interaction slice. Reuse the existing `MemoryProviderFoundationPanel`, existing `onUpdateMemoryProvider` action, and existing provider status model.

## Design

- Add local modal state to the memory provider panel.
- Render Settings > Memory provider fields only in the modal for Nowledge/OpenViking.
- Keep the main surface dense: toggles, provider segmented control, status, diagnostic, and Configure button.
- Preserve the existing safe input and detection behavior.

## Validation

- Existing SettingsView runtime tests.
- Web production build.
- Browser smoke for opening Settings > Memory, selecting Nowledge, opening Configure, and detecting local instance.
