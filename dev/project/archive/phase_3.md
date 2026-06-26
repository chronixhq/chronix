# Chronix Phase 3 — CLI Expansion, Configuration, and Resiliency

Owner: dsherwin
Last updated: 2026-01-05
Status: Draft

## Executive Summary
Phase 3 transitions Chronix from a UI-primary tool to a fully-featured CLI-driven platform. We will implement a rich set of subcommands for managing jobs, connections, and runs directly from the terminal. Additionally, we will introduce TOML configuration support and implement job-level resiliency features like automatic retries.

## Scope

### 1. CLI Expansion (The "Terminal First" Experience)
We will add high-quality subcommands using `kong` to provide management capabilities without needing the Web UI.
- **Jobs**:
  - `chronix job list`: Show all jobs, status, and next run time.
  - `chronix job enable/disable <id>`: Toggle job state.
  - `chronix job run <id>`: Trigger an immediate execution (Run Now).
  - `chronix job info <id>`: Detailed view of job config and variables.
- **Connections**:
  - `chronix connection list`: Show all Database, Shell, and Web Task connections.
  - `chronix connection test <id>`: Run a connectivity check.
- **Runs**:
  - `chronix run list`: Show recent job run history with filtering options.
  - `chronix run logs <run-uid>`: Stream or cat logs for a specific execution.
- **Agents**:
  - `chronix agent list`: Show registered agents and their health.

### 2. Configuration & Portability
- **TOML Support**: Implement `chronix.toml` support. Precedence: CLI Flags > Env Vars > TOML File > Persisted DB Settings > Defaults.
- **Data Dir Management**: Better defaults and overrides for `CHRONIX_DATA_DIR`.

### 3. Resiliency & Reliability
- **Retry Logic**: Add `retry_policy` to Action Steps (max retries, backoff strategy).
- **Graceful Handling**: Improved handling of "stale" jobs on server restart (auto-marking as failed or resuming if idempotent).

### 4. Advanced Admin
- **User Management**: `chronix user add/enable/disable` for bootstrap scenarios.
- **Token Management**: Issue and revoke API keys via CLI.

## Deliverables
1. Updated `cmd/app` with the new subcommand structure.
2. `internal/config` package for unified TOML/Env/Flag parsing.
3. Logic updates in `internal/executor.go` for step-level retries.
4. Documentation update in `README.md` and CLI `--help`.

## Milestones
- **M1: CLI Scaffolding**: Register subcommands and implement "List" views.
- **M2: CLI Operations**: Implement "Run", "Toggle", and "Test" commands.
- **M3: Configuration**: Implement TOML parsing and precedence logic.
- **M4: Resiliency**: Implement retry logic in the executor.
- **M5: Final Polish**: User management CLI and documentation.

## Acceptance Criteria
- Admin can perform full job/connection lifecycle via CLI.
- Configuration can be driven entirely by a `chronix.toml` file.
- Failed steps automatically retry based on the configured policy.
- All CLI commands provide clear, structured output (and optional JSON output for automation).
