# Chronix: Sovereign, High-Performance Task Automation

Chronix is a self-contained automation platform designed for engineers who demand total data sovereignty and operational simplicity. By eliminating external database dependencies and complex runtimes, Chronix provides a robust, enterprise-grade scheduler in a single, portable binary.



## 🚀 Key Capabilities

* **Integrated Persistence Engine**: All state and execution history live within a localized, high-concurrency internal engine—no external database (Postgres/MySQL) required.
* **Secure Agent Orchestration**: Manage tasks across remote infrastructure using encrypted WebSocket agents that phone home through firewalls.
* **Zero-Dependency Binary**: A single executable includes the web UI, the task engine, and the persistence layer.
* **Multi-Modal Tasking**: Native support for SQL assertions, Shell scripting, and Web/API hooks with sophisticated success criteria.
* **Sovereign Deployment**: Ideal for air-gapped environments, private clouds, and local workstations.

## 🛠 How it Works

Unlike traditional schedulers that require a "stack," Chronix is designed to be "unzip and run". 

1.  **Download**: Get the binary for your architecture.
2.  **Initialize**: Run `./chronix run` to bootstrap your secure admin account.
3.  **Automate**: Define your connections, actions, and recurring schedules through the intuitive Web UI.



## 📦 Distribution & Licensing

**Source Code**: The core source code for Chronix is currently private to protect our proprietary persistence and security architecture. 

**Release Binaries**: Compiled binaries for Linux, macOS, and Windows are distributed via our [Official Website](https://chronixhq.com/download).

**Licensing**: Chronix offers flexible licensing to suit different needs:
* **Subscription**: Continuous updates and major version upgrades for a flat monthly/yearly fee.
* **Perpetual**: Own a specific major version (v1.x) forever with a one-time purchase.
* **Free Tier**: Available for individuals and home-lab environments.

## 🤝 Community & Support

While the source is private, we are committed to a community-driven roadmap:
* **Issue Tracker**: Report bugs or suggest new task types/drivers in the [Issues](https://github.com/chronixhq/chronix/issues) tab.
* **Discussions**: Ask questions and share automation patterns in our [Discussions](https://github.com/chronixhq/chronix/discussions) area.
* **Private Beta**: We are currently in a closed beta phase. [Apply for Early Access here](https://chronixhq.com/beta).

---
[Website](https://chronixhq.com) | [Documentation](https://chronixhq.com/guides) | [Pricing](https://chronixhq.com/pricing)
