# Chronix Phase 2 — Engine, Validation, and Execution Packages (CLI deferred to Phase 3)

Owner: dsherwin
Last updated: 2025-09-24 00:46 local
Status: Draft (Actionable)

Executive Summary
- We are moving all CLI feature work to Phase 3. Phase 2 focuses on building the core scheduling and execution capabilities that the UI and REST can leverage: SQL validation, SQL testing, schedule decoding, job workers, and unified logging/alerts/reports.
- Deliverables are backend packages with clear APIs, integration points for the REST/UI, and acceptance tests. Where existing libraries are viable, we will adopt rather than build from scratch.

Out of Scope (deferred to Phase 3)
- CLI command set and UX (beyond the minimal run path already present)
- Advanced admin flows and non-essential endpoints
- Rich notification channels (email/Web Tasks) beyond basic hooks/logging

Primary Deliverables (Phase 2)
1) SQL Validation Package (engine-aware)
2) SQL Testing/Runner Package (transactional sandbox option)
3) Schedule Decoding Package ("is it time to run?" and next-run computation)
4) Job Workers (resilient executors with real-time status)
5) Observability: logs, alerts, reports unification for the above

Secondary/Support Deliverables
- Minimal REST endpoints that call these packages (UI-ready hooks)
- Data model updates for job runs and logs (SQLite first)
- Research notes with decisions and tradeoffs

Detailed Plan

1. SQL Validation Package
Goal
- Provide a server-side validator used by UI actions and backend that:
  - Understands the target SQL engine (sqlite, postgresql, mysql) and dialect nuances
  - Detects obviously invalid SQL before execution (syntax), and optionally policy violations (e.g., disallow DROP for read-only connections)
  - Supports variable placeholders and can validate with example bindings

Scope
- Accept an input structure: { engine: "sqlite"|"postgresql"|"mysql", sqlText: string, variables?: map[string]any }
- Return: { ok: bool, errors: []ValidationError, warnings: []ValidationWarning, normalizedSQL?: string }
- Support variable templating convention used by Actions (e.g., {{varName}}). Provide helper to transform to engine placeholders (e.g., $1 for Postgres, ? for SQLite/MySQL) for dry-run parsing.

Approach & Library Research (implement one of):
- PostgreSQL: pganalyze/pg_query_go (PostgreSQL parser), or use lib/pq-compatible syntax checks via EXPLAIN PARSE-only if DB available.
- MySQL: pingcap/parser (TiDB SQL parser) is mature for MySQL dialect.
- SQLite: sqlite SQL parser options are limited; strategies: use SQLite engine PREPARE on an in-memory connection; or modernc.org/sqlite with PREPARE to catch syntax.
- Cross-dialect: vitess-sqlparser supports MySQL/compatible; not ideal for PostgreSQL/SQLite.

Recommendation
- Favor PREPARE/EXPLAIN on ephemeral in-memory connections where possible:
  - SQLite: open :memory:, BEGIN IMMEDIATE; PREPARE; ROLLBACK.
  - PostgreSQL: connect to configured instance or a dev container (if available) and PREPARE; otherwise fall back to pg_query_go parsing.
  - MySQL: use pingcap/parser for offline parse; or PREPARE if a connection is available.
- Implement pluggable validators per engine. Start with SQLite (mandatory), Postgres (best-effort), MySQL (best-effort).

Acceptance Criteria
- Unit tests feed valid/invalid SQL per engine; variables are recognized and tolerated.
- Policy hooks exist to flag disallowed statements (e.g., DROP/ALTER) when configured.
- Validator returns structured errors/warnings that the UI can present.

2. SQL Testing/Runner Package
Goal
- Execute an Action’s SQL against a target DB with bound variables, returning structured results safely. Support sandbox mode: run inside a transaction and roll back.

Scope
- Input: { engine, dsn|db, sqlText, variables, options: { sandbox: bool, timeout: Duration, maxRows: int } }
- Behavior:
  - Prepare statement(s); bind variables; execute.
  - For SELECT, return rows (capped by maxRows), column metadata, duration.
  - For DML/DDL, return affected rows, duration. If sandbox, wrap in a transaction and rollback at the end.
  - Capture warnings and errors; include a summarized textual log.
- Output: { ok, error?, durationMs, rows?, columns?, affectedRows?, log[] }

Implementation Notes
- Use database/sql with driver-specific behaviors (SQLite: mattn/go-sqlite3 or modernc.org/sqlite; Postgres: pgx; MySQL: go-sql-driver/mysql).
- Variable binding: use the same transformation helper as validation to map {{var}} to placeholders.
- Timeouts via context.WithTimeout.
- Ensure safe scanning for large result sets (limit rows; stream or discard beyond cap).

Acceptance Criteria
- Integration tests cover SELECT and DML across supported engines (SQLite required for Phase 2 using local file/in-memory DB). Postgres/MySQL tests may be tagged/optional.
- Sandbox option leaves DB unchanged; non-sandbox mutates as expected.
- Clear, typed errors for invalid SQL, missing variables, and timeouts.

3. Schedule Decoding Package
Goal
- Given the current (or supplied) time and a schedule definition, determine whether it’s time to execute, and compute the next run time.

Scope
- Support schedule definitions used by UI: single-run instant, recurring via cron expression, and room for structured recurrence.
- Input: { kind: "single"|"recurring", tz?: string, startAt?: time, endAt?: time, cron?: string, structured?: { frequency, every, at?, daysOfWeek?, etc. } }
- Output: { dueNow: bool, nextRunAt?: time, window: { start?: time, end?: time } }

Implementation Notes
- Cron: use robfig/cron/v3 for robust parsing with TZ support (or a light wrapper). Accept standard 5-field cron; consider 6th field seconds in future.
- Structured schedules: compile them down to cron or compute directly.
- Respect time zones consistently; default to system TZ if not provided.
- Provide deterministic tick behavior for tests using a clock abstraction.

Acceptance Criteria
- Unit tests for edge cases: DST transitions, boundaries at start/end windows, disabled schedules.
- Deterministic next-run computation validated by tables of inputs/expected outputs.

4. Job Workers
Goal
- Spawn workers to execute due actions; provide real-time status, avoid hangs, and be resilient to failures and restarts.

Scope
- Worker lifecycle: Pending -> Running -> (Success|Error|Aborted|TimedOut)
- Heartbeats: periodic status emission (e.g., every 1–5s) for UI; include progress text and counters when available.
- Cancellation: context-driven cancellation (shutdown, timeout, user cancel). Ensure DB operations respect context.
- Concurrency: per-job lock to prevent overlaps; global concurrency limit; backpressure on queue.
- Persistence: record job run entries with start/end times, status, error messages, and logs.

Implementation Notes
- Use errgroup for sub-steps, ensure all goroutines are tied to contexts.
- Detect and handle stuck workers via heartbeat watchdog; mark as Aborted on recovery.
- Expose an observer channel or pub-sub for run updates (can back the REST SSE endpoint later).

Acceptance Criteria
- Integration test simulating multiple due jobs: no overlaps for same job; concurrency respected; timeouts enforced; heartbeats recorded/emitted.
- Recovery test: mark in-flight runs as Aborted on startup.

5. Logs, Alerts, Reports
Goal
- Provide consistent observability primitives across validation, testing, scheduling, and execution.

Scope
- Logs: structured slog everywhere; standard keys; job/run IDs attached.
- Alerts: MVP hook interface that can emit notifications (e.g., to log or a stub sink) on failure/success thresholds. Defer email/Web Tasks to Phase 3.
- Reports: persisted job_runs with summarized outputs; recent runs API to drive UI tables. Optional CSV/JSON export helpers.

Implementation Notes
- Define a minimal schema (or extend existing) for job_runs, job_run_logs.
- Provide helper to attach a logger and metrics recorder to each worker run.

Acceptance Criteria
- Recent runs endpoint returns last N runs with status, duration, and short message.
- Logs include job/run IDs and error key is consistently "error".

Interfaces and Contracts (Draft)
- Validator API: Validate(ctx, req) -> resp
- Tester API: Test(ctx, req) -> resp
- Scheduler API: NextRun(now, schedDef) -> time; Due(now, schedDef) -> bool
- Worker API: Execute(ctx, job, runOpts) -> runResult; emits heartbeats via channel

Minimal REST Touchpoints for UI (Phase 2)
- POST /validate-sql -> validator response
- POST /test-sql -> tester response (sandbox by default)
- POST /schedule/decode -> schedule decoder response (optional)
- GET /runs/recent?limit=50 -> recent run summaries
- SSE /events (optional) -> run updates

Data Model Notes (SQLite first)
- scheduled_jobs(id, name, enabled, schedule_kind, cron, tz, start_at, end_at, sql_text, created_at, updated_at)
- job_runs(id, job_id, started_at, finished_at, status, duration_ms, error_text, summary, created_at)
- job_run_logs(id, run_id, ts, level, message, fields_json)
- job_variables(job_id, name, value, created_at, updated_at) — optional for Phase 2 if variables are provided at run time

Risks & Mitigations
- Parser availability per engine: prefer PREPARE against the engine when possible; fall back to best-available parsers.
- Time-zone/DST edge cases: centralize TZ handling and add table-driven tests around DST boundaries.
- Long-running queries: enforce context timeouts and maxRows caps; surface partial results safely.
- Concurrency leaks: standardize on context + errgroup; watchdog and recovery paths.

Recommendations (Potentially Missed in Request)
- Secrets management for DB connections (avoid plain-text in logs/configs; redaction helper).
- Error taxonomy for API/UI (stable error codes for validator/tester/worker).
- Feature flags to gate engines and risky operations.
- Idempotency for job submissions to avoid duplicate runs on restarts.
- Metrics (Prometheus) for job counts, durations, failures by job/engine.
- Access control posture (even if single-user): protect destructive endpoints until initialized.

Milestones & Checklist
M1 — Packages Skeletons and Research
- [ ] Draft validator/tester/schedule/worker package interfaces and scaffolds
- [ ] Research notes and POC for per-engine validation strategies

M2 — SQLite First (Reference Engine)
- [ ] Implement validator via PREPARE on :memory:
- [ ] Implement tester with sandbox transactions and limits
- [ ] Implement schedule decoder with cron lib and TZ
- [ ] Implement worker with heartbeats and timeouts
- [ ] Persist job_runs and recent runs API

M3 — Postgres/MySQL Adapters (Best-effort)
- [ ] Pluggable validators (pg_query_go, pingcap/parser or PREPARE)
- [ ] Runner adapters and connection management

M4 — Observability & Hardening
- [ ] Structured logs with job/run IDs; error key standardization
- [ ] Watchdog/recovery on startup marks stale runs Aborted
- [ ] Add tests with fake clock; add race-enabled integration tests

M5 — Documentation
- [ ] Update README status and quickstart when Phase 2 completes
- [ ] Document APIs and examples for UI consumers

Acceptance (Phase 2 overall)
- Server exposes endpoints to validate and test SQL; scheduler + workers can run a simple SQLite job on schedule, record the run, and provide recent run summaries. CLI work remains deferred to Phase 3.

References
- Project Guidelines (root .junie/guidelines.md)
- dev/project/research_notes.md (record engine validation findings)
- robfig/cron/v3, pganalyze/pg_query_go, pingcap/parser, pgx, go-sql-driver/mysql, mattn/go-sqlite3 or modernc.org/sqlite
