# Research: M5.5 Settings Placeholder

## Decision: Deliver settings as an in-app desktop-style surface

**Rationale**: The user asked for a temporary M5.5 settings placeholder based on a desktop reference. Loomi already has a desktop-feeling web shell with sidebar, titlebar, mode tabs, and panels. An in-app settings view preserves the current workspace context and avoids introducing a separate preferences window before a broader desktop/runtime settings model exists.

**Alternatives considered**:

- Separate modal window: closer to some desktop apps, but would obscure the existing shell and add behavior not required for M5.5.
- Right drawer only: reuses existing right panel, but the reference design is a full settings area with category navigation and grouped cards.

## Decision: Actual settings are local-session controls only

**Rationale**: Current app behavior that can be safely changed from the browser is limited to UI/session state: default workspace mode for new/opening flows and mock runtime scenario for future mock runs. Persistence and account/team configuration are not yet specified and should not be implied by a placeholder slice.

**Alternatives considered**:

- Persist settings locally: useful later, but would introduce storage behavior not requested in the spec.
- Add backend settings persistence: out of scope and would require new data/API contracts.

## Decision: Provider/model gateway state is read-only in settings

**Rationale**: M5 provider credentials are backend-local by design. The settings surface can display data source mode, backend availability, and redacted provider capability, but it must not collect or show provider secrets in the browser.

**Alternatives considered**:

- Editable provider keys in Settings: rejected because it violates the M5 safety boundary.
- Hide provider state entirely: rejected because current configurable/readable model gateway state is useful to users.

## Decision: Future sections are explicit placeholders

**Rationale**: The reference design includes many settings categories, and the user asked to mock areas that are not currently real. Categories such as Providers, Connectors, Plugins, Skill, MCP, Notebook, Memory, Activity Recorder, Context, Safety, Tools, Routes, About, and Advanced should be visible as product IA previews while clearly marked as mock/not connected.

**Alternatives considered**:

- Only build General settings: too narrow for the requested placeholder surface.
- Implement all categories as functional: violates staged roadmap and would pull forward deferred platform complexity.

## Decision: Documentation and smoke validation are part of done

**Rationale**: The Loomi constitution requires runnable vertical slices and docs-site updates for non-trivial UI flows. M5.5 changes the settings UI flow and must update architecture/runbook/devlog/spec-kit references and run frontend/docs validation.

**Alternatives considered**:

- Skip docs because this is a placeholder: rejected because the placeholder defines visible product IA and safety boundaries.
