# Project Guidelines (Living Document)

Purpose
- Capture opinionated, practical conventions for this codebase.
- Keep services consistent, observable, and easy to operate.
- Evolve per project; upstream reusable improvements to your template.

Owner: dsherwin
Last updated: 2026-01-29

Current Status & Accomplishments (as of 2026-01-29)
- Pre-commit/Push Protocol: Established a standardized workflow triggered by the phrase "We are set to do another commit and push" to ensure code quality, consistency, and updated documentation across chronix-agent, chronix, and the React app.
- Planning: Added dev/project/phase_2.md (“Phase 2 — Make It Actually Work”) and linked it from README for visibility; this plan aligns work with these guidelines.
- CLI & Daemon: main wiring gates non-run subcommands via processCommand(); `run` path starts daemon services and calls rest_api.Start(). If uninitialized, an admin login code is generated and printed to the console.
- Logging: Using log/slog in main with structured messages. Note: standardize error field key to "error" consistently across the codebase in upcoming refactors.
- REST Harness: rest_api.Start() exists as the server entrypoint; readiness/health endpoints and restresponse standardization are part of Phase 2 and should be used consistently.
- Repo Layout: Repository follows the documented layout (cmd, internal, rest_api, dev, etc.).
- Frontend: Internal React app scaffold present under internal/chronix_react_app/ with modules (e.g., User/ForgotPassword, ScheduledJobs/CreateScheduledJob, Testing). Guidelines enforce @dsherwin/react-api-interface and base URL via AuthContext.
- Docs: Cowboy methodology doc present; README updated with Phase 2 plan reference.


1. Project Layout
- Follow this project’s layout as default:
  - cmd/app/: application entrypoint, CLI, logging init, rpc, db init
  - cmd/app/db/: generated GORM models and DB access
  - cmd/app/systemdata/: background jobs tied to app lifecycle
  - internal/: domain logic, collectors, integrations (gnmi, netconf, etc.)
  - rest_api/: Gin handlers and route wiring
  - build/: build artifacts and config samples
  - dev/: run configurations, tools support, schema.sql
  - docs/: documentation (auto-generated preferred)
- Variant: when it simplifies boundaries, place db package under internal instead of cmd/app.

Rationale: Keep the executable-specific wiring in cmd, core logic in internal, and REST-specific code isolated in rest_api.


2. Language, Style, and Tooling
- Go version: track stable in go.mod; upgrade after verification.
- Formatting: gofmt + goimports; enforce via CI.
- Linting: golangci-lint with sensible defaults; justify ignores in code.
- Logging: use log/slog everywhere.
  - Default to structured output and consistent field keys.
  - Standard keys: app, version, commit, build_date, pid, user, component, op, id, device, serial, error.
  - Levels: debug (diagnostics), info (lifecycle), warn (recoverable issues), error (failures).
  - Linux: prefer JournaldHandler by default. macOS: TextHandler to stdout.
  - No log.Fatal in libraries; prefer returning errors and logging at edges.
  - Plan for centralized logging backends; keep logger wiring clean to swap handlers.
- Concurrency: tie goroutines to context; prefer errgroup when fanning out; avoid leaks.
- Panics: only for unrecoverable init errors; otherwise return errors.
- Style: follow Go naming conventions; keep locals unexported unless needed; keep slog key names consistent and avoid spelling drift.
- CLI: use kong for CLI parsing with subcommands; keep commands small and composable; offer shell completions via kongplete.


3. Configuration
- Use app_settings for application settings.
  - Register settings in init() and call app_settings.Setup() during startup.
  - Provide RPC exposure for listing live settings (see rpc.SocketPath usage).
- Precedence: CLI flags > environment variables > persisted app_settings > defaults.
- Common env vars: LOG_LEVEL, HTTP_LISTEN_ADDR, RPC_SOCKET_PATH.
- Configuration files: prefer TOML when multiple files are needed; document paths.
- Data directory conventions: default per OS:
  - Linux (root): `/var/lib/<app>`
  - Linux (user): `~/.local/share/<app>` (respects `XDG_DATA_HOME`)
  - macOS: `~/Library/Application Support/<App>`
  - Windows: `C:\ProgramData\<App>`
- Allow override via `CHRONIX_DATA_DIR` or app-specific `DATA_DIR` env; ensure directory exists at startup.
- Validate config centrally and fail fast with clear messages.
- Never log secrets; redact sensitive values in logs and errors.


4. REST/HTTP
- Framework: Gin, with rest_api_server as the default server harness.
  - Listening address managed via app_settings (see http_listening_address).
- Middleware: central auth, logging, recovery. Keep per-route logic thin.
- API Responses: use restresponse helpers everywhere for consistency.
  - restresponse.RestSuccess(c, data)
  - restresponse.RestSuccessNoContent(c) for empty or trivial success responses (e.g., {"status":"ok"})
  - restresponse.RestErrorRespond(c, restresponse.<Code>, message, details...)
  - Do not use c.AbortWithStatusJSON in handlers; prefer restresponse.RestErrorRespond. After sending an error from a handler, just return. Reserve c.Abort for middleware that short-circuits before c.Next().
  - Map domain errors to stable restresponse codes (use the Code enum and HTTPStatusFromCode).
- Start HTTP server after RPC listener is ready.
- Endpoints: provide /healthz and /ready via rest_api_server if applicable.
- Route prefix: Do not use an /api prefix in the frontend; HTTP routes are mounted at the root (e.g., "/connections", not "/api/connections").
- Timeouts and cancellation: ensure per-request contexts are respected.
- Logging: consider JSON logs in production for aggregation.
- Security: validate inputs; only enable CORS as required.
  - Cookies/JWT: set HttpOnly; set Secure in production; use short TTLs for admin/session-elevation tokens; avoid putting sensitive data in JWT claims; rotate secrets.

Current note: Some handlers still use c.JSON directly; migrate to restresponse helpers as routes are touched.


5. Error Handling
- Avoid panics in production paths. Log and return (or exit with non-zero status at process edges).
- Prefer error wrapping with %w to preserve causality.
- Provide actionable context in messages and logs (who/what/where):
  - return fmt.Errorf("collect optics for device %s: %w", deviceID, err)
  - slog.Error("collect optics", "device", deviceID, "error", err)
- Use sentinel errors sparingly (mainly for control flow), rely on errors.Is/As.
- Don’t both log and return the same error deep inside libraries; log at boundaries.
- Use key name "error" consistently for error fields in logs.


6. Contexts and Shutdown
- Long-running goroutines must accept context or expose explicit Stop/Shutdown methods.
- Trap signals: SIGINT, SIGTERM, SIGQUIT, SIGHUP. Note: SIGKILL cannot be trapped.
- On shutdown:
  - Stop tickers/workers.
  - Close RPC listener.
  - Shutdown HTTP server (gracefully) if supported.
- Ensure cancellation is propagated to fanned-out work (use errgroup when applicable).


7. RPC
- Prefer Unix domain sockets by default; make path configurable (e.g., under XDG_RUNTIME_DIR for non-root users).
- Default socket permissions: 0660; optionally set group ownership for shared access. Ensure parent directory perms 0770 when creating.
- Prefer placing sockets under XDG_RUNTIME_DIR for non-root users; fall back to a temp dir when unavailable.
- Ensure clients close connections or use an rpc.Call helper that dials and closes per call.
- Validate inputs and return typed errors from handlers.


8. HTTP
- Start server after RPC is ready (ordering matters for readiness).
- Provide /healthz and /ready endpoints where applicable; integrate with deployment readiness checks.
- Use structured JSON logs in production for log aggregation.


9. Database
- ORM: GORM as the standard.
- Model generation: use gormdb2struct to generate structs from the DB schema (PostgreSQL/SQLite).
  - See .gormdb2struct.toml for sample config.
  - Use the “Rebuild Database Structs” run configuration as a reference workflow.
  - Important (SQLite): gormdb2struct currently panics when a column has a DEFAULT in SQLite schemas. Avoid DEFAULT clauses in dev/schema.sql. If you need defaults, enforce them at the application layer or via triggers. When updating the schema:
    1) Remove DEFAULT clauses from schema.sql.
    2) Rebuild internal/db/assets/schema.db with: sqlite3 internal/db/assets/schema.db ".read dev/schema.sql"
    3) Re-run: go run github.com/dan-sherwin/gormdb2struct@latest .gormdb2struct.toml
- Generated query objects usage:
  - After db.SetDefault(gormDB), prefer db.<Model> (e.g., db.Notification) directly instead of db.Q.<Model>.
  - Example: db.Notification.Where(db.Notification.Category.Eq("job")).Order(db.Notification.CreatedAt.Desc()).Limit(20).Find()
- Migrations: maintain schema.sql in dev/ (or a migrations tool in future); keep schema as the source of truth for generation.
- Transactions: pass context and tx explicitly; keep SQL concerns in the db layer.


10. Dependencies
- Pin major.minor Go version in go.mod (e.g., `go 1.24`).
- Run `go mod tidy` and `go vet` in CI.
- Prefer `golangci-lint` locally and in CI; keep a curated ruleset and justify ignores.


11. Versioning
- Inject build info via -ldflags: Version, Commit, BuildDate.
- Log build info at startup and expose via a CLI command.


12. Testing
- Strategy: prefer integration and end-to-end tests over exhaustive per-function unit tests.
  - Still include focused unit tests for RPC handlers and helpers.
  - Use table-driven tests where they add clarity.
  - Use -short or build tags to separate slow tests when needed.
- Use the race detector in CI for tests (`-race`).
- Avoid tests that depend on /var/run paths; use temp directories/sockets instead.
- Observability in tests: emit structured logs to aid troubleshooting.
- Aim for meaningful coverage, not numeric goals. Prioritize correctness and regressions.


13. Observability (Logs, Metrics, Tracing)
- Logs: slog structured logs with consistent field keys.
- Metrics: expose Prometheus metrics when relevant; document key SLIs.
- Tracing: adopt OpenTelemetry when cross-service visibility is needed; propagate context.


14. CI/CD
- CI gates: fmt, lint, build, tests (with race), govulncheck, vet, and mod tidy verification.
- Releases: tag semantic versions; keep changelog entries in PRs.


15. Documentation
- Prefer automated documentation generation for functional and file-level docs.
- Generate README.md content using tooling/AI; keep it accurate and up-to-date.
- Keep concise HOWTOs for local dev, running, and debugging.


16. Code Review
- Small focused PRs; descriptive titles; include tests when applicable.
- Block on lint/test failures; resolve or justify comments.


17. Housekeeping & Security
- Remove dead code promptly.
- Keep dependencies current (Renovate/Dependabot or scheduled updates).
- Restrict socket permissions to 0660; set group ownership when required.
- Avoid logging secrets; redact sensitive fields in logs and errors.

18. Collaboration & Task Review
- If a requested task appears egregiously wrong or there’s an obviously better approach, pause and ask clarifying questions before proceeding. The goal is to avoid harmful changes and converge on the best path quickly.


19. Frontend (React App)
- For the internal React UI under internal/chronix_react_app/ use @dsherwin/react-api-interface for all HTTP calls (apiGet, apiPost, apiPut, apiPatch, apiDelete). Do not use window.fetch directly.
- Base URL is configured in the app’s AuthContext via setAPIBaseURL; dev uses http://localhost:6060, production uses window.location.origin.
- Example: const res = await apiPost('/initialize', payload); const data = await apiGet('/server/status').
- Error handling: apiGet returns parsed JSON; apiPost/apiPut return a Response-like object — check res.ok or use helpers as they are introduced.
- MUI version: We standardize on Material UI v7.3.4 for this app. Always consult v7.3 docs (https://mui.com/material-ui/react-button/) and APIs for guidance, not older v5/v6 content. Prefer the API search scoped to v7.3.
  - Do: use `slotProps` APIs and `sx` styling; follow v7 component prop names and deprecations.
  - Do not: use legacy `@mui/styles`, `makeStyles`, or v5-only patterns; avoid examples that rely on deprecated props.
  - Notes specific to v7.3.4 we rely on right now:
    - `TextField` native input attributes go under `slotProps.htmlInput` (not `inputProps`).
    - `Typography` `paragraph` prop is deprecated in v7; use the `component` prop instead when you need a `<p>` element.
- **Imports & Linting:** Whenever you make ANY changes to React files, you MUST check the imports to ensure they are correct and complete. After making changes to the React app, you MUST run `npm run lint` within `internal/chronix_react_app/` to ensure no missing imports or syntax errors.
- MUI TextField: inputProps is deprecated. Use slotProps.htmlInput for native input attributes. For numeric-only input, prefer:

```tsx
<TextField
  label="Value"
  placeholder="1"
  value={(step.expectation as ExpectRowsAffected).value || ''}
  onChange={(e) => onChange({
    kind: 'rowsAffected',
    op: (step.expectation as ExpectRowsAffected).op || '>=',
    value: e.target.value.replace(/[^0-9]/g, ''),
  })}
  sx={{ minWidth: { xs: '100%', md: 160 } }}
  slotProps={{
    htmlInput: {
      inputMode: 'numeric',
      pattern: '[0-9]*',
    },
  }}
  error={!!fieldErrors?.expectationValueNum}
  helperText={fieldErrors?.expectationValueNum}
/>
```

20. Pre-commit/Push Protocol
Whenever the user states "We are set to do another commit and push", automatically perform the following:
1. chronix-agent:
   - Perform full code analyzation.
   - Perform optimizations, cleanup, and remove unused code.
   - Run `cmd/chronix-agent/dev/agent-ci-local.sh` and ensure it passes.
2. chronix & chronix react:
   - Perform full code analyzation.
   - Perform optimizations, cleanup, and remove unused code.
   - Run `dev/ci-local.sh` and ensure it passes.
3. Documentation:
   - Update `docs/help.md`, `docs/reference.md`, `docs/features_and_capabilities.md`, and `docs/testing_checklist.md` with all recent updates and additions.

Changelog (for this document)
- 2026-01-29: Updated Frontend: Added mandatory check for imports whenever React files are changed. Updated Last updated date.
- 2026-01-28: Frontend: Added requirement to verify imports and run `npm run lint` for any React changes. Updated Current Status and Last updated date.
- 2026-01-18: Added Section 20 "Pre-commit/Push Protocol" triggered by "We are set to do another commit and push". Updated "Current Status" and "Last updated" date.
- 2025-11-12: Frontend: document MUI TextField `inputProps` deprecation; require `slotProps.htmlInput` usage with numeric input example. Updated Last updated date.
- 2025-09-25: Updated REST/HTTP guidelines to deprecate apiresponse in favor of restresponse; require RestSuccess/RestSuccessNoContent/RestErrorRespond usage consistently across handlers. Updated Current Status to reflect restresponse standardization.
- 2025-09-24: Updated Last updated date; added “Current Status & Accomplishments” summarizing verified repo state (CLI run gating, rest_api harness, admin code flow, React app scaffold); linked Phase 2 plan from README.
- 2025-09-13: Database: clarified gorm/gen convention — after db.SetDefault, prefer db.<Model> (e.g., db.Notification) over db.Q.<Model> throughout the codebase.
- 2025-09-11: Added guidance to question egregiously wrong tasks or propose better approaches before proceeding; updated Last updated date accordingly.
- 2025-09-11: Documented React app API usage: use @dsherwin/react-api-interface (apiGet/apiPost/etc.) for all HTTP calls; do not use window.fetch; noted base URL configuration in AuthContext.
- 2025-09-10: Refined to match current Chronix repo conventions: added CLI guidance (kong + kongplete), documented OS-specific data directory defaults and CHRONIX_DATA_DIR override, clarified RPC socket directory perms (0770) and XDG_RUNTIME_DIR preference, expanded REST security (cookies/JWT best practices), and minor wording/consistency updates.
- 2025-09-09: Initial version and alignment with preferences and operational details: slog with standard keys and handlers, Gin + rest_api_server with apiresponse, app_settings precedence (CLI > env > persisted > defaults), Unix socket RPC with 0660 perms, ldflags build info, race-enabled tests, GORM + gormdb2struct, integration-first testing, automated docs.
