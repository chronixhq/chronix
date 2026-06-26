---
title: Getting Started
description: Setup guide for the Chronix automation platform.
---

# Chronix Help Guide: Getting Started Step-by-Step

Welcome to Chronix! This guide will walk you through the process of setting up Chronix from scratch, configuring your first connections, actions, and jobs, and ensuring you are notified of their status.

---

## 1. Installation

Chronix is distributed as a single, self-contained binary with no external dependencies. Choose the method that fits your environment:

### Manual Download
1. Download the latest binary for your platform:
   - **Linux AMD64**: [Chronix](https://dist.chronixhq.com/latest/linux-amd64/chronix) | [Agent](https://dist.chronixhq.com/latest/linux-amd64/chronix-agent)
   - **Linux ARM64**: [Chronix](https://dist.chronixhq.com/latest/linux-arm64/chronix) | [Agent](https://dist.chronixhq.com/latest/linux-arm64/chronix-agent)
   - **macOS Apple Silicon**: [Chronix](https://dist.chronixhq.com/latest/darwin-arm64/chronix) | [Agent](https://dist.chronixhq.com/latest/darwin-arm64/chronix-agent)
   - **macOS Intel**: [Chronix](https://dist.chronixhq.com/latest/darwin-amd64/chronix) | [Agent](https://dist.chronixhq.com/latest/darwin-amd64/chronix-agent)
   - **Windows AMD64**: [Chronix](https://dist.chronixhq.com/latest/windows-amd64/chronix) | [Agent](https://dist.chronixhq.com/latest/windows-amd64/chronix-agent)
2. Move the binary to a directory in your PATH (e.g., `/usr/local/bin`).
3. Ensure it is executable by running: `chmod +x chronix`.

### One-Line Install (Linux, macOS, & Windows Bash)
You can install Chronix with a single command:

```bash
curl -fsSL https://chronixhq.com/install.sh | sh
```
This interactive script detects your platform, downloads the latest binary, installs it to your chosen directory, and guides you through the initial protocol and port configuration.

---

## 2. Initial Startup and Administration

### Starting Chronix
To start the Chronix server, navigate to the directory containing the `chronix` binary. You can run it directly in the foreground or install it as a system service:

**Option A: Foreground Execution**
```bash
./chronix run
```
You can optionally override network settings via CLI flags:
```bash
# Disable HTTPS and force HTTP to port 8080
./chronix run --disable-https --force-http-port=8080
```

**Option B: Service Installation (Recommended)**
```bash
# macOS & Linux (requires sudo)
sudo ./chronix service install
sudo ./chronix service status

# Windows
.\chronix.exe service install
.\chronix.exe service status
```
Chronix will initialize its data directory and start the embedded web server.

### Bootstrapping the Admin Account
On the first run, Chronix is in an "Uninitialized" state.
1.  **Check the Console**: Look for a message like:
    ```text
    System uninitialized
    Go to http://localhost:5170
    Admin login code: xxxx xxxx xxxx
    ```
2.  **Access the Web UI**: Open your browser and navigate to the provided URL (default is `http://localhost:5170`).
3.  **Enter the Admin Code**: Input the 12-character code displayed in your console.
4.  **Create Admin Profile**: Once authenticated with the code, you will be prompted to set up your primary admin account. Provide your name, email address, and a strong password. This will be your primary way to access the system going forward.

### Admin Password Recovery
Chronix is designed to be completely sovereign, which means it does not rely on external services for password resets. If you lose your admin password:
1.  **Access the CLI**: On the server where Chronix is running, execute:
    ```bash
    ./chronix adminCode
    ```
2.  **Get the Code**: The CLI will output a 12-character "Admin login code" that is valid for 10 minutes.
3.  **Login with Code**: Open the web UI and append `/settings` to the URL (e.g., `http://localhost:5170/settings`).
4.  **Enter the Code**: Input the code from the CLI to gain temporary admin access.
5.  **Reset Password**: Navigate to **Settings > Users**, select the account you need to recover, and update the password.

---

## 3. Notification Setup

Stay informed about your jobs by setting up Email or SMS notifications. Go to **Settings** in the sidebar.

### Email (SMTP) Configuration
Chronix can send alerts via any standard SMTP server.
- **SMTP Host & Port**: The address of your mail server (e.g., `smtp.gmail.com`) and port (usually `587` or `465`).
- **SMTP Login & Password**: Your authentication credentials.
- **SMTP Secure**: Choose between `SSL`, `TLS`, or `None`.
- **From Name & Email**: How the emails will appear in the recipient's inbox.
- **System Alert Emails**: A comma-separated list of emails that should receive all system-wide alerts.

### SMS (Twilio) Configuration
Chronix supports Twilio for SMS alerts.
- **Twilio Account SID & Token**: Found in your Twilio Console.
- **From Number**: Your Twilio SMS-enabled phone number.
- **System Alert Phones**: A comma-separated list of phone numbers (in E.164 format, e.g., `+1234567890`) to receive critical alerts.

### Webhook Configuration
For integration with external services (Slack, Discord, custom APIs), you can configure outgoing Webhooks.
- **Target URL**: The endpoint where the JSON payload will be sent.
- **HMAC Secret**: (Optional) For signing the payload, allowing the receiver to verify the source.
- **Events**: Subscribe to specific event categories like `job`, `connection`, or `system`.

---

## 4. Creating Connections

Connections define *where* your tasks will run. Navigate to **Connections** and click **New Connection**.

### Database Connections
Used for SQL Tasks (SQLite, PostgreSQL, MySQL, MariaDB).
- **Name**: A descriptive name (e.g., "Production Postgres").
- **Driver**: Select the database type.
- **DSN (Connection String)**: The standard connection string for your database.
- **Agent**: (Optional) Select a Chronix Agent if the database is in a private network. See [Chronix Agents](#7-chronix-agents) for setup instructions.
- **Duplication**: You can duplicate an existing connection to quickly create a new one with similar settings.

### Shell Connections
Used for Shell and Scripting Tasks.
- **Mode**:
    - `Localhost`: Run directly on the Chronix server.
    - `SSH`: Run on a remote host via SSH.
- **Host & Port**: (For SSH) The remote address and port.
- **Authentication**: (For SSH) Support for Password or Private Key.
- **SSH Key Generation**: You can generate a new Ed25519 key pair directly from the UI. Choose between **OpenSSH (Recommended)** and **PEM (PKCS#8)** formats. After generation, the public key is provided for easy deployment to your target server's `authorized_keys` file.
- **Sudo**: Enable this if your commands require root privileges. You can provide a sudo password for secure injection.
- **Common Shells**: When defining steps in a Shell Action, you can select from a list of common shells (e.g., `/bin/bash` for Linux/macOS or absolute paths for Windows like `C:\Windows\System32\cmd.exe`). Using absolute paths on Windows ensures reliable execution even if the `PATH` environment variable is limited.

### Web Task Connections
Used for HTTP/API Tasks.
- **Base URL**: The root URL of the API (e.g., `https://api.example.com`).
- **Authentication**: Supports `None`, `Basic Auth`, `Bearer Token`, or `Custom Header`.

---

## 5. Creating Actions

Actions define *what* will be executed. An Action consists of one or more **Steps**. Navigate to **Actions** and click **New Action**.

### Defining Steps
- **Name**: A name for the specific step.
- **SQL/Command/URL**: The actual logic to execute (SQL query, shell command, or API endpoint).
- **Timeout**: How long Chronix should wait for this step to complete before failing.

### Success Criteria (Expectations)
For each step, you can define what constitutes a "success":
- **SQL**: Check rows affected, require at least one row, require zero rows, or verify a specific value in the first or last result row.
- **Shell**: Check the exit code (usually `0` for success) or look for specific text in the output (stdout/stderr).
- **Web**: Check the HTTP status code (e.g., `200`) or validate the JSON response using JSONPath.

### Failure Policy
- **Exit on Failure**: Stop the entire Action if this step fails (default).
- **Continue on Failure**: Proceed to the next step even if this one fails.

---

## 6. Creating Jobs

Jobs are the final piece that ties everything together. They define *when* an Action runs on a specific Connection. Navigate to **Jobs** and click **New Job**.

### Configuration
1.  **General Info**: Name and optional description.
2.  **Target Connection**: Choose the Connection you created earlier.
3.  **Action**: Choose the Action to be executed.
4.  **Schedule**:
    - **Single Shot**: Run once at a specific date and time.
    - **Recurring**: Use the human-friendly builder (e.g., "Every hour", "Daily at 3:00 PM") or a standard **Cron** expression.
    - **Overlap Protection**: Chronix will not allow the same job to be queued or started again while it already has a queued or running execution.
5.  **Variables**: If your Action uses placeholders like `{{customer_id}}`, you can define their values here. This allows you to reuse the same Action for different Jobs.
6.  **Notifications & Reporting**: 
    - Toggle whether you want to receive alerts for this specific job's **Success** or **Failure**.
    - **Include Output**: Enable this to have the actual task results (SQL counts, script logs, etc.) sent directly in the notification. This is perfect for daily reports or actionable debugging.

Once saved, your Job will appear in the **Dashboard** and execute automatically according to its schedule!

---

## 7. Chronix Agents

Chronix Agents are lightweight binaries that allow you to execute tasks on remote servers or within private networks. They connect back to the main Chronix server via an encrypted WebSocket, bypassing the need for complex firewall rules or VPNs.

### Setup and Registration
To set up a new agent:
1.  **Download the Agent Binary**: Use the platform-specific URL from the [Installation](#1-installation) section.
2.  **Register the Agent**: Run the following command on the target machine:
    ```bash
    ./chronix-agent register <chronix-server-ip> <agent-name>
    ```
    *Replace `<chronix-server-ip>` with the IP or hostname of your Chronix server, and `<agent-name>` with a unique name for this agent.*
3.  **Approve in the Web UI**: Log in to the Chronix web UI. You will see a notification for a pending agent registration. Approve it to complete the setup.
4.  **Start the Agent**: You can start the agent directly or install it as a service. **Note: The agent must be successfully registered before it can be installed as a system service.**

    **Option A: Service Installation (Recommended)**
    ```bash
    ./chronix-agent service install
    ./chronix-agent service status
    ```

    **Option B: Foreground Execution**
    ```bash
    ./chronix-agent
    ```

### Agent Monitoring
Once connected, you can monitor your agents from the **Agents** page in the Web UI.
-   **Status & Connectivity**: See real-time online/offline status and last seen timestamps.
-   **Running User**: Chronix identifies and displays the OS user the agent is currently running as (e.g., `root`, `administrator`, or a specific service account).
-   **Metadata**: View technical details about the agent's host, including architecture, CPU count, and detailed platform information such as distribution (Linux), marketing name (macOS), and build numbers (Windows). This information is automatically refreshed every time the agent connects.

### Agent Management & Lifecycle
The `chronix-agent` binary supports several commands for managing its lifecycle:

-   **Inspect Current Registration**:
    ```bash
    ./chronix-agent info
    ```
    Shows the saved agent name, UUID, server target, config file path, server pin status, local running state, and a quick server probe.

-   **Stop/Status/Restart**:
    ```bash
    ./chronix-agent status
    ./chronix-agent stop
    ./chronix-agent restart
    ```
-   **Service Management**:
    ```bash
    ./chronix-agent service status
    ./chronix-agent service stop
    ./chronix-agent service start
    ```
-   **Reset Server Pin**: If the server's certificate changes, use `reset` to trust the new certificate (TOFU):
    ```bash
    ./chronix-agent reset
    ```
-   **Repair Server Address or Registration**: If the saved server address changes or the server loses the agent record, refresh the existing identity instead of creating a new one:
    ```bash
    ./chronix-agent repoint <server[:port]>
    ./chronix-agent repair-register
    ./chronix-agent repair-register --server <server[:port]>
    ./chronix-agent repair-register --name "<new-name>"
    ./chronix-agent rename "<new-name>"
    ```
    `repoint` changes only the saved server target. `repair-register` and `rename` keep the same UUID and keys while refreshing the server-side agent record.
-   **Unregister Agent**: To completely remove an agent's identity:
    ```bash
    ./chronix-agent unregister
    ```
    `unregister` makes a best-effort attempt to notify the server first, but it still removes the local registration even if the server cannot be reached.

### Managing Agents from the Server CLI
You can also manage connected agents directly from the Chronix server:

-   **List Agents**: 
    ```bash
    ./chronix status    # Check server status first
    ./chronix agents list
    ```
-   **Update Agents**:
    ```bash
    ./chronix agents update all           # Update all agents to latest
    ./chronix agents update <uuid>        # Update specific agent to latest
    ```

---

## 8. Maintenance and Updates

Chronix includes a built-in update mechanism to keep your server and agents running the latest versions with the latest features and security fixes.

### Server Updates
You can check for and apply updates to the Chronix server using the CLI:

-   **Check for Updates**:
    ```bash
    ./chronix update check
    ```
-   **Apply Update**:
    ```bash
    ./chronix update apply
    ```
    This command downloads the latest binary, replaces the current one, and restarts the Chronix daemon if it is running.

### Automatic Updates
Updates can be fully automated via the Web UI under **Settings > Updates**:
-   **Enabled**: Toggle automatic checks and updates.
-   **Update Window**: Define a preferred time window (e.g., `02:00`) when updates should be applied to minimize disruption.
-   **Agent Updates**: Choose whether agents should also be updated automatically when a new version is released.
-   **Update Notices**: If updates are available and automatic mode is not enabled, Chronix shows a notice on the Dashboard and in **Settings > Overview** with a direct link back to **Settings > Updates**.
-   **Bulk Agent Updates**: When multiple agents have updates available, **Settings > Updates** includes an **Update All** button that starts updates for all online outdated agents and skips any offline agents.

---

## 9. Activity, Reporting, and Branding

Chronix provides comprehensive observability into your automated environment.

### Activity Timeline
The **Activity** tab shows a unified history of job runs and administrative actions.
- **Server-Side Filtering**: Search by action, user, or date range across the entire history.
- **Pagination**: Navigate large datasets efficiently.

### Exporting Reports
You can export the activity log for compliance or internal reporting.
- Click the **Export** button on the Activity page.
- Choose between **CSV**, **HTML**, or **PDF** formats.

### Custom Branding
You can customize the Chronix interface for your team or environment.
- Navigate to **Settings > Branding**.
- Provide a **Custom Logo URL** and **Brand Name**.
- These will replace the Chronix identity in the top navigation bar.

### User Interface & Navigation
Chronix features a modern React UI with real-time updates.
- **Context-Sensitive Help**: Look for the `?` icon (SectionHelp) next to page titles. Clicking it will open the relevant section of this help guide.
- **Real-Time Logs**: Watch job runs execute live via Server-Sent Events (SSE).

---

## 10. Feedback and Bug Reporting

Chronix includes an integrated feedback system to help us improve the platform.

### Submitting Feedback
1. Click the **Feedback** icon in the sidebar.
2. Choose between **Bug Report** or **Feature Request**.
3. Provide a clear **Summary** and detailed **Description**.
4. **Attachments**: You can upload multiple files (screenshots, logs, etc.) to provide additional context.

### Managing Feedback (Admins Only)
Admins can view and manage all submitted feedback under **Settings > Feedback**.
- **Status Tracking**: Update the status of reports (e.g., Open, In Progress, Resolved, Closed).
- **Admin Attachments**: Admins can add their own attachments to a report for internal tracking.
- **Deletion**: Reports and their associated files can be permanently removed when no longer needed.

---

## 11. Open Source

Chronix is published as a fully open-source project. There are no paid editions, entitlement checks, or feature gates.

### What Is Included
- The Chronix application source.
- The public Chronix website source.
- Chronix Crucible, the release-validation and testing companion project.

### Contributing
Use the public GitHub issue tracker for bugs and feature requests. Pull requests should include a short explanation, relevant tests, and any documentation updates needed for user-visible behavior.
