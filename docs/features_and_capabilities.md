---
title: Features and Capabilities
description: Chronix features, execution models, and operational capabilities.
---

# Chronix Features and Capabilities (Marketing Reference)

This document is a living record of Chronix's features, architectural advantages, and unique selling points. It serves as the primary source of truth for generating marketing materials, documentation, and the main website content.

---

## 🚀 Core Philosophy: "Zero to Automation in Seconds"
Chronix is a self-contained, high-performance task automation and scheduling platform designed for modern engineering teams who value simplicity without sacrificing power.

- **One-Line Interactive Install**: Deploy the server in seconds with `curl -fsSL https://chronixhq.com/install.sh | sh`. The script guides you through platform detection and network configuration.
- **Zero-Dependency Execution**: A single binary (Server or Agent) is all you need. No external databases, runtimes, or complex dependencies.
- **SQLite-Powered Sovereignty**: All state, history, and settings live in a localized, portable SQLite database with high-performance WAL (Write-Ahead Logging) enabled.
- **Contained Application Model**: Every configuration, credential, and certificate lives in a user-defined data directory, making backups and migrations as simple as copying a folder.

---

## 🛠 The "Big Three" Task Types
Chronix is the Swiss Army Knife of automation, supporting three primary task pillars:

### 1. SQL Tasking
- **Engine Agnostic**: Native support for **SQLite**, **PostgreSQL**, **MySQL**, **MariaDB**, **MS SQL Server (T-SQL)**, **Oracle**, and **Snowflake**.
- **Dialect Selection**: Choose between `generic`, `sqlite`, `postgres`, `mysql`, and `tsql` to enable dialect-specific syntax highlighting and validation in the editor.
- **Multi-Step Orchestration**: Chain multiple SQL queries into a single action with granular control over failure policies (Exit vs. Continue).
- **Session-Sticky Execution**: Multi-step actions use a single database connection, enabling reliable use of `last_insert_rowid()`, temporary tables, and session variables across steps.
- **One-Click Duplication**: Rapidly scale your connection library by duplicating existing Database, Shell, or Web Task connections.
- **Dynamic Template Variables**: Inject runtime data into queries and expectations using `${var}` or `{{var}}` syntax.
- **Deep Assertion Engine**: Verify results with built-in assertions for rows affected, row existence, zero-row expectations, or specific field values (e.g., "Verify status field is 'ACTIVE'").
- **Zero-Row Guardrails**: Mark a SQL step as successful only when it returns zero rows, which is ideal for discrepancy scans, drift checks, and "report only when something is wrong" queries.
- **Execution Traceability**: Review exact snapshots of substituted SQL for any historical run, including the **bind arguments** used for parameterized queries.
- **Result Data Visibility**: View query results directly in the UI with structured tables for successful database actions.

### 2. Shell & Scripting
- **Local & Remote Power**: Execute shell commands or full scripts locally or via **SSH** outbound to remote hosts.
- **Advanced Result Assertions**: Beyond exit codes, verify success via regex patterns, line-by-line equality, or output containment.
- **Integrated SSH Key Management**: Generate modern Ed25519 key pairs (OpenSSH or PEM formats) directly from the UI, featuring one-click copy-to-clipboard for rapid public key deployment.
- **Sudo-Aware**: Seamlessly execute privileged commands with secure password injection.
- **Environment & Working Dir Control**: Full control over environment variables and execution paths.
- **Intelligent Log Truncation**: Configurable Head/Tail policies to capture critical output without bloating the database.

### 3. Web Tasks (HTTP & API)
- **First-Class HTTP Support**: Native execution of `GET`, `POST`, `PUT`, `DELETE`, and `PATCH` requests.
- **Integrated Authentication**: Built-in support for Basic Auth, Bearer Tokens (JWT), and custom API Key headers.
- **Response Capture & Variable Piping**: Extract response data (via JSONPath or Regex) and pipe it into subsequent steps, creating a powerful "glue" between disparate APIs.
- **Agent-Proxied API Calls**: Execute HTTP tasks from a **Chronix Agent**, allowing you to automate private, internal APIs without exposing them to the internet.
- **Performance Verification**: Assert on response status codes and latency (e.g., "Ensure API responds within 500ms").

---

## 🤖 The Chronix Agent System
The Agent is a lightweight, remote execution worker that allows you to extend Chronix's reach into any network environment without firewall pain.

- **MITM Protection (TOFU Pinning)**: Security you can trust. Agents automatically pin the server's public key fingerprint on the first connection (**Trust On First Use**), preventing Man-In-The-Middle attacks even in self-signed environments.
- **Encrypted WebSockets**: All command and control traffic is tunneled through a TLS-encrypted WebSocket.
- **Ed25519-Signed Identity**: Every agent authenticates with a unique, cryptographically signed JWT. No shared secrets.
- **OS User Visibility**: Agents report the local OS user they are running as, providing essential context for troubleshooting and security audits.
- **Rich Platform Visibility**: Agents report detailed OS information including distribution (Linux), version, and architecture (e.g., "Microsoft Windows 11 Pro 23H2 (Build 22631)", "Tahoe 26.2 (arm64)", "Sequoia 15.1 (arm64)"). This information is refreshed on every connection to ensure accuracy.
- **Safe Service Installation**: Built-in safeguards ensure agents are properly registered with the server before allowing installation as a system service, preventing non-functional deployments.
- **Firewall Friendly**: Agents "phone home" to the server, meaning no inbound ports need to be opened on remote systems.
- **Full Feature Parity**: Anything the server can do locally (SQL, Shell, Web Tasks), the Agent can do from its local network position.

---

## 📅 Precision Scheduling Engine
- **Human-Friendly Recurring Rules**: Support for complex patterns like "The last Friday of every month" or "Every Tuesday at 2:00 PM."
- **Standard Cron**: Full 5-field cron support for traditional power users.
- **Single-Shot Execution**: Schedule one-time tasks for the future with to-the-minute precision.
- **Active Concurrency Control**: Prevents overlapping runs of the same job. If a trigger fires while that job is already queued or running, Chronix rejects the duplicate dispatch and surfaces the reason in logs and alerts.
- **Lifecycle Windows**: Define precisely when a recurring schedule starts and ends.

---

## 🖥 User Interface & Experience
- **Embedded Web UI**: A professional, responsive React dashboard served directly from the Chronix binary.
- **Real-Time SSE Updates**: Watch your tasks execute in real-time. Logs, progress bars, and status chips update instantly via Server-Sent Events.
- **Run Now Progress Panel**: Trigger manual runs and monitor them in a dedicated progress card. The card automatically dismisses 15 seconds after completion, allowing for review without clutter.
- **Update Awareness**: When newer Chronix or Agent builds are available and updates are not fully automatic, the Dashboard and Settings Overview surface dismissible notices that deep-link straight to **Settings > Updates**.
- **Modern Material UI Design**: A clean, modern interface built on the current Material UI stack and tuned for real-world operations work.
- **Context-Sensitive Help**: Use the integrated `SectionHelp` icons throughout the interface to jump directly to the relevant part of the documentation for each configuration page.
- **Sandbox Action Testing**: Validate your logic in real-time with an interactive test dialog. Select target connections and provide mock variables to see immediate results.
- **Connection Health Monitoring**: Automatic background checks ensure your database and shell targets are online.

---

## 🔐 Production-Ready Security
- **Secure Bootstrap Flow**: Admin code-based initialization prevents unauthorized first-time access.
- **Encrypted Secret Storage**: All database and SSH credentials are encrypted at rest using a master key unique to your installation.
- **Audit Logging**: Comprehensive user activity logging records every change made to the system.
- **Certificate Autopilot**: Automatically generates and manages self-signed TLS certificates, or allows for custom PEM injection.

---

## 📊 Observability & Operations
- **Unified Activity Timeline**: A single view showing all system events, job results, and user actions with server-side pagination and advanced filtering.
- **Activity Report Exporting**: Generate professional activity reports in **CSV**, **HTML**, and **PDF** formats, filtered by date, user, or action type.
- **Proactive Alerting**: Instant notifications via **Email (SMTP)**, **SMS (Twilio)**, and **Outgoing Webhooks** for job failures or system warnings.
- **Professional Notifications**: Emails are delivered in a modern, status-colored HTML format with a plain-text fallback. Includes structured detail tables and formatted result snapshots.
- **Custom Branding (Whitelabeling)**: Professionalize the interface with your own logo and brand name when running Chronix for your team or environment.
- **Automated Reporting (Job Output)**: Go beyond pass/fail alerts. Configure jobs to include their actual output (SQL result counts, Shell logs, or API responses) directly in the email or SMS notification, turning Chronix into a proactive data delivery engine.
- **Structured Logging**: `slog`-based structured logging for easy ingestion into log management platforms.
- **Detailed Run History**: Full capture of every run's inputs, outputs, logs, and performance metrics.
- **Automated Log Retention**: Keep your database lean with configurable retention policies that automatically purge historical data.
- **Integrated Feedback System**: Users can submit bug reports and feature requests directly from the dashboard, including multi-file attachments. Admins can manage these submissions, track status, and add administrative attachments.
- **Comprehensive Lifecycle Management**: Powerful CLI and RPC control for stopping, restarting, and checking status.
- **Native Service Integration**: Install Chronix and Chronix Agent as native system services (systemd, LaunchAgent, Windows Service) with simple `service` commands.
- **Unix Socket RPC**: Powerful local CLI control for integration with other server-side tools.

---

## Open Source Availability
Chronix is 100% open source. The app, website, and Chronix Crucible validation suite are published together so users can inspect, run, modify, and contribute to the full system.

- **No paid editions**: All features are available in the public source tree.
- **No license keys**: Chronix does not phone home for entitlement checks or gate features behind signed keys.
- **Community-first development**: Issues, pull requests, release validation, and documentation live in the open.
