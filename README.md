# Chronix: Sovereign, High-Performance Task Automation

<p align="center">
  <img src="https://chronixhq.com/Chronix-white.png" width="400" alt="Chronix Logo">
  <br>
  <b>Zero-Dependency Automation. Total Sovereignty.</b>
</p>

---

Chronix is a self-contained automation platform designed for engineers who demand total data sovereignty and operational simplicity. By eliminating external database dependencies and complex runtimes, Chronix provides a robust, enterprise-grade scheduler in a single, portable binary.

## 🚀 Key Capabilities

*   **Integrated Persistence Engine**: All state and execution history live within a localized, high-concurrency internal engine—no external database (Postgres/MySQL) required.
*   **Secure Agent Orchestration**: Manage tasks across remote infrastructure using encrypted WebSocket agents that "phone home" through firewalls via Ed25519-signed JWTs.
*   **Multi-Modal Tasking**: Native support for **SQL assertions** (SQLite, PG, MySQL), **Shell scripting** (local/SSH), and **Web/API hooks** with JSONPath response piping.
*   **Human-Friendly Scheduling**: Define rules like "Last Friday of the month" or standard 5-field Cron with active concurrency control.
*   **Real-Time Observability**: Watch execution live via SSE (Server-Sent Events) and a unified activity timeline.
*   **Sovereign Deployment**: Ideal for air-gapped environments, private clouds, and local workstations. One binary, zero dependencies.

## 🛠 How it Works

Unlike traditional schedulers that require a complex "stack," Chronix is designed to be "unzip and run".

1.  **Download**: Get the binary for your architecture (Linux, macOS, Windows).
2.  **Initialize**: Run `./chronix run` to bootstrap your secure admin account and start the Web UI.
3.  **Automate**: Define your connections, actions, and recurring schedules through the intuitive React-based dashboard.

## 🖥 Visual Dashboard

<p align="center">
  <img src="https://chronixhq.com/screenshots/dashboard.png?i=2" width="800" alt="Chronix Dashboard">
  <br>
  <i>The Chronix Dashboard: Real-time observability and task orchestration.</i>
</p>

<p align="center">
  <img src="https://chronixhq.com/screenshots/task-editor.png" width="800" alt="Chronix Task Editor">
  <br>
  <i>Visual Task Editor: Configure SQL, Shell, and Web tasks with response piping.</i>
</p>

<p align="center">
  <img src="https://chronixhq.com/screenshots/connections.png" width="800" alt="Chronix Connections">
  <br>
  <i>Unified Connectivity: Multi-protocol support with real-time health monitoring.</i>
</p>

## 📦 Distribution & Licensing

**Source Code**: The core source code for Chronix is currently private to protect our proprietary persistence and security architecture.

**Release Binaries**: Compiled binaries for Linux, macOS, and Windows are distributed via our [Official Website](https://chronixhq.com/download).

**Licensing**: Chronix offers flexible licensing to suit different needs:
*   **Free Tier**: 1 Server, 1 Agent, 5 Jobs. Ideal for individuals and home-labs.
*   **Subscription**: Continuous updates and major version upgrades (Pro/Enterprise).
*   **Perpetual**: Own a specific major version (v1.x) forever with a one-time purchase—perfect for air-gapped systems.

## 🤝 Community & Support

While the source is private, we are committed to a community-driven roadmap:
*   **Issue Tracker**: Report bugs or suggest new task types/drivers in the [Issues](https://github.com/chronixhq/chronix/issues) tab.
*   **Discussions**: Ask questions and share automation patterns in our [Discussions](https://github.com/chronixhq/chronix/discussions) area.
*   **Private Beta**: We are currently in a closed beta phase. [Apply for Early Access here](https://chronixhq.com/beta).

---
[Website](https://chronixhq.com) | [Documentation](https://chronixhq.com/guides/getting-started) | [Pricing](https://chronixhq.com/pricing) | [Contact Support](mailto:support@chronixhq.com)
