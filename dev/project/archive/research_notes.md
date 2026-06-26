# Chronix Research & Brainstorming Notes

This file is used to capture ideas, comparisons, and reasoning that inform Chronix's development. Once decisions are finalized, they will be summarized in the milestone file.

---

## 1. Scheduling Formats

Chronix will support three types of scheduling interfaces:

1. **Single-shot** – A specific timestamp for one-time execution.
   - UI: Text field with a calendar and time picker.
   - Stored as a single `run_at` timestamp.

2. **Recurring** – A structured UI similar to Apple Calendar, but with minute-level granularity.
   - Frequency types: minute, hour, day, week, month, year.
   - Modifiers: day-of-week toggles, nth weekday patterns, intervals (e.g. every 5 minutes).
   - Constraints: start and end time range.

3. **Advanced (Cron)** – Raw 5-field cron entry for power users.
   - UI: 5 text fields with validation.
   - Stored directly as cron string.

Note: Internal granularity is limited to **minutes**. Seconds-level execution is not supported in v0.1.

Libraries/tools to consider:
- [robfig/cron](https://github.com/robfig/cron)
- [go-co-op/gocron](https://github.com/go-co-op/gocron)

---

## 2. Schedule Constraints

- All recurring schedules will support an optional:
  - **Start timestamp**
  - **End timestamp**
- Single-shot jobs will ignore all recurrence constraints.
- These constraints can be stored in a structured rule object and evaluated at run-time.

---


## 3. Database Connections

Chronix will manage database targets as reusable configuration objects, separate from scheduled jobs.

### Named Database Connections

- Users define connections independently in a “Connections” section.
- Each connection includes:
  - DB type (e.g., Postgres, MySQL)
  - Host, port
  - Username
  - Password (optionally encrypted)
  - DB name
  - Optional fields: SSL mode, schema, extra DSN params
- Connections are assigned a user-friendly name (e.g., `prod-reports`, `dev-db`)

### Usage in Jobs

- The primary object in Chronix is a **Job**
- Each Job includes:
  - A schedule (single-shot, recurring, or cron)
  - One or more Action definitions (SQL statements)
  - A reference to one or more pre-configured database connections
- At runtime, the Action(s) will be executed against each selected DB

### UI Behavior

- Users configure connections from a dedicated section of the UI
- When creating a Job, users select one or more databases to target from a list of existing connections

### Security Notes

- Passwords will be encrypted at rest
- UI should mask stored credentials
- Future enhancement: vault-backed or runtime-supplied secrets

---

## 4. Action Definition UI

Chronix will initially support only raw SQL input for Action definitions.

- Users will provide a SQL statement as plain text in the UI.
- Multi-statement Actions will be supported as long as the target DB allows it.

### Validation Strategy

- The UI will offer a “Verify” button to check the SQL's validity before saving.
- Verification will be performed by:
  - Wrapping the SQL in a transaction
  - Attempting execution against the selected DB(s)
  - Capturing any SQL syntax or runtime errors
  - Rolling back the transaction to avoid changes
- If no errors are thrown, the SQL is considered valid.

---

## 5. Post-Execution Handling

After an Action executes, Chronix will capture:

- Success/failure status
- Any error messages
- Execution duration
- Timestamp of run
- Rows affected (optional per DB support)

### Notifications

Each Job will include a **Notification** section with two optional subsections:

1. **Notify on Completion**
   - List of email addresses and/or SMS numbers to notify when an Action completes (success or failure).

2. **Notify on Error**
   - List of email addresses and/or SMS numbers to notify only when the Action fails.

Notes:
- Email is not tied to user accounts (we do not collect user emails in v1).
- Allows third-party recipients (e.g. team leads, on-call SMS).
- SMS/email backend TBD—can use pluggable providers (e.g., SendGrid, Twilio).
- Notifications can be configured per Job.

#### Global Notification Defaults

- Chronix will support a global default list of emails and/or SMS numbers for Action notifications.
- This setting will live in the application settings section of the UI.
- Fields:
  - Default email(s) on completion
  - Default email(s) on error
  - Default SMS on completion
  - Default SMS on error
- When creating a new Job, these defaults will pre-populate the notification section but can be edited or removed.
- Stored in Chronix's existing settings system (no relational linkage to Jobs).

### SMS Delivery Strategy

- SMS delivery will be supported using one or more third-party providers.
- Chronix v1 will initially support **Twilio**, with an interface designed for future pluggable provider support.
- Other providers considered for future support: Nexmo (Vonage), Plivo, Telnyx, Bandwidth.

#### Provider Configuration

- A global settings section will allow the admin to configure provider credentials.
- Example (Twilio):
  - `account_sid`
  - `auth_token`
  - `from_number`

- SMS features in the UI will only be enabled if a provider is configured.
- SMS-specific fields will then appear in the Job notification section.

#### Job-Level SMS Behavior

- Admin selects which provider to use per Job (or default to global setting).
- One or more recipient phone numbers can be specified.
- Numbers will be validated for format before saving.

- Notifications can be set for:
  - Completion
  - Error

---

## 6. Viewing History

Since the web UI is now the primary interface, viewing history will be handled in the Job detail view.

- Each Job will include a log section displaying past executions.
- This section will show success/failure, run time, and possibly error messages.
- Log entry format and detail level will be determined during implementation.
- Retention policy will be decided later as part of system tuning.

No CLI log viewing is planned for v1.

---

## 7. Concurrency & Duplication Protection

Chronix will allow the user to define concurrency behavior per Job.

### Concurrency Setting (per Job)

- **Allow concurrent runs** (default): 
  - Multiple executions of the same Job may run simultaneously.
  - No locking or coordination is performed.

- **Disallow concurrent runs**:
  - If an instance of the Job is already running, the scheduler will **skip** the current scheduled execution.
  - No queuing or retry will be done in v1.

### UI

- This will appear as a toggle or dropdown in the Job creation/edit form.
- Labeled as “Allow concurrent execution” (checkbox or select)

---

## 8. User Accounts & Authentication

### Initial Setup

- Chronix will require an admin user to be created on first run.
- This is a single-user system for v1 (multi-user may come later).
- Admin user credentials will be required to access the web UI.
- CLI access implies full system access.

### User Schema

- Fields:
  - `id` UUID
  - `username` (unique)
  - `password_hash`
  - `created_at`
  - `last_login_at`
  - `is_admin` (defaults to true)

### Password Reset

- Password reset will be CLI-based in v1.

Steps:
1. Run CLI command:
   ```bash
   chronix user reset-password --username admin
   ```
2. CLI prints a reset token.
3. User visits `/reset?token=...` in browser.
4. Web UI prompts for new password.
5. Token is invalidated after use or after timeout.

- This is acceptable because CLI access implies full trust (same as direct DB access).

---

## Future Enhancements

- Introspection of schema to allow query building or smart validation
- Parameterized SQL
- Action templates/snippets
- Allow Web Task targets or Slack integration
- Support for multiple SMS providers simultaneously
- Usage limits and delivery logs
- Rate limiting or retry logic
- Manual run or re-run capability from the UI
- Queued or delayed execution after prior run completes
- Max concurrency limits
- Execution timeout enforcement
- Multi-user support
- User roles / permissions
- Email-based password reset
- 2FA

---

## Brainstorming Log

This section captures live brainstorming sessions. Entries are timestamped and organized for easy conversion into finalized specs.

### [2025-08-27 00:11 Local] Session title goes here

#### Decisions
- (None yet)

#### Open Questions
- (None yet)

#### Ideas
- (None yet)

#### Risks / Assumptions
- (None yet)

#### Action Items
- [ ] Owner: —

### [2025-08-27 00:12 Local] Actions: Capabilities & Scope v0.1

#### Decisions
- v1 scope: Each Job executes exactly one Action against exactly one database connection.
- Initial Action type is Raw SQL; multi-statement allowed where the target DB permits it.
- Validation approach: wrap SQL in a transaction, attempt execution, capture errors, and rollback during validation.

#### Open Questions
- Parameterization: Include named parameters in v1, or defer to v1.x?
- Timeout: Per-Action timeout separate from job/global? Enforced via context deadlines?
- Safety controls: Require an explicit "allow write" toggle? Offer read-only/dry-run or EXPLAIN modes?
- Result handling: Store row counts and optional sample of SELECT results, or only metadata?
- Error classification: Distinguish syntax/runtime/connectivity/timeout for UX and notifications?
- Future (explicitly out of scope for v1): multi-DB targets and multiple Actions per Job.

#### Ideas
- Action schema v1: { id, name, type=sql, sql_text, requires_write: bool, timeout_ms?, notes }.
- Parameters (if included): Named placeholders with preview of rendered SQL before execution.
- Dry-run/EXPLAIN: Provide EXPLAIN (db-dependent) to increase confidence before saving/running.
- Pre/Post checks: Optional pre-check (row count threshold) and post-check (affected rows min/max).
- Runtime vars later: e.g., {{now}}, {{job_id}}, {{run_id}} when templating arrives.
- Templates/snippets: Defer to next version; keep the design pluggable.

#### Risks / Assumptions
- SQL dialect differences complicate validation and EXPLAIN support.
- Long-running queries require timeouts and cancellation handling.
- Partial-success risk reduced in v1 due to single DB per Job, but failures still need clear reporting.
- Security: Protect credentials and avoid leaking sensitive data in logs/notifications.

#### Action Items
- [ ] Draft Action schema fields and defaults for v1 (Owner: DS, due: 2025-08-30)
- [ ] Spike: context-based timeouts on DB exec per driver (Owner: DS, due: 2025-08-30)
- [ ] Propose error classification taxonomy and mapping to notifications (Owner: DS)
- [ ] Decide v1 stance for parameters and placeholder syntax if included (Owner: DS)
- [ ] Define minimal result recording: status, error, duration, rows_affected (Owner: DS)


### [2025-08-27 00:13 Local] Actions authoring: parameters and pre-binding validation

#### Decisions
- (None yet)

#### Open Questions
- Placeholder syntax: prefer one for v1? Options: `{{name}}`, `:name`, `@name`, `$1`-style (positional). Named placeholders improve readability and mapping.
- Dialect during Action creation: require selecting a dialect (Postgres/MySQL) for better parsing, or allow "Generic SQL" with limited checks?
- Validation scope pre-DB: basic lexical/structural parsing only, or lightweight linting (e.g., forbid obvious mistakes like trailing commas, unmatched parentheses)?
- Parameter typing: Should parameters have declared types and optional defaults at Job creation?

#### Ideas
- Use named placeholders `{{param}}` in the editor for v1; store a Parameters panel with: name, type (string/int/float/bool/timestamp/json), required, default, description.
- Render parameters at execution time to driver-specific placeholders (e.g., Postgres `$1..$n`, MySQL `?`) to keep authoring consistent.
- Editor validation stages:
  - Stage 1 (Generic): tokenize SQL, check balanced quotes/parentheses, validate parameter names, detect obviously invalid tokens; ignore contents of `{{param}}`.
  - Stage 2 (Dialect-aware optional): if a dialect is selected, run a dialect parser (where available) with parameters substituted by safe literals for parsing only.
  - Stage 3 (Post-bind): once the Job selects a connection, enable “Validate against DB” (exec in txn and rollback) and optional EXPLAIN.
- Possible Go libs:
  - MySQL flavor: Vitess SQL parser (stable, well-known) for dialect-aware checks.
  - Postgres: consider pg_query bindings; if too heavy, fall back to Stage 1 Generic checks for v1.
  - Generic tokenizer: implement simple lexer to support Stage 1 quickly.
- UX:
  - SQL editor with inline errors and a “Check syntax” button.
  - “Parameters” side panel auto-detects `{{name}}` occurrences and prompts to define them.

#### Risks / Assumptions
- Parser-dialect mismatch may cause false failures or passes.
- Native parser dependencies (e.g., pg_query) can increase build complexity.
- Overly strict checks could block valid cross-dialect SQL; keep Generic mode permissive.

#### Action Items
- [ ] Decide placeholder syntax for v1 (recommend `{{name}}`). (Owner: DS)
- [ ] Choose dialect support for pre-binding checks: Generic-only vs. Generic+MySQL (Vitess) vs. add Postgres later. (Owner: DS)
- [ ] Spike Stage 1 Generic tokenizer with balanced-delimiter checks and param validation. (Owner: DS)
- [ ] Define parameter schema (name, type, required, default, description) and UI representation. (Owner: DS)
- [ ] Plan Stage 3 DB validation flow once a connection is selected (transaction + rollback). (Owner: DS)


### [2025-08-27 00:14 Local] Actions authoring: decisions on placeholders, dialect, types, timeout

#### Decisions
- Placeholder syntax: `{{name}}` for v1 authoring.
- Dialect at Action creation: allow selecting a specific dialect (Postgres/MySQL) for thorough parsing; choose "Generic SQL" for cross-dialect Actions to keep checks permissive.
- Parameter types: do not collect/require types in v1 (support custom types; treat as opaque values entered at Job creation).
- Default Action timeout: 60s.

#### Open Questions
- Should v1 allow a per-Action timeout override in the UI (falling back to default 60s if unset)?

#### Ideas
- Editor UX: SQL editor + Dialect selector (Postgres | MySQL | Generic). Switching to Generic relaxes checks to Stage 1 only.
- Parameters panel (no types): auto-detect `{{name}}` and prompt for name + optional description; values supplied at Job creation.
- Validation button behavior:
  - Generic: lexical checks only (balanced quotes/parentheses, placeholder names, simple token sanity).
  - Dialect-selected: attempt dialect parser with placeholders substituted by safe literals for parsing only.
  - Post-bind (once a connection is chosen): DB-backed validation in a transaction with rollback; optionally EXPLAIN.

#### Risks / Assumptions
- Without declared types, some runtime bindings may fail due to driver coercion; mitigate via preview rendering and example values during Job creation.
- Dialect parser availability may differ by DB; keep Generic path robust.

#### Action Items
- [ ] Decide on per-Action timeout override availability in v1 UI (Owner: DS)
- [ ] Implement Stage 1 generic tokenizer/validator (Owner: DS)
- [ ] Add Dialect selector and wire validation modes (Owner: DS)
- [ ] Set default timeout=60s in backend; honor global default if later introduced (Owner: DS)


### [2025-08-27 00:15 Local] Actions: timeout override decision

#### Decisions
- Per-Action timeout override is allowed in v1; default remains 60s when unset.

#### Open Questions
- (None for this topic)

#### Ideas
- UI: Optional numeric field “Timeout (seconds)”; empty = use default (60s). Validate range (e.g., 1–3600).
- Backend: Store nullable integer seconds; at run-time, resolve to provided value or fallback default (60s).

#### Risks / Assumptions
- Users might set too small/large values; clamp to a sane range and document defaults.

#### Action Items
- [ ] Backend: Add optional timeout_seconds to Action model; default to 60s if null (Owner: DS)
- [ ] UI: Add optional Timeout (seconds) field with helper text and validation (Owner: DS)


### [2025-08-27 00:16 Local] Multi-step Actions (pipeline) — v1.x brainstorm

#### Decisions
- Keep v1 simple: one Action per Job, single DB connection. Multi-step Actions are targeted for v1.x.

#### Open Questions
- Step boundaries: Are steps always SQL-only in v1.x, or do we allow non-SQL steps (HTTP call, script) later?
- Transaction model: per-step transactions only, or optional job-level transaction when all steps hit the same DB?
- Expected result evaluation timing: immediate after step vs. aggregate checks after multiple steps?
- Failure policy defaults: if a step fails expectation, should default be “stop pipeline” or “continue”?
- Observability: which step details should be stored (input params snapshot, rows affected, sample rows)?

#### Ideas
- Step model (sequential execution):
  - id, name, order
  - type: sql (v1.x), future types: http, noop, wait
  - sql_text (uses {{param}} placeholders), dialect (or Generic)
  - timeout_seconds (required per step)
  - expected: list of matcher objects
  - on_failure: stop | continue
  - notes
- Expected result matchers (first-pass set):
  - no_error: step considered pass if execution succeeds
  - completes: explicitly ignore result and only require completion (same as no_error)
  - rows_affected: { op: >= | == | <=, value: N }
  - scalar_equal: { query: SELECT expr, equals: value }
  - row_exists: { query: SELECT 1 FROM … WHERE … LIMIT 1 }
  - regex_on_text: { source: error|stdout|result_text, pattern: "…" } (db-dependent)
  - jsonpath_match (for JSON-returning queries where supported)
- Control flow:
  - Default: stop on matcher failure; allow per-step override to continue
  - Optional “continue_on_error” separate from expectation failure (e.g., retry later feature)
- Timeouts:
  - Required per step; clamp, e.g., 1–3600s; default inherit from Action unless overridden
- UI concepts:
  - Step list with drag-reorder; each step has editor, expectations panel, timeout field, failure policy dropdown
  - Execution preview: shows resolved SQL with example parameter values before run
- Execution semantics:
  - Strictly sequential within a Job; no parallelism in v1.x pipeline
  - Each step executes in its own transaction unless a job-level transaction is explicitly supported and DB matches

#### Risks / Assumptions
- Cross-DB differences complicate expectation matchers (e.g., rows_affected semantics, JSON operators)
- Overly rich matcher language can delay delivery; start with no_error, rows_affected ops, and row_exists
- Job-level transactions across multiple steps may hold long locks; default to per-step transactions

#### Action Items
- [ ] Define minimal matcher set for v1.x (proposed: no_error, rows_affected, row_exists) (Owner: DS)
- [ ] Draft Step schema (fields above) and persistence model (Owner: DS)
- [ ] Prototype UI for step editor and expectations panel (Owner: DS)
- [ ] Decide on default failure policy and allowed overrides (Owner: DS)
- [ ] Evaluate feasibility of job-level transaction when all steps target same DB (Owner: DS)


### [2025-08-27 00:17 Local] Multi-step Actions — decisions (matchers, policy, scope, transactions)

#### Decisions
- Minimal matcher set for v1.x: no_error, rows_affected (>=/==/<=), row_exists.
- Default failure policy: stop pipeline when a step fails its expectation (overridable per step to continue).
- Step types scope for v1.x: SQL-only; non-SQL (HTTP/script/etc.) deferred.
- Transaction model: selectable — per-step transactions OR one per-action (job-level) transaction when applicable.

#### Open Questions
- Per-action (job-level) transaction applicability rules: only when all steps target the same DB connection and the DB supports required semantics. Finalize guardrails.

#### Action Items
- [ ] Specify UI control for transaction mode (Per step | Entire action) with eligibility hints (Owner: DS)
- [ ] Define backend execution guards: enforce same-connection requirement for per-action transactions (Owner: DS)
- [ ] Document behavior for partial failures under each transaction mode (Owner: DS)


### [2025-08-27 00:18 Local] v1 Actions: editor + backend contract (forward-compatible)

#### Scope (v1)
- One Action per Job, single database connection per Job.
- Action type: Raw SQL only. Multi-statement allowed if DB permits.
- Placeholders: `{{name}}` (named). Values provided at Job creation; no parameter types collected in v1.
- Dialect at creation: Postgres | MySQL | Generic (Generic = permissive checks only).
- Timeout: default 60s; optional per-Action override in seconds.

#### Editor (UI) contract
- Fields
  - SQL text (multiline editor, shows line numbers)
  - Dialect selector: Postgres | MySQL | Generic
  - Timeout (seconds, optional; empty = 60s default)
  - Notes (optional)
- Parameters panel
  - Auto-detect `{{name}}` occurrences in SQL and list unique names
  - No type selection in v1. Optional description per parameter
  - Values are bound at Job creation (not in Action editor)
- Validation buttons
  - Check syntax (pre-DB):
    - Generic: lexical checks only (balanced quotes/parentheses; valid placeholder names `[A-Za-z_][A-Za-z0-9_]*`; disallow unterminated string literals)
    - Dialect-selected: attempt dialect parse with placeholders substituted by safe literals (e.g., `NULL` or `'x'`) for parsing only
  - Validate against DB (post-bind, in Job editor once a connection is set): execute in a transaction and rollback; capture errors; optional EXPLAIN
- Messaging
  - Inline squiggles + panel with errors/warnings (code, message, line/col)

#### Backend model (v1)
- Action
  - id: UUID
  - name: string
  - dialect: enum { postgres, mysql, generic }
  - sql_text: string
  - timeout_seconds: int? (nullable)
  - notes: string? (nullable)
  - created_at, updated_at
- Job
  - id: UUID
  - action_id: UUID (FK)
  - connection_id: UUID (FK to named DB connection)
  - schedule: object (existing)
  - params: map<string,string> (values for `{{name}}`)
  - created_at, updated_at

#### Execution semantics (v1)
- Build context with timeout = action.timeout_seconds || 60
- Compile SQL:
  - Replace `{{name}}` with driver placeholders (Postgres `$1..$n`, MySQL `?`)
  - Assemble args in a deterministic order (sorted by placeholder discovery order or alpha; recommend discovery order by first appearance)
- Run in a transaction for validation mode; for actual runs, transaction behavior follows Action/DB defaults (no cross-DB semantics in v1)
- Capture execution record: status, error (if any), duration_ms, rows_affected (if available), started_at, finished_at
- Notifications triggered per Job settings

#### Validation APIs (suggested)
- POST /actions/validate
  - body: { sql_text, dialect }
  - mode: generic | dialect
  - response: { ok: bool, errors: [{ code, message, line?, column? }], warnings: [...] }
- POST /jobs/validate-action
  - body: { job_id }
  - executes in txn against job.connection_id with job.params applied; rolls back

#### Error shape (standard)
- { code: string, message: string, details?: object }
- Codes examples: SYNTAX_ERROR, UNBALANCED_DELIMITER, DIALECT_PARSE_ERROR, TIMEOUT, CONNECTION_ERROR, RUNTIME_ERROR

#### Security/logging
- Do not log full SQL with parameter values by default; include masked previews and parameter names only
- Audit trail stores minimal necessary data

#### Forward-compatibility with multi-step (v1.x)
- Internally, an Action can be represented as a pipeline with 1 step in v1; later, add steps[] with: { name, order, sql_text, dialect, timeout_seconds, expected[], on_failure }
- Transaction mode field reserved for v1.x (per-step vs per-action) but hidden/disabled in v1
- Expectation matchers (v1.x): start with no_error, rows_affected ops, row_exists

#### Action Items
- [ ] Implement Generic lexical validator (balanced delimiters, placeholder name rules) (Owner: DS)
- [ ] Wire dialect parse for MySQL via Vitess; Postgres TBD or fallback to Generic (Owner: DS)
- [ ] Add Action fields (dialect enum, timeout_seconds, notes) and migrations (Owner: DS)
- [ ] Implement placeholder compiler to driver placeholders + args ordering (Owner: DS)
- [ ] Add /actions/validate and /jobs/validate-action endpoints (Owner: DS)
- [ ] UI: Editor, Dialect selector, Timeout field, Parameters panel with auto-detect (Owner: DS)


### [2025-08-27 00:19 Local] v1 Actions — UX build checklist

#### Form fields
- SQL editor: multiline, monospace, line numbers, soft-wrap toggle
- Dialect selector: Postgres | MySQL | Generic
  - Help text: “Generic = permissive checks only; choose a dialect for stricter parsing.”
- Timeout (seconds, optional)
  - Empty uses default 60s; validate range (1–3600)
- Notes (optional)

#### Parameters panel
- Auto-detect unique `{{name}}` placeholders from SQL and list them
- Allow editing description per parameter (no types in v1)
- Indicate values are provided at Job creation (not here)
- Warnings:
  - Placeholder in list but not in SQL → warning
  - Placeholder used in SQL but not documented in panel → auto-add

#### Validation controls
- Button: “Check syntax (pre-DB)”
  - Generic: lexical checks only (balanced quotes/parentheses; valid placeholder names; no unterminated strings)
  - Dialect-selected: run dialect parser with placeholders substituted by safe literals (parse-only)
- In Job editor (post-bind): “Validate against DB”
  - Execute in transaction and rollback; capture errors; optional EXPLAIN

#### Error/warning display
- Inline annotations in editor (squiggles)
- Side panel list with entries: { code, message, line, column }
- Codes: SYNTAX_ERROR, UNBALANCED_DELIMITER, DIALECT_PARSE_ERROR, TIMEOUT, CONNECTION_ERROR (post-bind)

#### Edge cases
- Empty SQL or only comments → invalid (block save)
- Duplicate placeholders allowed; resolve order by first appearance for arg mapping
- Placeholders must match `[A-Za-z_][A-Za-z0-9_]*`
- Semicolons permitted (multi-statement) where DB allows

#### Save rules
- Action can be saved after pre-DB check passes (or user proceeds in Generic)
- DB-backed validation available later in Job editor once a connection is selected

#### Notes for v1.x (multi-step forward-compat)
- Keep visual hierarchy so this form can later become “Step 1” of a steps list
- Reserve placement for “Expectation matchers” and “Failure policy” panels (hidden in v1)
