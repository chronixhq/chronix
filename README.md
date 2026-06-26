
# Chronix

**Zero-dependency task automation. Run SQL, Shell, and Web Tasks on time, every time.**

Chronix is a self-contained, high-performance task automation and scheduling platform. It is designed for engineers who need a powerful, reliable "set-and-forget" automation engine without the complexity of external databases, heavy runtimes, or multi-service deployments.

---

## ⚡️ Key Features

- **🚀 Single-Binary Power**: A zero-dependency Go binary. Copy it and run.
- **🛠 Multi-Step Orchestration**: Chain SQL queries, Shell scripts, and API calls into complex, resilient workflows.
- **🤖 Remote Agents**: Extend your reach into any network environment with lightweight, phone-home agents—no firewall inbound ports required.
- **🖥 Embedded Web UI**: Manage everything through a professional, real-time React dashboard with live SSE logs and progress tracking.
- **📅 Human-Friendly Scheduling**: Use standard Cron or natural language rules like "The last Friday of every month."
- **🔐 Production-Ready Security**: TOFU pinning for agents, encrypted secret storage, and secure admin bootstrap.
- **📊 Deep Observability**: Unified activity timelines, structured logging, and instant Email/SMS alerting.

---

## Open Source

Chronix is 100% open source. The server, embedded web app, public website, and Chronix Crucible validation suite are published for everyone to inspect, run, modify, and improve.

- Licensed under the ISC License.
- No paid editions, license keys, subscriptions, or feature gates.
- Development happens in public through GitHub issues and pull requests.

---

## 🛠 Task Types

### 1. SQL Tasking
Automate your data infrastructure across **SQLite**, **PostgreSQL**, and **MySQL**. Assert on results, check rows affected, or verify specific field values.

### 2. Shell & Scripting
Run commands locally or via **SSH**. Full support for `sudo`, environment variables, and advanced regex-based output verification.

### 3. Web Tasks (HTTP)
The glue for your microservices. Execute API calls, capture response data via JSONPath, and pipe it into subsequent steps.

---

## 🤖 The Chronix Agent

Chronix Agents allow you to execute tasks on remote servers or within private networks. They connect back to the Chronix server via an encrypted WebSocket, bypassing the need for complex firewall rules or VPNs.

---

## 🚀 Quick Start

### 1. Install and Run the Server
You can run Chronix directly, or install it as a native system service for your platform (Linux/macOS/Windows).

```bash
# Run directly (foreground)
./chronix

# Or install as a system service (recommended)
./chronix service install
./chronix service status
```

### 2. Bootstrap Admin
On the first run, Chronix will generate a one-time admin code. Navigate to the Web UI (default: `http://localhost:5170`) and use this code to create your admin account.

### 3. Schedule a Job
Once logged in, you can create a **Connection**, define an **Action** (one or more steps), and schedule a **Job** to run it.

---

## 📦 Installation

Download the latest binary for your platform using the following pattern:
`https://dist.chronixhq.com/latest/<platform>/<app>`

**Platforms:** `linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64`

**Example (Linux AMD64):**
- [Chronix Server](https://dist.chronixhq.com/latest/linux-amd64/chronix)
- [Chronix Agent](https://dist.chronixhq.com/latest/linux-amd64/chronix-agent)

---

## 🛠 CLI Commands

Chronix and Chronix Agent provide a unified set of commands for lifecycle management:

- `stop`: Gracefully stop the running instance.
- `status`: Check if the application is running.
- `restart`: Restart the application.
- `service`: Manage the native OS service (install, uninstall, start, stop, status).
- `version`: Print version and release notes.

---

## 🏗 Data Directory

Chronix stores all its state (SQLite database, keys, and logs) in a single directory.

**Standard (Root/System-wide):**
- **Linux**: `/var/lib/chronix`
- **Windows**: `C:\ProgramData\Chronix`

**User-specific (Non-root):**
- **Linux**: `~/.local/share/chronix` (follows XDG spec)
- **macOS**: `~/Library/Application Support/Chronix`

You can override this by setting the `CHRONIX_DATA_DIR` environment variable or using the `-D` flag.

---

## 🧪 Testing

To ensure stability and correctness, we provide both automated unit tests and a comprehensive manual testing checklist.

- **Automated Tests**: Run `go test ./...` to execute the backend test suite.
- **Local Release Gate**: Run `./dev/ci-local.sh` for the full local validation pass used before commits, pushes, and releases.
- **Manual Checklist**: Follow the [Release Testing Checklist](docs/testing_checklist.md) for a step-by-step verification of the entire system, including UI and Agent flows.
