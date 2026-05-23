# Research: M3 Auth, Thread, and Message

## Decision: Continue with Go standard library HTTP for M3 product routes

**Rationale**: M2 already established a small `net/http` service with health/readiness handlers. M3 adds a compact JSON API for identity, threads, and messages, which fits the standard library route pattern and keeps routing behavior transparent for learning.

**Alternatives considered**:

- Gin/Echo/Fiber: rejected because M3 does not need middleware stacks, route groups, or framework-specific abstractions.
- GraphQL: rejected because M3 needs straightforward local CRUD and smokeable curl contracts.
- Separate API process for product data: rejected because M3 is a vertical slice on top of the M2 service boundary.

## Decision: Represent local identity as a durable fixed user row, not a session

**Rationale**: The spec requires ownership for all product data without production authentication. A fixed local user boundary gives threads/messages a durable owner while avoiding a session table, request header user switching, or anonymous records that would need retrofits later. The API and seed command should ensure the local user exists with an idempotent upsert.

**Alternatives considered**:

- Session table: rejected by clarification and unnecessary for one local development identity.
- User id request header: rejected because it would create fake multi-user semantics and weaken the fixed-identity boundary.
- Migration-inserted demo user plus demo data: rejected because migrations should remain schema-focused and seed/demo behavior must be explicit.

## Decision: Add schema version `000002` for users, threads, and messages

**Rationale**: M2 intentionally created only the migration baseline. M3 is the first milestone that owns product data, so it should create the business tables in one reversible migration pair. Keeping users, threads, and messages together makes rollback/reapply smoke checks match the feature boundary.

**Alternatives considered**:

- Split users, threads, and messages into separate migrations: rejected because M3 ships them as one minimal product-data slice and they cannot be demonstrated independently.
- Create future run/event/tool tables now: rejected because those belong to M4 and later.
- Use PostgreSQL UUID generation extensions: rejected because local text ids avoid an extension dependency and are sufficient for one local API boundary.

## Decision: Use pgxpool with explicit transactions for user message creation

**Rationale**: pgxpool is already present from M2. Message creation must insert or return an existing idempotent user message and update the parent thread's `updated_at` in the same user-visible operation. A repository/service transaction keeps this atomic without adding an ORM.

**Alternatives considered**:

- ORM: rejected because M3 has three tables and benefits from visible SQL while learning the data model.
- No transaction: rejected because message insert and parent thread update could diverge.
- Frontend-only duplicate suppression: rejected because retries and double-clicks must be safe at the durable boundary.

## Decision: Enforce optional client message idempotency with a partial unique index

**Rationale**: `client_message_id` is optional, so messages without it should always create new records. When provided, uniqueness should be scoped to the same thread and local identity. A partial unique index on `(thread_id, user_id, client_message_id) where client_message_id is not null` expresses exactly that database invariant.

**Alternatives considered**:

- Global client message id uniqueness: rejected because clients may reuse keys across unrelated threads.
- Required client message ids: rejected because the spec makes them optional.
- In-memory duplicate tracking: rejected because refresh/restart would lose the guarantee.

## Decision: Use `/v1` JSON endpoints with structured API errors

**Rationale**: `/v1` gives the frontend and docs a stable contract without implying production deployment. Error responses need stable codes, human-readable messages, and request ids, so handlers should centralize an error envelope and avoid leaking lower-level database details.

**Alternatives considered**:

- Reuse `/healthz` and `/readyz` style root-only payloads for errors: rejected because product errors need stable machine-readable codes.
- Expose database errors directly: rejected because they may leak internal details or secrets.
- Hide request ids from clients: rejected because local diagnosis and docs require correlatable diagnostics.

## Decision: Configure frontend real/mock switching with `VITE_LOOMI_API_BASE_URL`

**Rationale**: Vite exposes build-time environment variables with the `VITE_` prefix, and the current frontend already routes data through `web/src/apiClient.ts`. Keeping mock behavior when the variable is absent preserves the M1 demonstration shell. When configured, the real API client should surface fetch/API errors because silent mock fallback would hide backend readiness failures.

**Alternatives considered**:

- Runtime user toggle in the UI: rejected because M3 only needs developer configuration, not a product feature.
- Always require the backend: rejected because M1 mock shell must remain usable without a backend.
- Auto-fallback from real API to mock data: rejected by clarification because it would mask configured-backend failures.

## Decision: Follow React effect cleanup guidance for frontend API loads

**Rationale**: React documentation recommends ignoring stale responses in Effects that synchronize with external systems such as network requests. Real API mode introduces asynchronous thread/message loads; cleanup guards keep stale responses from overwriting state when the selected thread changes quickly.

**Alternatives considered**:

- Ignore cleanup because local API is fast: rejected because development React behavior and quick navigation can still produce stale responses.
- Add a data-fetching library: rejected because M3 can meet requirements with React state, Effects, and the existing client seam.

## Decision: Provide an explicit idempotent seed command

**Rationale**: The spec requires demo data to be explicit and absent from migrations. `go run ./cmd/loomi-seed` can create the fixed local user, a deterministic demo thread, and one deterministic user message through the same product-data boundary used by the API.

**Alternatives considered**:

- Insert demo rows from migration `000002`: rejected because migrations must not contain demo data.
- Frontend-only mock seed: rejected because it would not validate real durable data.
- Destructive reset seed: rejected for M3 because the feature only requires seed creation, not cleanup or fixture reset.

## Decision: M3 readiness requires a clean schema version of at least 2

**Rationale**: M2 readiness only proved the service baseline. M3 must report not ready when the database is migrated only through M2 and ready only after the product-data schema is present and clean.

**Alternatives considered**:

- Keep readiness at version >= 1: rejected because it would pass before required M3 tables exist.
- Check table existence manually instead of migration version: rejected because the migration version is the source of schema state for this repository.
- Require seed data for readiness: rejected because readiness should verify infrastructure/schema, not demo content.
