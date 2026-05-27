# Plan: M56 Memory Provider Card Selection

## Slice

Frontend-only UI polish inside `MemoryProviderFoundationPanel`.

## Design

- Use a three-column card grid on desktop and one column on narrow screens.
- Cards call the existing provider update action.
- Selected cards show a stronger border and radio-style dot.

## Validation

- SettingsView runtime tests.
- Web production build.
- Browser smoke already covers selecting Nowledge and opening Configure after the card change.
