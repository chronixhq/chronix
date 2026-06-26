# Chronix Release Testing Checklist

This document is the living manual testing suite for Chronix releases. Use it before broader rollouts so the core product, agent workflows, update paths, and integrations are exercised without relying on memory.

- [ ] **Documentation & Inline Help Sync**:
  - [ ] Verify changed workflows are reflected in `docs/help.md`.
  - [ ] Verify affected `SectionHelp` icons in the React UI still open the correct and current help content.
  - [ ] Verify command and lifecycle changes are reflected in `docs/reference.md`.

## 📋 Table of Contents
1. [Fresh Bootstrap & Admin Initialization](#1-fresh-bootstrap--admin-initialization)
2. [User Management & RBAC](#2-user-management--rbac)
3. [Connection Matrix: SQL Databases](#3-connection-matrix-sql-databases)
4. [Connection Matrix: Shell & Scripts](#4-connection-matrix-shell--scripts)
5. [Connection Matrix: Web Tasks (API)](#5-connection-matrix-web-tasks-api)
6. [Action Sandbox & Assertion Engine](#6-action-sandbox--assertion-engine)
7. [Job Scheduling & Execution Engine](#7-job-scheduling--execution-engine)
8. [The Agent System (Remote Workers)](#8-the-agent-system-remote-workers)
9. [Observability, Reports & Alerts](#9-observability-reports--alerts)
10. [Open Source Readiness](#10-open-source-readiness)
11. [Infrastructure & Resilience](#11-infrastructure--resilience)

---

## 1. Fresh Bootstrap & Admin Initialization
*Goal: Verify the zero-dependency, self-contained startup flow.*

- [ ] **Data Directory Creation**:
  - [x] Set `CHRONIX_DATA_DIR` to a non-existent path.
  - [x] Run `./chronix run`.
  - [x] Verify directory is created with `chronix.db` and `master.key`.
- [ ] **Interactive Installation (install.sh)**:
  - [ ] Run `curl -fsSL https://chronixhq.com/install.sh | sh`.
  - [ ] Verify platform detection and manual override.
  - [ ] Verify selection of installation directory.
  - [ ] Verify binary download and permission setup.
  - [ ] Verify interactive protocol (HTTP/HTTPS) and port configuration.
- [ ] **Admin Code Flow**:
  - [ ] Verify 12-char code appears in console.
  - [ ] Attempt login at `http://localhost:5170` with an *incorrect* code (Verify error).
  - [ ] Login with the *correct* code.
- [ ] **Initial Configuration**:
  - [ ] Fill "Create Admin Profile".
  - [ ] Set "Server URL" to `http://localhost:5170`.
  - [ ] Verify redirect to Dashboard.
  - [ ] Verify "System initialized" entry in Activity Timeline.

## 2. User Management & RBAC
*Goal: Verify sovereignty and permission boundaries.*

- [ ] **Admin Profile**:
  - [ ] Change Name and Email. Verify update.
  - [ ] Change Password. Log out and back in with new password.
- [ ] **User Roles**:
  - [ ] Create **Regular User**.
  - [ ] Log in as Regular User.
  - [ ] Verify "Admin" and "Settings" menu items are hidden.
  - [ ] Try to access `/settings/overview` manually via URL (Verify redirect/403).
- [ ] **Account States**:
  - [ ] Admin: Disable Regular User.
  - [ ] Verify Regular User is immediately kicked (or fails next request).
  - [ ] Verify Regular User cannot log in.

## 3. Connection Matrix: SQL Databases
*Goal: Test all supported SQL drivers and proxying modes.*

- [ ] **Local Connections (Server-Side)**:
  - [ ] **SQLite**: Driver: `sqlite`, DSN: Path to `chronix.db`. Test Connection.
  - [ ] **PostgreSQL**: Driver: `postgres`, DSN: `postgres://user:pass@localhost:5432/db`. Test Connection.
  - [ ] **MySQL/MariaDB**: Driver: `mysql`, DSN: `user:pass@tcp(localhost:3306)/db`. Test Connection.
  - [ ] **MS SQL Server (T-SQL)**: Driver: `mssql` or `sqlserver`, DSN: `sqlserver://user:pass@localhost:1433?database=db`. Test.
  - [ ] **Oracle**: Driver: `oracle`, DSN: `oracle://user:pass@localhost:1521/xe`. Test.
  - [ ] **Snowflake**: Driver: `snowflake`, DSN: `account.region/db/schema?warehouse=wh&role=rl&user=usr&password=pwd`. Test.
- [ ] **Agent-Proxied Connections (Remote execution)**:
  - [ ] **SQLite via Agent**: Mode: `agent`, Driver: `sqlite`, Agent: [LinuxAgent]. Test.
  - [ ] **PostgreSQL via Agent**: Mode: `agent`, Driver: `postgres`, Agent: [LinuxAgent]. Test.
  - [ ] **MySQL/MariaDB via Agent**: Mode: `agent`, Driver: `mysql`, Agent: [LinuxAgent]. Test.
  - [ ] **MS SQL Server via Agent**: Mode: `agent`, Driver: `mssql`, Agent: [LinuxAgent]. Test.
- [ ] **Connection Duplication**:
  - [ ] Select an existing database connection and click **Duplicate**.
  - [ ] Verify name is "Copy Of ...".
  - [ ] Verify navigation to the edit page of the new connection.
  - [ ] Verify sidebar/menu updates with the new connection.
- [ ] **Health Monitoring**:
  - [ ] Set a database connection to "Auto-check" every 30 seconds.
  - [ ] Stop the remote database.
  - [ ] Verify connection status turns Red in UI and a Notification is created.
  - [ ] Start database. Verify recovery to Green.

## 4. Connection Matrix: Shell & Scripts
*Goal: Test all transport modes, operating systems, and privilege escalation.*

- [ ] **Localhost (Server)**:
  - [ ] Mode: `localhost`. Test Connection (returns server hostname).
- [ ] **SSH Connections**:
  - [ ] **Linux (Password)**: Mode: `ssh`, Host/Port/User/Password. Test.
  - [ ] **Linux (Key Generation - OpenSSH)**: Mode: `ssh`, Click "Generate Key Pair", select "OpenSSH". Verify private key populated and public key displayed. Copy public key and deploy to target. Test.
  - [ ] **Linux (Key Generation - PEM)**: Mode: `ssh`, Click "Generate Key Pair", select "PEM". Verify private key populated with `BEGIN PRIVATE KEY`. Test.
  - [ ] **Linux (Key)**: Mode: `ssh`, Upload Private Key (no passphrase). Test.
  - [ ] **Linux (Key + Passphrase)**: Mode: `ssh`, Upload Key + provide Passphrase. Test.
  - [ ] **macOS (SSH)**: Mode: `ssh`, Target macOS host. Test.
  - [ ] **Windows (SSH)**: Mode: `ssh`, Target Windows host with OpenSSH. Test.
- [ ] **Agent Connections (Direct)**:
  - [ ] **Agent - Linux**: Mode: `agent`, Select approved Linux agent. Test.
  - [ ] **Agent - macOS**: Mode: `agent`, Select approved macOS agent. Test.
  - [ ] **Agent - Windows**: Mode: `agent`, Select approved Windows agent. Test.
- [ ] **Connection Duplication**:
  - [ ] Select an existing shell connection and click **Duplicate**.
  - [ ] Verify name is "Copy Of ...".
  - [ ] Verify navigation to the edit page.
- [ ] **Privilege Escalation**:
  - [ ] **Sudo (Localhost)**: Connection set to use sudo. Run `whoami`. Result should be `root`.
  - [ ] **Sudo (Agent)**: Connection set to use sudo. Run `whoami`. Result should be `root`.
  - [ ] **Run As User**: Connection set to `run_as = nobody`. Run `whoami`. Result should be `nobody`.

## 5. Connection Matrix: Web Tasks (API)
*Goal: Test all authentication schemes and proxying.*

- [ ] **Authentication Modes**:
  - [ ] **None**: Test against `https://httpbin.org/get`.
  - [ ] **Basic Auth**: Provide User/Pass. Test against `https://httpbin.org/basic-auth/user/pass`.
  - [ ] **Bearer Token**: Provide JWT. Test against `https://httpbin.org/bearer`.
  - [ ] **Custom Header**: Name: `X-API-Key`, Value: `secret`. Test against `https://httpbin.org/headers`.
- [ ] **Transport Modes**:
  - [ ] **Local Web Task**: Run directly from server.
  - [ ] **Agent-Proxied Web Task**: Run through an agent. Verify it can reach an internal-only URL.
- [ ] **Connection Duplication**:
  - [ ] Select an existing web task connection and click **Duplicate**.
  - [ ] Verify name is "Copy Of ...".
  - [ ] Verify navigation to the edit page.

## 6. Action Sandbox & Assertion Engine
*Goal: Exercise the multi-step logic and all result validation types.*

- [ ] **SQL Action - Assertions**:
  - [ ] `rowExists`: SELECT count(*) FROM jobs.
  - [ ] `noRowsReturned`: Run a query that should return no discrepancy rows. Verify success on zero rows and failure when rows are returned.
  - [ ] `fieldEqualsFirst`: SELECT 'foo' as val. Assert `val == 'foo'`.
  - [ ] `fieldEqualsLast`: SELECT val FROM (SELECT 'a' as val UNION SELECT 'b' as val) ORDER BY val. Assert `val == 'b'`.
  - [ ] `rowsAffected` (==): DELETE from temp table where 1=0. Assert `affected == 0`.
  - [ ] `rowsAffected` (>=): INSERT into temp table. Assert `affected >= 1`.
  - [ ] **Session Persistence**: Step 1 (CREATE TEMP TABLE test(id int)) -> Step 2 (INSERT INTO test VALUES(1)) -> Step 3 (SELECT * FROM test). Assert Success.
  - [ ] **Bind Arguments**: Run a query with `{{var}}`. Verify in Run Details that the actual value passed is shown in the "SQL Arguments" section.
  - [ ] Multi-step: Step 1 (CREATE TABLE) -> Step 2 (INSERT) -> Step 3 (SELECT).
- [ ] **Shell Action - Assertions**:
  - [ ] `contains`: echo "Chronix", assert contains "Chronix".
  - [ ] `notContains`: echo "Success", assert not contains "Error".
  - [ ] `exitCodeEquals`: `exit 0` (Success) vs `exit 42` (Assert 42).
  - [ ] `firstLineEquals`: echo -e "Line1\nLine2", assert "Line1".
  - [ ] `lastLineEquals`: echo -e "Line1\nLine2", assert "Line2".
  - [ ] `regexMatch`: echo "Version 1.2.3", assert `Version \d+\.\d+\.\d+`.
- [ ] **Web Task Action - Piping & Captures**:
  - [ ] **JSONPath Capture**: GET `/json`. Capture `v_title` from `$.slideshow.title`.
  - [ ] **Header Capture**: GET `/response-headers?X-Test=ABC`. Capture `v_header` from header `X-Test`.
  - [ ] **Regex Capture**: GET `/html`. Capture `v_match` from body `<h1>(.*?)</h1>`.
  - [ ] **Piping**: Use `{{v_title}}` in a subsequent Step's URL or Body.
- [ ] **Expectations**:
  - [ ] `statusCode`: Assert `200` == `200`, `404` != `200`, `latency` < `500ms`.
  - [ ] `bodyContains`: GET `/get`, assert contains `"url"`.

## 7. Job Scheduling & Execution Engine
*Goal: Verify the precision and resilience of the background runner.*

- [ ] **Schedule Types**:
  - [ ] **Interval**: Every 1 minute. Verify 3 consecutive runs.
  - [ ] **Human-Friendly**: "Every day at [CurrentTime + 2 mins]". Verify execution.
  - [ ] **Cron**: `*/2 * * * *`. Verify execution on even minutes.
- [ ] **Variable Overrides**:
  - [ ] Create Action with `echo {{target}}`.
  - [ ] Create Job. Set `target = "Production"`.
  - [ ] Run Job. Verify logs show "Production".
- [ ] **Concurrency Control**:
  - [ ] Create a Shell Action that runs `sleep 30`.
  - [ ] Run Job. While running, click "Run Now" again.
  - [ ] Verify the second dispatch is rejected because the same job is already queued or running.
- [ ] **Cancellation**:
  - [ ] Start a long-running job.
  - [ ] Click "Cancel" in the UI.
  - [ ] Verify process is killed and status is "Canceled".

## 8. The Agent System (Remote Workers)
*Goal: Verify security and cross-platform execution.*

- [ ] **Registration Flow**:
  - [ ] Register Agent from CLI.
  - [ ] Verify "Pending" status in UI.
  - [ ] Approve. Verify "Connected".
  - [ ] **OS User Tracking**:
    - [ ] Verify the agent's "Running As" user matches the local OS user in the **Agents List**.
    - [ ] Verify the "Running As" user is correctly displayed in the **Agent Detail** view.
- [ ] **Security (TOFU)**:
  - [ ] Stop Agent. Change Server SSL Cert (or delete `cert.pem`).
  - [ ] Restart Agent.
  - [ ] Verify Agent rejects connection due to "Pin mismatch" (Man-in-the-middle protection).
- [ ] **Cross-OS Execution**:
  - [ ] **Linux**: Run `ls -la`.
  - [ ] **macOS**: Run `sw_vers`. Verify detailed version (e.g. "Tahoe 26.2") is shown in UI.
  - [ ] **Windows**: Run `dir` (Verify correct shell detection/usage). Verify detailed version (e.g. "Microsoft Windows 11 Pro 23H2 (Build 22631)") is shown in UI.
- [ ] **Agent Update**:
  - [ ] Trigger "Update Agent" from UI.
  - [ ] Verify Agent downloads new binary, restarts, and reconnects with new version.
- [ ] **Update Notices**:
  - [ ] Put server or agent updates in `notify` mode instead of automatic.
  - [ ] Verify the Dashboard and **Settings > Overview** show an update notice with a link to **Settings > Updates**.
  - [ ] Dismiss the notice and verify it stays dismissed while navigating during the same browser session.

## 9. Observability, Reports & Alerts
*Goal: Verify the data remains visible and actionable.*

- [ ] **Real-time SSE**:
  - [ ] Open Job Run detail.
  - [ ] Trigger run.
  - [ ] Verify progress bar, logs, and steps update *live* without refresh.
  - [ ] **Run Now Auto-Dismiss**: Trigger "Run Now". Wait for completion. Verify the progress panel automatically closes after 15 seconds.
- [ ] **Reporting**:
  - [ ] Export Activity Log to **CSV**.
  - [ ] Export Activity Log to **HTML**.
  - [ ] Export Activity Log to **PDF**.
- [ ] **Notifications & Alerts**:
  - [ ] Configure SMTP. Trigger a failed job. Verify HTML Email received with status-appropriate color (Red).
  - [ ] Trigger a successful job with "Notify on Success". Verify HTML Email received with Green header.
  - [ ] Verify "Include Output" in Email: Email should contain actual SQL/Shell results in structured HTML tables or code blocks.
  - [ ] Configure Webhook. Trigger a job. Verify payload received.

## 10. Open Source Readiness
*Goal: Verify the public build has no edition gates, license-key workflows, or monetization copy.*

- [ ] **Feature Availability**:
  - [ ] Create multiple jobs, agents, connections, actions, and users without edition-limit errors.
  - [ ] Export Activity Log to CSV, HTML, and PDF without entitlement prompts.
  - [ ] Configure SMS, webhooks, and custom branding without upgrade prompts.
- [ ] **UI Audit**:
  - [ ] Verify Settings contains no Licensing page.
  - [ ] Verify no copy mentions paid tiers, license keys, subscriptions, or feature upgrades.
- [ ] **Public Snapshot**:
  - [ ] Verify local development artifacts, private keys, databases, and logs are ignored and absent from the public repository snapshot.

## 11. Infrastructure & Resilience
*Goal: Verify the system survives real-world failure scenarios.*

- [ ] **Panic Recovery**:
  - [ ] (Dev only) Trigger a mock panic in a job worker.
  - [ ] Verify the system recovers and logs the stack trace without crashing the whole server.
- [ ] **Process Restart**:
  - [ ] Kill the `chronix` process while a job is "Queued".
  - [ ] Restart.
  - [ ] Verify the job is picked up or marked "Aborted" correctly.
- [ ] **Database Locking**:
  - [ ] Stress test: Enqueue 50 jobs simultaneously.
  - [ ] Verify SQLite WAL handles the concurrency without "Database is locked" errors.

## 12. Lifecycle Management & OS Integration
*Goal: Verify native service behavior and unified control.*

- [ ] **Native Service Installation**:
  - [ ] Run `./chronix service install`. Verify success.
  - [ ] Verify service is running (e.g., `systemctl status chronix` on Linux).
  - [ ] **Agent Service Safety**:
    - [ ] Attempt to run `./chronix-agent service install` on a machine with *no* existing registration.
    - [ ] Verify it fails with a message instructing the user to run `register` first.
  - [ ] Run `./chronix-agent service install` after successful registration. Verify success.
  - [ ] Verify agent service is running.
- [ ] **CLI Lifecycle Commands**:
  - [ ] `./chronix status`: Verify it reports correct status.
  - [ ] `./chronix stop`: Verify process is terminated.
  - [ ] `./chronix restart`: Verify process re-starts.
  - [ ] Repeat for `chronix-agent`.
- [ ] **Agent Identity Management**:
  - [ ] Run `./chronix-agent info`. Verify name, UUID, server target, pin status, local agent state, and server probe output are shown.
  - [ ] Run `./chronix-agent repoint <server[:port]>`. Verify only the saved server target changes and the same identity is preserved.
  - [ ] Run `./chronix-agent repair-register`. Verify the existing UUID and keys are reused.
  - [ ] Run `./chronix-agent rename "<new-name>"`. Verify the local name changes and the server-side record updates.
  - [ ] Run `./chronix-agent reset`. Verify `ServerSPKIB64` is cleared in `config.json` (or the agent re-TOFU on next connect).
  - [ ] Run `./chronix-agent unregister`. Verify the server is notified when reachable and `config.json` is deleted even if the server is not.
- [ ] **Service Management Commands**:
  - [ ] `./chronix service status`: Verify it reports service status.
  - [ ] `./chronix service stop`: Verify service stops.
  - [ ] `./chronix service start`: Verify service starts.
  - [ ] Repeat for `chronix-agent`.
- [ ] **Instance Takeover**:
  - [ ] Start one instance of `chronix`.
  - [ ] Try to start another instance from a different terminal.
  - [ ] Verify "already running" prompt.
  - [ ] Select 'y' and verify the new instance takes over the ports and the old one exits.
- [ ] **Uninstallation**:
  - [ ] Run `./chronix service uninstall`.
  - [ ] Verify service is removed from the system.
  - [ ] Repeat for `chronix-agent service uninstall`.

## 13. Feedback & Bug Reporting
*Goal: Verify the end-to-end feedback loop and attachment handling.*

- [ ] **Submission Flow**:
  - [ ] Click "Feedback" in sidebar.
  - [ ] Submit a **Bug Report** with summary and description.
  - [ ] Submit a **Feature Request** with multi-file attachments.
  - [ ] Verify "Success" notification.
- [ ] **Admin Management**:
  - [ ] Navigate to **Settings > Feedback**.
  - [ ] Verify both submissions appear in the list.
  - [ ] Open a report and update its **Status** to "In Progress".
  - [ ] Add an **Admin Attachment** to a report.
  - [ ] Verify attachments can be downloaded and viewed.
- [ ] **Cleanup**:
  - [ ] Delete a report.
  - [ ] Verify report is removed from DB.
  - [ ] Verify associated files are removed from the `feedback/` data directory.
