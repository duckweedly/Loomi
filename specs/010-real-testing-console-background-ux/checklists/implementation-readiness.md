# Implementation Readiness Checklist: Real Testing Console & Background UX

**Purpose**: Validate that the M6.5 requirements are complete, clear, consistent, and ready for implementation before code changes begin.  
**Created**: 2026-05-24  
**Feature**: `010-real-testing-console-background-ux`  
**Audience**: Implementer + PR reviewer  
**Depth**: Standard pre-implementation gate

## Requirement Completeness

- [x] CHK001 Are Provider Test Console requirements complete for configured provider identity, family, model, base URL, status, message, and last-check timing if available? [Completeness, Spec §FR-001–FR-004]
- [x] CHK002 Are requirements defined for both configured backend providers and browser local draft provider fields? [Completeness, Spec §FR-001, §FR-007, §FR-008]
- [x] CHK003 Are requirements complete for no-provider empty state, including environment guidance and restart expectation? [Completeness, Spec §FR-006]
- [x] CHK004 Are requirements defined for provider connection checking, success, and failed states? [Completeness, Spec §FR-003–FR-004]
- [x] CHK005 Are Chat/Composer requirements complete for mock, `real_api`, and `model_gateway` runtime capabilities? [Completeness, Spec §FR-009–FR-013]
- [x] CHK006 Are Background tasks requirements complete for selected Chat run job and empty state? [Completeness, Spec §FR-015, Clarifications]
- [x] CHK007 Are requirements defined for all required job states: queued, leased, retrying, recovering, completed, failed, cancelled, and dead? [Completeness, Spec §FR-016]
- [x] CHK008 Are Timeline requirements complete for every required M6 event type? [Completeness, Spec §FR-021–FR-031]
- [x] CHK009 Are Composer state requirements complete for queued, running, retrying, recovering, stopped, failed, closed, cancelled, and provider unavailable? [Completeness, Spec §FR-032–FR-036]
- [x] CHK010 Are documentation requirements complete for architecture docs, runbooks, devlog, roadmap, and validation reporting? [Completeness, Spec §FR-042–FR-046]

## Requirement Clarity

- [x] CHK011 Is “Test connection” clearly specified as a per-provider UI action, including the fallback behavior when backend readiness is aggregate-only? [Clarity, Clarifications]
- [x] CHK012 Is “configured provider” clearly defined as backend-derived and authoritative for real model calls? [Clarity, Spec §Key Entities]
- [x] CHK013 Is “local draft” clearly defined as browser-session-only, unsaved, and non-authoritative for real calls? [Clarity, Spec §FR-007–FR-008]
- [x] CHK014 Is “provider unavailable” defined clearly enough to distinguish it from queued, running, failed, and stopped generation states? [Clarity, Spec §FR-010–FR-012]
- [x] CHK015 Is the Background tasks “current” selection priority explicitly defined and unambiguous? [Clarity, Clarifications]
- [x] CHK016 Is `recovering` clearly defined as either backend job status or derived read-only UI observation state? [Clarity, Clarifications]
- [x] CHK017 Is the Composer retry/regenerate availability matrix precise enough to avoid implementer interpretation? [Clarity, Contract §Composer State Matrix]
- [x] CHK018 Are disabled retry/regenerate reasons explicitly enumerated? [Clarity, Spec §FR-036]
- [x] CHK019 Is “sanitized provider error” defined with enough examples to prevent API key or bearer token leakage? [Clarity, Spec §FR-005]
- [x] CHK020 Are event fallback requirements clear for unknown or future worker/job events? [Clarity, Contract §M6 Event Labels]

## Requirement Consistency

- [x] CHK021 Are provider readiness requirements consistent between Settings, Chat/Composer, and quickstart flows? [Consistency, Spec §FR-001–FR-013]
- [x] CHK022 Are local draft requirements consistent across spec, plan, data model, and UI contract? [Consistency, Spec §FR-007–FR-008]
- [x] CHK023 Are Background tasks job states consistent across spec, data model, UI contract, and tasks? [Consistency, Spec §FR-016]
- [x] CHK024 Are Timeline event names consistent with M6 worker job event names and not renamed into incompatible terms? [Consistency, Spec §FR-021–FR-031]
- [x] CHK025 Are Composer state labels consistent with provider unavailable behavior and runtime capability rules? [Consistency, Spec §FR-032–FR-036]
- [x] CHK026 Are English and Chinese terminology requirements aligned for Provider Test Console, Background tasks, runtime, worker, job, diagnostics, provider unavailable, backend capability, and Composer states? [Consistency, Spec §FR-037–FR-041]
- [x] CHK027 Are docs-site targets consistent between plan, tasks, and project CLAUDE.md documentation requirements? [Consistency, Spec §FR-042–FR-046]
- [x] CHK028 Are M7 exclusions consistent across spec, plan, tasks, and docs requirements? [Consistency, Non-Goals]

## Acceptance Criteria Quality

- [x] CHK029 Are success criteria measurable without relying on implementation details like component names or internal APIs? [Measurability, Spec §Success Criteria]
- [x] CHK030 Is the “determine whether providers are configured within 10 seconds” criterion objectively verifiable from the UI requirement text? [Measurability, Spec §SC-001]
- [x] CHK031 Is the provider failure secret-safety criterion measurable with explicit forbidden secret patterns or examples? [Measurability, Spec §SC-003]
- [x] CHK032 Is “users are never shown a generating state when provider readiness is unavailable” testable against defined Composer states? [Measurability, Spec §SC-004]
- [x] CHK033 Is “mock-mode chat remains usable” defined precisely enough to avoid blocking mock mode through provider readiness? [Measurability, Spec §SC-005]
- [x] CHK034 Is the Background tasks observability success criterion tied to defined snapshot fields and state labels? [Measurability, Spec §SC-006]
- [x] CHK035 Is Timeline event readability tied to the explicit M6 event list? [Measurability, Spec §SC-007]
- [x] CHK036 Is Composer control correctness tied to the Composer State Matrix rather than subjective “clear” behavior? [Measurability, Spec §SC-008]

## Scenario Coverage

- [x] CHK037 Are primary Settings scenarios specified for configured providers, no providers, successful provider check, failed provider check, and local draft edits? [Coverage, Spec §Scenarios 1–5]
- [x] CHK038 Are Chat scenarios specified for provider unavailable real modes and mock mode unaffected behavior? [Coverage, Spec §Scenarios 6–7]
- [x] CHK039 Are Background tasks scenarios specified for current run/job and no-task empty state? [Coverage, Spec §Scenarios 8–9]
- [x] CHK040 Are Timeline scenarios specified for all required M6 events and unknown/future events? [Coverage, Spec §Scenario 10]
- [x] CHK041 Are Composer scenarios specified for every required run/job state and retry/regenerate condition? [Coverage, Spec §Scenario 11]
- [x] CHK042 Are local real testing runbook scenarios specified for env startup, provider test, real message, queued run, worker job, SSE timeline, cancel, recovery, and failure? [Coverage, Spec §FR-044]

## Edge Case Coverage

- [x] CHK043 Are requirements defined for provider check failures that include secret-like text? [Edge Case, Spec §Edge Cases]
- [x] CHK044 Are requirements defined for a local draft that differs from backend configured providers? [Edge Case, Spec §Edge Cases]
- [x] CHK045 Are requirements defined for real provider-dependent modes when no provider is configured? [Edge Case, Spec §FR-010–FR-012]
- [x] CHK046 Are requirements defined for mock mode when no real provider is configured? [Edge Case, Spec §FR-013]
- [x] CHK047 Are requirements defined for a job that exists without selected Chat run context? [Edge Case, Clarifications]
- [x] CHK048 Are requirements defined for a run that exists without job event history? [Edge Case, Spec §Edge Cases]
- [x] CHK049 Are requirements defined for unavailable worker diagnostics? [Edge Case, Spec §Edge Cases]
- [x] CHK050 Are requirements defined for retry exhausted or dead job state? [Edge Case, Spec §FR-016, §FR-027]
- [x] CHK051 Are requirements defined for cancellation while queued or leased? [Edge Case, Spec §FR-028]
- [x] CHK052 Are requirements defined for out-of-order worker/job events preserving stream order? [Edge Case, Clarifications]
- [x] CHK053 Are requirements defined for locale changes while a status is active? [Edge Case, Spec §Edge Cases]

## Non-Functional Requirements

- [x] CHK054 Are security/privacy requirements explicit for API key, bearer token, and provider credential non-disclosure? [Security, Spec §FR-005]
- [x] CHK055 Are requirements explicit that M6.5 does not persist API keys or introduce final settings secret storage? [Security, Non-Goals]
- [x] CHK056 Are observability requirements defined for worker diagnostics, latest events, Timeline groups, and raw event identity preservation? [Observability, Spec §FR-017–FR-031]
- [x] CHK057 Are localization requirements complete for English and Chinese copy across all new user-visible states? [Localization, Spec §FR-037–FR-041]
- [x] CHK058 Are accessibility expectations for new buttons, disabled states, CTA, status labels, and panels specified or intentionally deferred? [Gap, Accessibility]
- [x] CHK059 Are performance expectations for provider check loading and panel rendering specified or intentionally deferred? [Gap, Performance]
- [x] CHK060 Are deterministic automated test fixture requirements defined for retry, recovery, cancellation, failed, and dead states? [Reliability, Clarifications]
- [x] CHK061 Are documentation-as-done requirements explicit enough to satisfy project instructions? [Documentation, Spec §FR-042–FR-046]

## Dependencies & Assumptions

- [x] CHK062 Are assumptions documented for reusing existing M5.5/M6 provider readiness, runtime mode, worker job, and timeline surfaces? [Assumption, Spec §Assumptions]
- [x] CHK063 Are dependencies on existing backend provider env configuration clearly stated? [Dependency, Spec §Assumptions]
- [x] CHK064 Are constraints documented for not rewriting the M6 job data model unless a real bug is discovered? [Constraint, Non-Goals]
- [x] CHK065 Are constraints documented for not rewriting the frontend runtime architecture? [Constraint, Non-Goals]
- [x] CHK066 Are constraints documented for not adding a large i18n dependency? [Constraint, Non-Goals]
- [x] CHK067 Are validation command expectations documented without hardcoding unverified script names? [Assumption, Plan §Validation Commands]
- [x] CHK068 Are docs-site build requirements documented with Bun as required by project instructions? [Dependency, Plan §Validation Commands]

## Ambiguities & Conflicts

- [x] CHK069 Are all earlier analysis ambiguities resolved in the spec clarifications section? [Ambiguity, Clarifications]
- [x] CHK070 Are there any remaining conflicting definitions of `recovering` across spec, data model, contract, and tasks? [Conflict, Clarifications]
- [x] CHK071 Are there any remaining conflicting definitions of “current background task” across spec, plan, and tasks? [Conflict, Clarifications]
- [x] CHK072 Are there any remaining references implying local draft affects backend provider calls? [Conflict, Spec §FR-007–FR-008]
- [x] CHK073 Are there any remaining requirements that imply M7 tool call, approval, or tool execution protocol work? [Conflict, Non-Goals]
- [x] CHK074 Are there any subjective terms like “clear”, “readable”, or “polish” without concrete labels, state tables, or acceptance criteria? [Ambiguity]
