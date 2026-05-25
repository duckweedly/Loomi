# Research: M12 Real MCP Smoke Closeout

## Decision: Use a Go test subprocess as the local stdio MCP fixture

**Rationale**: Existing M11/M12 tests already use test subprocesses for local stdio fixtures. Reusing that pattern gives a real process boundary, real stdin/stdout, and real `Content-Length` frames without adding fixture binaries or dependencies.

**Alternatives considered**:

- External fixture binary: rejected because it adds install/build surface.
- Fake executor: rejected because M12.5 must prove `StdioMCPToolExecutor`.
- Remote MCP mock: rejected because remote MCP is out of scope.

## Decision: Drive approval through HTTP and execution through the worker

**Rationale**: The closeout needs evidence across API approval and worker execution, not only service methods. HTTP approve plus `Worker.ProcessOne` proves the scoped endpoint and existing M6/M7 resume path without starting a long-running server.

**Alternatives considered**:

- Browser-only smoke: useful but environment-dependent.
- Service-only approve: rejected as weaker evidence because HTTP approval would remain unexercised.

## Decision: Record browser smoke limitation if local app cannot run

**Rationale**: Browser smoke depends on a runnable local API, database, frontend, and deterministic provider fixture. Backend/httpapi/runtime smoke covers the same state transitions when browser smoke is not practical in the current work session.

**Alternatives considered**:

- Force browser validation: rejected because it could require unrelated environment setup.
- Skip UI evidence silently: rejected because docs must state limitations.

## Decision: Keep M12.5 as documentation/evidence closeout

**Rationale**: The user explicitly excluded remote MCP, marketplace, plugin install, sandbox, automation, and multi-tool loop work. The correct deliverable is proof of the existing local path.

**Alternatives considered**:

- Add MCP admin UI: rejected as new platform capability.
- Add multi-call loop support: rejected as explicitly out of scope.
