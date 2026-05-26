# Research: M11 Tool Catalog Visibility

## Decision: Static read-only catalog

**Rationale**: The current tools are compiled internal tools. A static catalog is enough to make the surface visible without introducing plugin loading or permission editing.

**Alternatives rejected**: editable permissions, dynamic plugin registry, MCP-derived tools.

## Decision: Settings Tools becomes read-only

**Rationale**: Users need visibility before controls. Read-only status avoids implying auto-approval or permission changes.

**Alternatives rejected**: keeping placeholder copy; adding toggles before a permission model exists.
