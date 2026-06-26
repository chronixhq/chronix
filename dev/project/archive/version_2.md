# Chronix Version 2 — CLI, Resiliency, and Enterprise Security

Owner: dsherwin
Last updated: 2026-01-06
Status: Planning

## Executive Summary
Version 2 focuses on expanding Chronix into a robust, CLI-driven platform with enhanced resiliency, transactional safety, and enterprise-grade security features. This phase builds upon the core execution engine established in Version 1.

## Scope

### 1. CLI Expansion (The "Terminal First" Experience)
Provide a comprehensive set of subcommands using `kong` for management without the Web UI.
- **Jobs**: List, enable/disable, run (immediate trigger), and detailed info.
- **Connections**: List all types and run connectivity tests.
- **Runs**: List history with filtering and stream logs for specific executions.
- **Agents**: List registered agents and health status.

### 2. Configuration & Portability
- **TOML Support**: Implement `chronix.toml` support for global configuration.
- **Precedence Logic**: CLI Flags > Env Vars > TOML File > Persisted DB Settings > Defaults.
- **Data Dir Management**: Improved defaults and overrides for `CHRONIX_DATA_DIR`.

### 3. Resiliency & Reliability
- **Step-Level Retry Logic**: Add `retry_policy` (max retries, backoff strategy) to Action Steps.
- **Graceful Handling**: Improved handling of "stale" jobs on server restart (auto-marking as failed or resuming if idempotent).
- **Session Persistence**: Maintain connection state (temporary tables, variables, working directories) across multiple steps in a single job execution.

### 4. Advanced Admin & Security
- **Role-Based Access Control (RBAC)**: Implement granular permissions beyond the simple Admin flag.
- **User Management CLI**: Subcommands to add, enable, or disable users.
- **Token Management**: Issue and revoke API keys via CLI.

### 5. Transactional Safety
- **SQL Sandbox Mode**: Implement transactional rollback for SQL testing. Allow running actions in a "sandbox" where all changes are rolled back after execution, ensuring production data remains safe during testing.

## Deliverables
1. Updated `cmd/app` with the new subcommand structure.
2. `internal/config` package for unified TOML/Env/Flag parsing.
3. Logic updates in `internal/executor.go` for step-level retries and session persistence.
4. RBAC implementation in `internal/auth` and database schema.
5. SQL Sandbox support in `internal/sqlrunner`.
6. Documentation updates for CLI usage and configuration.

## Milestones
- **M1: CLI & Configuration**: CLI scaffolding and TOML parsing.
- **M2: Resiliency**: Step-level retries and session persistence.
- **M3: Transactional Safety**: SQL Sandbox mode.
- **M4: Security**: RBAC and User Management CLI.
- **M5: Final Polish**: Documentation and acceptance testing.
