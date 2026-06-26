# Internal Architecture Redesign

## Status

This redesign is complete.

This document began as the architecture plan for breaking up the top-level `internal` package and now also serves as the implementation record.

Implemented outcome summary:

- top-level handwritten `internal` was drained as a working package
- execution, progress, job queue/history, notifications, events, user logic, settings, and server runtime now live in explicit subpackages
- the final package names landed slightly differently than the earliest draft:
  - `appstate` was split into `cxsettings` and `serverruntime`
  - `authz` landed as `cxuser`
  - `notify`, `progress`, `jobrun`, `execution`, `activity`, and `events` all became real package boundaries
- transitional bridge shims were removed once callers were moved onto the new packages directly

## Purpose

Chronix has outgrown the "single large `internal` package" stage.
The current structure was a good fit while the product shape was still forming, but the top-level
`internal` package now mixes application state, execution engines, progress persistence, notifications,
HTTP-adjacent helpers, and server lifecycle code in one namespace.

This redesign is intended to:

- reduce the amount of handwritten code that lives directly under `internal/`
- establish clearer domain and service boundaries
- enforce one-way dependency flow
- make package extraction possible without circular dependency problems
- preserve existing behavior while restructuring underneath it

## Hard Constraints

- Generated files under `internal/db` must not be hand-edited.
- `internal/db/*.gen.go`, `internal/db/gen.go`, and `internal/db/db_sqlite.go` are generated artifacts.
- Schema truth lives in `dev/schema.sql`.
- `internal/db/assets/schema.db` is derived from `dev/schema.sql`.
- `gormdb2struct` plus `.gormdb2struct.toml` regenerates the generated DB layer.
- Handwritten companion files under `internal/db/...` are allowed and safe.
- The redesign target is the handwritten runtime/orchestration structure around `internal/db`, not the generated DB layer itself.

## Current Problem Areas

### 1. Top-Level `internal` Is Acting Like A Catch-All Package

Current top-level responsibilities include:

- application settings and user management
- server bootstrap and status
- SSE and SPA middleware
- notifications and alert dispatch
- run progress and run-memory tracking
- job queue orchestration
- SQL, shell, and web task execution
- cleanup jobs and helper utilities

This makes dependency flow difficult to reason about and encourages code to reach "sideways" through the parent package.

### 2. Execution Logic Is Not Isolated From Product Side Effects

The executor area currently combines:

- loading DB objects
- decrypting connection secrets
- choosing local vs agent runner
- evaluating expectations
- capturing variables
- precreating/persisting run steps
- publishing progress events

This is the main reason an immediate move into a separate executor package would create cycles.

### 3. App State And Runtime Services Are Blended

Settings, server status, notifications, progress, and auth/user logic all live in the same top-level namespace,
which makes it harder to see which code is domain logic versus runtime coordination.

## Dependency Rules For The Redesign

These rules should drive every move:

1. Leaf packages do work; upper packages orchestrate.
2. Lower packages must never import upward into orchestration packages.
3. `internal/db` is an infrastructure dependency, not a business-logic home.
4. Execution engines must not directly call SSE, notification, or queue code.
5. Progress/event publication should happen through explicit sinks or service boundaries.
6. `cmd` remains executable wiring.
7. `cxrestapi` remains the HTTP/API boundary.
8. `pkg` remains reusable support code with no upward dependency on Chronix runtime packages.

## Target Package Layout

This section records the original target map that guided the redesign. The implemented result stayed close to this, with the main naming differences called out in the status section above.

### `internal/appstate`

Purpose:

- global server/application state
- current settings cache
- server status
- build metadata
- runtime configuration sync into supporting services

Likely contents:

- `buildinfo.go`
- `consts.go`
- `cxsettings.go`
- pieces of `server.go` related to status only

Notes:

- This package should expose application state, not server transport behavior.

### `internal/authz`

Purpose:

- Chronix user model helpers
- login/auth-related user operations
- email availability / password handling

Likely contents:

- `cxuser.go`

Notes:

- Keep HTTP token/cookie mechanics in `cxrestapi`, not here.

### `internal/activity`

Purpose:

- user activity recording and retrieval

Likely contents:

- `user_activity.go`

### `internal/notify`

Purpose:

- notification creation
- assignment to users
- unseen counts
- alert dispatch orchestration
- alert formatting/rendering

Likely contents:

- `notifications.go`
- `notifyutil.go`
- `alerts.go`
- `alerts_output.go`
- `alerts_html.go`
- `alerts_config.go`

Notes:

- This should own "create notification and dispatch alerts" behavior.
- SSE fanout for notifications can remain here if it is specific to notification delivery.

### `internal/progress`

Purpose:

- run progress buffer
- run progress snapshots
- run lifecycle event publication
- DB persistence hooks for run/step status changes

Likely contents:

- `progress.go`
- `run_logger.go`
- parts of `job_runs_mem.go` that belong to run-state tracking

Notes:

- This package is one of the main keys to breaking executor cycles.
- The executor should talk to this package through an interface or sink, not through the parent package.

### `internal/jobrun`

Purpose:

- queued/running/finished in-memory run history
- run aggregation helpers
- queue-facing run metadata helpers

Likely contents:

- `job_runs_mem.go`
- `aggregateRunOutput()` currently in `jobqueue.go`

Notes:

- This is optional as a separate package.
- It may merge cleanly into `internal/progress` if that feels simpler during implementation.

### `internal/execution`

Purpose:

- orchestrated job execution
- SQL, shell, and web task execution engines
- expectation evaluation
- variable substitution and capture

Likely contents:

- `executor.go`
- `executor_expectations.go`
- `executor_capture.go`
- `executor_shell.go`
- `webtask_executor.go`
- `webtask_executor_expectations.go`
- `webtask_executor_capture.go`

Possible substructure:

- `internal/execution/sql`
- `internal/execution/shell`
- `internal/execution/webtask`

Notes:

- The first real boundary here is between execution and progress/persistence side effects.
- This package should not directly depend on job queue code or generic app-state globals.

### `internal/serverruntime`

Purpose:

- embedded SPA assets
- anonymous server routes for embedded frontend
- server status endpoint integration
- TLS helper coordination tied to server runtime

Likely contents:

- `server.go`
- `spa_middleware.go`
- `tls_helpers.go`
- `ip.go`

Notes:

- This package is about server runtime behavior, not HTTP listener setup.
- HTTP listener bootstrap remains under `cxrestapi`.

### `internal/events`

Purpose:

- SSE session management and event fanout

Likely contents:

- `sse.go`

Notes:

- This can also remain small and independent if it only depends on low-level user/session state.
- It should not know about executors or queue logic directly.

### Keep Existing Subpackages

These already look like proper subpackages and should generally remain so:

- `internal/agentmux`
- `internal/connhealth`
- `internal/db`
- `internal/notifier`
- `internal/scheduler`
- `internal/secret`
- `internal/shellrun`
- `internal/sqlrunner`
- `internal/sshutil`
- `internal/svc`
- `internal/updater`
- `internal/webtaskrun`

## Executor-Specific Redesign Strategy

The executor is the proving ground for the redesign.

### Why It Cannot Be Moved As-Is

`ExecuteJob()` currently depends on:

- DB loading and persistence
- connection decryption
- SQL/shell/webtask runners
- progress publication through `ProgressOn*`

If it moves into `internal/execution` while still calling `ProgressOn*` in the parent package,
the dependency graph becomes:

- job queue imports execution
- execution imports parent `internal`

That creates the cycle.

### Correct Solution

Split execution from run-progress side effects.

The executor package should accept a small event sink or reporter interface, for example:

- run queued/started
- step started
- step finished
- run finished

The concrete implementation can live in the progress package and bridge to:

- DB persistence
- SSE fanout
- in-memory snapshots

This lets execution become a leaf package instead of a parent-aware package.

## Proposed Migration Order

This order is designed to reduce risk and preserve behavior.

### Phase 1. Establish The New Service Boundaries

- create target packages without moving everything yet
- move small self-contained files first
- add thin compatibility wrappers only where necessary

Best early moves:

- `activity`
- `authz`
- `notify`
- `events`

These are comparatively low-risk and clarify the package map.

### Phase 2. Extract Progress As Its Own Boundary

- move progress/run-status logic into a dedicated package
- move run logging and run state persistence behind explicit functions
- make queue and executor depend on that package instead of on top-level helpers

This phase is the main dependency-unlocking step.

### Phase 3. Extract Execution

- introduce an execution service/sink interface
- move SQL/shell/webtask execution into `internal/execution`
- keep orchestration behavior stable through adapter code

This is the highest-value architectural move.

### Phase 4. Extract App State And Server Runtime

- move settings/server status/build info into `internal/appstate`
- move embedded SPA/server runtime helpers into `internal/serverruntime`

This reduces the amount of non-execution code living at the top level.

### Phase 5. Reduce Or Eliminate Top-Level `internal` As A Working Package

End-state goal:

- top-level `internal` should contain little or no substantive handwritten runtime logic
- it may remain as a shallow facade temporarily if needed during migration
- but the real business/runtime ownership should live in subpackages

## Safety Rules During Refactor

- never hand-edit generated DB files
- move code in small phases with builds/tests after each phase
- avoid broad behavior changes while moving packages
- introduce interfaces only where they break dependency direction cleanly
- do not create "abstraction theater"; every new interface must remove a real coupling
- prefer explicit structs and direct calls over framework-style indirection

## Validation Requirements

Each migration phase should keep these green:

- `go test ./cxrestapi/... ./internal/...`
- `go build ./dev/cxrelease`
- `./dev/ci-local.sh` at stable checkpoints

## Recommended First Execution Phase

The first active implementation phase should be:

1. create `internal/progress` as a real package boundary
2. move progress/run-state helpers there
3. adapt queue and executor to use it
4. only then move executor into `internal/execution`

This gives the best chance of avoiding circular imports while preserving behavior.

## Non-Goals

These are not goals of the redesign unless required by dependency cleanup:

- new product features
- release workflow changes
- DB schema redesign
- changing generated DB output format
- changing external API behavior unless necessary to preserve internal structure

## Completion Criteria

This redesign is considered complete when:

- the top-level handwritten `internal` package is no longer the primary home of Chronix runtime logic
- execution, progress, notifications, app state, and runtime/server concerns are separately packaged
- dependency flow is one-way and understandable
- package extraction no longer depends on reaching back into a parent catch-all package
- all tests and local CI pass
