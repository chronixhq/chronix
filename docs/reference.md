---
title: Reference Manual
description: Technical specifications for CLI and configuration.
---

# Chronix Comprehensive Reference Manual

This document provides a detailed technical reference for all aspects of Chronix, including CLI commands, system configuration, task execution details, and administrative operations.

---

## 1. CLI Commands Reference

The Chronix binary supports several subcommands for service management and administration.

| Command | Description |
| :--- | :--- |
| `run` | Starts the Chronix server. Supports network override flags: `--disable-http`, `--disable-https`, `--force-http-port`, `--force-https-port`, `--force-agent-port`. |
| `service` | Manage the native OS service (install, uninstall, start, stop, status). Note: macOS and Linux require `sudo` for these commands. |
| `stop` | Gracefully shuts down the running Chronix server daemon. |
| `status` | Reports the current operational state of the server. |
| `restart` | Restarts the Chronix server daemon. |
| `adminCode` | Generates a new one-time admin login code. |
| `agents` | Manage connected agents (list, update). |
| `update` | Manage Chronix server updates (check, apply). |
| `version` | Shows the current version and release notes. |
| `systemdata` | Shows live system resource usage. |
| `suspendServer` | Pauses all job scheduling and execution. |
| `unsuspendServer` | Resumes job scheduling and execution. |

### Global Flags and Environment Variables
- `-D`, `--datadir`: Path to the data directory (overrides default).
- `CHRONIX_DATA_DIR`: Environment variable to set the data directory.
- `-c`, `--conf`: Path to a configuration file (reserved for future use).

---

## 2. Agent CLI Commands Reference

The `chronix-agent` binary supports subcommands for registration, service management, and local control.

| Command | Description |
| :--- | :--- |
| `run` | Starts the agent in the foreground (default behavior). |
| `register` | Registers the agent with a Chronix server. |
| `info` | Shows the saved agent identity, server target, pin status, local runtime state, and a quick server probe. |
| `repoint` | Updates the saved server host/port without changing the agent UUID, keys, or name. |
| `repair-register` | Refreshes the existing agent identity with the server and can optionally update the saved server target or name. |
| `rename` | Changes the local agent name and refreshes the server-side agent record using the existing UUID and keys. |
| `unregister` | Removes the local agent identity and registration. |
| `reset` | Clears the pinned server certificate (TOFU reset). |
| `service` | Manage the native OS service (install, uninstall, start, stop, status). |
| `stop` | Gracefully shuts down the running agent daemon. |
| `status` | Reports the current operational state of the agent. |
| `restart` | Restarts the agent daemon. |
| `version` | Shows the current version and release notes. |

---

## 3. System Configuration

Chronix settings are stored in the database and can be managed via the **Settings** page in the Web UI.

### Network Settings
Settings are stored in the database but can be overridden at startup via CLI flags (see `run` command).
- **HTTP Enabled/Port**: Toggle the insecure HTTP server (default: `5170`). Override: `--disable-http`, `--force-http-port`.
- **HTTPS Enabled/Port**: Toggle the secure HTTPS server (default: `5171`). Override: `--disable-https`, `--force-https-port`.
- **HTTPS Mode**: Supports `selfsigned` or manual PEM injection.
- **Agent Port**: Port for incoming Agent WebSocket connections (default: `5172`). Override: `--force-agent-port`.

### Security
- **Admin Login Codes**: One-time codes are valid for 10 minutes.
- **TOFU (Trust On First Use)**: Agents pin the server's TLS certificate on the first connection. If the certificate changes, the agent will refuse to connect until reset.
- **Encryption**: Sensitive fields like passwords and private keys are encrypted at rest using a master key (`master.key`) located in the data directory.

---

## 4. Task Type Specifications

### SQL Tasks
- **Supported Engines**: SQLite, PostgreSQL, MySQL, MariaDB, MS SQL Server (T-SQL), Oracle, Snowflake.
- **Dialects**: Choose between `generic`, `sqlite`, `postgres`, `mysql`, and `tsql` (for MS SQL Server).
- **Dialect Handling**: Dialect-specific syntax is supported per action.
- **Session Persistence**: Multi-step actions use a single, persistent database connection. This ensures that session-scoped state, such as `last_insert_rowid()` in SQLite or temporary tables, is preserved across all steps in the action.
- **Expectation Types**:
    - `rowsAffected`: Compare against the count of modified rows.
    - `rowExists`: Verify if a query returns at least one row.
    - `noRowsReturned`: Verify that a query returns zero rows.
    - `fieldEqualsFirst`: Verify if a specific field in the first row matches a value.
    - `fieldEqualsLast`: Verify if a specific field in the last returned row matches a value.
    - `fieldEquals`: Backward-compatible alias for first-row equality.
    - `noError`: Simply verify the query executed without an error.
- **Output Capture**: Support for capturing query results into variables for use in subsequent steps.
- **Connection Management**: Connections can be duplicated with a single click, allowing for rapid creation of similar environments (e.g., dev/staging/prod) without re-entering credentials.

### Shell Tasks
- **Modes**:
    - `localhost`: Runs as the user owning the Chronix process.
    - `ssh`: Runs via SSH.
- **SSH Authentication**: Supports Password and Private Key (Ed25519). Supports OpenSSH and PKCS#8 PEM formats for private keys.
- **Run Modes**: Supports `command` (single line) or `script` (multi-line block).
- **Output Capture**: Configurable truncation (Head/Tail) and max byte limits (default: 64KB). Capture specific output into variables.

### Web Tasks (HTTP)
- **Methods**: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`.
- **Variable Capture**:
    - **JSONPath**: Extract values from JSON responses (e.g., `$.status`).
    - **Regex**: Extract values from any response body via capture groups.

---

## 5. Scheduling Engine

Chronix uses a high-precision scheduler that supports multiple definition formats.

- Same-job overlap is prevented by the job queue. A job cannot be queued again while it is already queued or running.

### JSON Structure
Schedules are stored as JSON in the database:
- `kind`: `single` or `recurring`.
- `cron`: Standard 5-field cron string (e.g., `0 0 * * *`).
- `structured`: A nested object for human-friendly recurrence rules.
- `startAt` / `endAt`: Lifecycle boundaries for the job.

### Cron Syntax
Supports the standard `minute hour dom month dow` format.
Example: `*/15 * * * *` (Every 15 minutes).

---

## 6. User and Admin Management

### Roles
- **Admin**: Full access to all settings, connections, actions, and user management.
- **User**: Access to view activity and manage their own profile. (Permissions for specific resources are expanded in Phase 2/3).

### Password Policy
- Passwords are hashed using `bcrypt`.
- Admins can force-reset user passwords or disable accounts.

### Password Recovery (CLI-Based)
Since Chronix does not support SMTP-based password recovery by design, the CLI must be used for administrative recovery:
1.  Run `chronix adminCode` to generate a 10-minute temporary login token.
2.  Navigate to `/settings` in the web UI.
3.  Authenticate with the code to access the "Setup Admin" profile (ID 0).
4.  Use this access to modify existing users and reset their passwords.

---

## 7. Notification and Reporting

Chronix supports sophisticated alerting and reporting mechanisms.

### Alerting Policies
- **Success/Failure**: Per-job toggle for notifications on successful or failed runs.
- **Reporting (Include Output)**: When enabled, the notification payload includes snapshots of the task output:
    - **SQL**: Includes rows affected or rows count.
    - **Shell**: Includes truncated `stdout` and `stderr`.
    - **Web**: Includes the HTTP response body.

### Delivery Channels
- **Email (SMTP)**: Formats reporting data into Markdown-style tables and code blocks.
- **SMS (Twilio)**: Provides a high-level summary of the final step's outcome.
- **Webhooks (Outgoing)**: Sends a JSON POST request to a configured URL. Includes an `X-Chronix-Signature` header (HMAC-SHA256) if a secret is provided.

---

## 8. Activity & Reporting Engine

Chronix maintains a unified activity log combining job executions and user actions.

### Unified Query Logic
The activity system uses a SQL `UNION ALL` approach to combine `job_runs` and `user_activity` tables, allowing for efficient server-side pagination and filtering across disparate data types.

### Export Formats
- **CSV**: A flat file containing all selected activity fields.
- **HTML**: A standalone, styled document suitable for browser viewing or printing.
- **PDF**: A professional, paginated report generated using the `fpdf` library.

---

## 9. Feedback System

The Feedback system allows users to communicate with administrators directly through the platform.

### Data Storage
Feedback data is stored in three primary tables:
- `bug_reports`: Stores bug summaries, descriptions, and current status.
- `feature_requests`: Stores feature request summaries, descriptions, and current status.
- `feedback_attachments`: Stores metadata for files uploaded by users or admins.

### File Storage
Uploaded attachments are stored in the `feedback/` subdirectory of the Chronix data directory. Filenames are obfuscated using a combination of the report ID and a high-precision timestamp to prevent enumeration.

### Status Enum
The following statuses are supported for both bug reports and feature requests:
- `open`
- `in-progress`
- `resolved`
- `closed`
- `rejected`

---

## 10. Branding and Suspension State

Chronix is fully open source and does not gate premium features behind license checks.

### Suspension Logic
Jobs, agents, actions, and connections can be marked as **Suspended** by administrative workflows or safety checks. Suspended items remain in the database but are ignored by the scheduler and runner until they are re-enabled.

### Branding Configuration
Custom branding settings are stored in the `cx_settings` table:
- `brand_logo_url`: Remote or local URL for the primary UI logo.
- `brand_name`: Text identity for the application.

---

## 11. The Chronix Agent System

Agents are lightweight workers that connect back to the server via an encrypted WebSocket tunnel.

- **WebSocket Tunnel**: Uses `wss://` on port 5172 (default) for secure, encrypted command and control.
- **Phone Home**: Agents initiate the connection, bypassing the need for inbound firewall rules on the agent's network.
- **JWT Authentication**: Agents authenticate using a unique JWT signed with an Ed25519 key pair generated locally during registration.
- **OS Tracking**: Agents capture and report the local OS user account, distribution (Linux), marketing name and version (macOS), detailed version and build info (Windows), and architecture. This information is collected during registration and refreshed on every connection to ensure administrators have the most accurate context for the agent's execution environment.
- **Identity Repair**: Agents can repoint to a new server address or refresh their existing registration without generating a new UUID. This helps recover cleanly from server address changes, rebuilt servers, or agent rename operations.
- **Relay Updates**: Agents can be updated or reverted to specific versions via the server. The server acts as a relay, caching binaries from the distribution server and serving them to agents over an authenticated HTTPS endpoint (`/agent/update/:version/:platform`).

---

## 12. Maintenance and Updates

Chronix features a robust update system for both the server and its connected agents.

### Installation Script
The easiest way to install or update Chronix is via the official installation script:
`curl -fsSL https://chronixhq.com/install.sh | sh`

### Update Manifests
Chronix checks for updates by fetching JSON manifests from `dist.chronixhq.com`.

**Distribution URLs:**
- `latest.json`: `https://dist.chronixhq.com/latest.json`
- `versions.json`: `https://dist.chronixhq.com/versions.json`
- **Latest Binaries**: `https://dist.chronixhq.com/latest/<platform>/<app>`
  *(Where app is `chronix` or `chronix-agent`)*

### CLI Update Commands
- `update check`: Queries the distribution server (or the running daemon via RPC) to see if a newer version is available.
- `update apply`: Downloads the latest binary, verifies its SHA256 hash and Ed25519 signature, and replaces the current executable. If the daemon is running, it is automatically restarted.

### UI Update Awareness
- When updates are available and updater mode is not set to full automatic, Chronix surfaces notices in both the main Dashboard and **Settings > Overview**.
- Those notices link directly to **Settings > Updates** and can be dismissed for the current browser session.

### Agent Versioning
- `agents list`: Displays the current version of all connected agents. It also checks for the latest available version and flags agents that are out of date.
- `agents update <uuid> [version]`: Sends a command to the specified agent(s) to download and apply an update. If `version` is omitted, the latest version is used.
- `agents update all [version]`: Sends the update command to all online agents. In the Web UI, **Settings > Updates** also offers an **Update All** action when multiple agents are out of date.

### Chronix and Agent Versioning Policy
- Chronix server and `chronix-agent` do not need to share the exact same version number, even though they live in the same repository.
- `major.minor` should stay aligned when the server and agent belong to the same compatibility train. Matching `major.minor` versions are expected to be compatible with each other.
- `patch` versions may move independently for server-only or agent-only maintenance work.

**Release rules:**
- Server-only UI, docs, help, or maintenance fixes: bump `chronix` patch only.
- Agent-only CLI, runtime, or recovery fixes: bump `chronix-agent` patch only.
- Shared protocol, registration, update, or lifecycle behavior changes: bump both components into the next shared `minor`.
- Breaking compatibility between server and agent: bump both components to the next `major`.

### Advanced Version Management
- `version list`: Lists all available versions for both the server and agents.
- `version revert <version>`: Force-installs a specific version of the Chronix server, whether it is an upgrade or a downgrade.

---

## 13. Data Management

### Data Directory
By default, Chronix stores all files in:
- **Linux**: `/var/lib/chronix`
- **macOS**: `~/Library/Application Support/Chronix`
- **Windows**: `C:\ProgramData\Chronix`

### Critical Files
- `chronix.db`: The SQLite database containing all state.
- `master.key`: The key used for encrypting credentials. **Do not lose this file.**
- `cert.pem` / `key.pem`: TLS certificates for HTTPS and Agent connections.
