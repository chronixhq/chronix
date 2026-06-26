/*
 Navicat Premium Data Transfer

 Source Server         : Chronix Schema
 Source Server Type    : SQLite
 Source Server Version : 3035005 (3.35.5)
 Source Schema         : main

 Target Server Type    : SQLite
 Target Server Version : 3035005 (3.35.5)
 File Encoding         : 65001

 Date: 13/09/2025 14:32:47
*/

PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for auth_keys
-- ----------------------------
DROP TABLE IF EXISTS "auth_keys";
CREATE TABLE auth_keys (
    auth_key TEXT
);

-- ----------------------------
-- Table structure for cx_settings
-- ----------------------------
DROP TABLE IF EXISTS "cx_settings";
CREATE TABLE "cx_settings" (
    server_url         TEXT,
    smtp_host          TEXT,
    smtp_port          INTEGER,
    smtp_secure        TEXT,
    smtp_from_name     TEXT,
    smtp_from_email    TEXT,
    smtp_login         TEXT,
    smtp_password      TEXT,
    http_enabled       BOOL    CHECK (http_enabled IN (0,1)),
    http_port          INTEGER,
    https_mode         TEXT,
    https_port         INTEGER,
    https_enabled      BOOL    CHECK (https_enabled IN (0,1)),
    https_cert_pem     TEXT,
    https_key_pem      TEXT,
    sms_provider       TEXT,
    sms_from_number    TEXT,
    twilio_username    TEXT,
    twilio_password    TEXT,
    twilio_account_sid TEXT,
    twilio_api_key     TEXT,
    twilio_api_secret  TEXT,
    agent_enabled      BOOL    CHECK (agent_enabled IN (0,1)),
    agent_port         INTEGER,
    system_alert_emails TEXT,
    system_alert_phones TEXT,
    alert_on_agent_lost BOOL CHECK (alert_on_agent_lost IN (0,1)),
    updater_enabled    BOOL    CHECK (updater_enabled IN (0,1)),
    updater_mode       TEXT,
    updater_window_start TEXT,
    updater_agent_enabled BOOL    CHECK (updater_agent_enabled IN (0,1)),
    updater_agent_mode    TEXT,
    updater_agent_window_start TEXT,
    brand_logo_url     TEXT,
    brand_name         TEXT,
    id                 integer not null
        constraint cx_settings_pk
            primary key
);

-- ----------------------------
-- Table structure for cx_users
-- ----------------------------
DROP TABLE IF EXISTS "cx_users";
CREATE TABLE "cx_users" (
    "id"                    INTEGER NOT NULL,
    "email"                 TEXT    NOT NULL,
    "name"                  TEXT    NOT NULL,
    "phone"                 TEXT,
    "password"              TEXT,
    "force_password_change" BOOL    NOT NULL CHECK ("force_password_change" IN (0, 1)),
    "enabled"               BOOL    NOT NULL CHECK ("enabled" IN (0, 1)),
    "suspended"             BOOL    NOT NULL CHECK ("suspended" IN (0, 1)),
    "admin"                 BOOL    NOT NULL CHECK ("admin" IN (0, 1)),
    "sv"                    integer NOT NULL,
    "time_zone"             TEXT,
    "time_format"           TEXT,
    PRIMARY KEY ("id")
);

-- ----------------------------
-- Table structure for notification_recipients
-- ----------------------------
DROP TABLE IF EXISTS "notification_recipients";
CREATE TABLE notification_recipients (
    notification_id INTEGER   NOT NULL,
    user_id         INTEGER   NOT NULL,
    seen            BOOL      NOT NULL CHECK (seen IN (0, 1)),
    seen_at         TIMESTAMP,
    removed_at      TIMESTAMP,
    delivered_at    TIMESTAMP NOT NULL,
    PRIMARY KEY (notification_id, user_id),
    FOREIGN KEY (notification_id) REFERENCES notifications (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES cx_users (id) ON DELETE CASCADE
);

-- ----------------------------
-- Table structure for notifications
-- ----------------------------
DROP TABLE IF EXISTS "notifications";
CREATE TABLE "notifications" (
    "id"         INTEGER,
    "created_at" TIMESTAMP NOT NULL,
    "category"   TEXT      NOT NULL,
    "severity"   TEXT      NOT NULL,
    "subject"    TEXT      NOT NULL,
    "origin"     TEXT,
    "data"       JSONB,
    PRIMARY KEY ("id")
);

-- ----------------------------
-- Indexes structure for table notification_recipients
-- ----------------------------
CREATE INDEX "main"."idx_notification_recipients_user_seen"
    ON "notification_recipients" (
                                  "user_id" ASC,
                                  "seen" ASC,
                                  "delivered_at" ASC
        );

-- ----------------------------
-- Table structure for user_activity
-- ----------------------------
DROP TABLE IF EXISTS "user_activity";
CREATE TABLE "user_activity" (
    "id"         INTEGER,
    "user_id"    INTEGER   NOT NULL,
    "created_at" TIMESTAMP NOT NULL,
    "action"     TEXT      NOT NULL,
    "details"    TEXT,
    "ip"         TEXT,
    "user_agent" TEXT,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("user_id") REFERENCES cx_users ("id") ON DELETE CASCADE
);

-- ----------------------------
-- Indexes structure for table user_activity
-- ----------------------------
CREATE INDEX "main"."idx_user_activity_user_created"
    ON "user_activity" (
                        "user_id" ASC,
                        "created_at" DESC
        );
CREATE INDEX IF NOT EXISTS "idx_user_activity_created" ON "user_activity" ("created_at" DESC);

-- ----------------------------
-- Table structure for db_connections
-- ----------------------------
DROP TABLE IF EXISTS "db_connections";
CREATE TABLE "db_connections" (
    "id"                          INTEGER,
    "name"                        TEXT      NOT NULL,
    "driver"                      TEXT      NOT NULL CHECK ("driver" IN ('postgres','mysql','sqlite','mssql','oracle','snowflake')), -- postgres | mysql | sqlite | mssql | oracle | snowflake
    "dsn"                         TEXT      NOT NULL,
    "description"                 TEXT,
    "auto_check_enabled"          INTEGER   NOT NULL CHECK ("auto_check_enabled" IN (0, 1)),
    "auto_check_interval_seconds" INTEGER   NOT NULL CHECK ("auto_check_interval_seconds" >= 0),
    "agent_uuid"                  TEXT,
    "created_at"                  TIMESTAMP NOT NULL,
    "updated_at"                  TIMESTAMP NOT NULL,
    "last_status"                 TEXT,
    "last_error"                  TEXT,
    "last_checked_at"             TIMESTAMP,
    "alert_emails"                TEXT,
    "alert_phones"                TEXT,
    "notify_on_failure"           BOOL CHECK ("notify_on_failure" IN (0, 1)),
    "enabled" BOOL NOT NULL CHECK ("enabled" IN (0, 1)),
    "suspended" BOOL NOT NULL CHECK ("suspended" IN (0, 1)),
    PRIMARY KEY ("id"),
    FOREIGN KEY (agent_uuid) REFERENCES agents (uuid) ON DELETE SET NULL,
    CONSTRAINT "uq_db_connections_name" UNIQUE ("name")
);

-- Indexes for db_connections
CREATE INDEX IF NOT EXISTS "idx_db_connections_driver"
    ON "db_connections" ("driver" ASC);
CREATE INDEX IF NOT EXISTS "idx_db_connections_agent"
    ON "db_connections" ("agent_uuid" ASC);

-- ----------------------------
-- Table structure for actions
-- ----------------------------
DROP TABLE IF EXISTS "actions";
CREATE TABLE "actions" (
    "id"          INTEGER,
    "name"        TEXT      NOT NULL,
    "dialect"     TEXT      NOT NULL CHECK ("dialect" IN ('postgres','mysql','generic')), -- postgres | mysql | generic
    "description" TEXT,
    "notes"       TEXT,
    "action_type" TEXT      NOT NULL CHECK ("action_type" IN ('database','shell','webtask')),
    "enabled"     BOOL      NOT NULL CHECK ("enabled" IN (0, 1)),
    "suspended"   BOOL      NOT NULL CHECK ("suspended" IN (0, 1)),
    "created_at"  TIMESTAMP NOT NULL,
    "updated_at"  TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "uq_actions_name" UNIQUE ("name")
);

CREATE INDEX IF NOT EXISTS "idx_actions_type" ON "actions" ("action_type" ASC);

-- ----------------------------
-- Table structure for action_steps
-- ----------------------------
DROP TABLE IF EXISTS "action_steps";
CREATE TABLE "action_steps" (
    "id"              INTEGER,
    "action_id"       INTEGER NOT NULL,
    "step_order"      INTEGER NOT NULL CHECK ("step_order" >= 0),
    "name"            TEXT    NOT NULL,
    "sql_text"        TEXT    NOT NULL,
    "timeout_seconds" INTEGER,
    "expectation"     JSONB,
    "output_capture"   JSONB,
    "on_failure"      TEXT,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("action_id") REFERENCES actions (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS "idx_action_steps_action_order" ON "action_steps" ("action_id" ASC, "step_order" ASC);

-- ----------------------------
-- Table structure for jobs
-- ----------------------------
DROP TABLE IF EXISTS "jobs";
CREATE TABLE "jobs" (
    "id"                  INTEGER,
    "name"                TEXT      NOT NULL,
    "description"         TEXT,
    "notes"               TEXT,
    "connection_id"       INTEGER,
    "action_id"           INTEGER   NOT NULL,
    "target_kind"         TEXT      NOT NULL CHECK ("target_kind" IN ('database','shell','webtask')),
    "shell_connection_id" INTEGER,
    "webtask_connection_id" INTEGER,
    "schedule_json"       JSONB     NOT NULL, -- full schedule payload from UI (single/recurring/structured/cron)
    "enabled"             BOOL      NOT NULL CHECK ("enabled" IN (0, 1)),
    "suspended"           BOOL      NOT NULL CHECK ("suspended" IN (0, 1)),
    "alert_emails"        TEXT,
    "alert_phones"        TEXT,
    "notify_on_success"   BOOL CHECK ("notify_on_success" IN (0, 1)),
    "notify_on_error"     BOOL CHECK ("notify_on_error" IN (0, 1)),
    "notify_include_output" BOOL CHECK ("notify_include_output" IN (0, 1)),
    "created_at"          TIMESTAMP NOT NULL,
    "updated_at"          TIMESTAMP NOT NULL,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("connection_id") REFERENCES db_connections (id) ON DELETE RESTRICT,
    FOREIGN KEY ("action_id") REFERENCES actions (id) ON DELETE RESTRICT,
    FOREIGN KEY ("shell_connection_id") REFERENCES shell_connections (id) ON DELETE RESTRICT,
    FOREIGN KEY ("webtask_connection_id") REFERENCES webtask_connections (id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS "idx_jobs_conn" ON "jobs" ("connection_id" ASC);
CREATE INDEX IF NOT EXISTS "idx_jobs_action" ON "jobs" ("action_id" ASC);
CREATE INDEX IF NOT EXISTS "idx_jobs_target_kind" ON "jobs" ("target_kind" ASC);
CREATE INDEX IF NOT EXISTS "idx_jobs_shell_conn" ON "jobs" ("shell_connection_id" ASC);


-- ----------------------------
-- Table structure for job_variables
-- ----------------------------
DROP TABLE IF EXISTS "job_variables";
CREATE TABLE "job_variables" (
    "id"     INTEGER,
    "job_id" INTEGER NOT NULL,
    "name"   TEXT    NOT NULL,
    "value"  TEXT,
    PRIMARY KEY ("id"),
    FOREIGN KEY ("job_id") REFERENCES jobs (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS "uq_job_variables_job_name" ON "job_variables" ("job_id" ASC, "name" ASC);
CREATE INDEX IF NOT EXISTS "idx_job_variables_job" ON "job_variables" ("job_id" ASC);

-- ----------------------------
-- job_runs: one row per execution attempt of a job
-- ----------------------------
DROP TABLE IF EXISTS "job_runs";
CREATE TABLE "job_runs" (
    "id"                  INTEGER,            -- surrogate PK for joins
    "run_uid"             TEXT      NOT NULL, -- stable external ID (string used by current code/UI)
    "job_id"              INTEGER   NOT NULL, -- FK -> jobs.id

    -- denormalized snapshots for convenience (optional but useful for list views)
    "job_name"            TEXT,               -- snapshot of job name
    "connection_id"       INTEGER,            -- snapshot of target connection
    "action_id"           INTEGER,            -- snapshot of action
    "connection_kind"     TEXT CHECK ("connection_kind" IN ('database','shell','webtask') OR "connection_kind" IS NULL),
    "shell_connection_id" INTEGER,
    "webtask_connection_id" INTEGER,
    -- lifecycle timestamps
    "queued_at"           TIMESTAMP NOT NULL,
    "started_at"          TIMESTAMP,          -- null until running
    "finished_at"         TIMESTAMP,          -- null until finished

    -- status lifecycle: queued | running | success | error | cancelled
    "status"              TEXT      NOT NULL CHECK ("status" IN ('queued','running','success','error','cancelled')),

    -- who/what triggered
    "triggered_by"        INTEGER,            -- FK -> cx_users.id (nullable for system/cron)
    "trigger_source"      TEXT,               -- e.g., "cron", "manual", "api", etc.

    -- outcome summary & metrics
    "message"             TEXT,               -- short result/err summary for list view
    "rows_affected"       INTEGER,            -- aggregate across steps when relevant
    "error_code"          TEXT,               -- optional domain error code
    "error_details"       TEXT,               -- optional long error text

    -- freeform JSON for additional summary payloads (counts, sample values)
    "summary"             JSONB,

    PRIMARY KEY ("id"),
    CONSTRAINT "uq_job_runs_run_uid" UNIQUE ("run_uid"),
    FOREIGN KEY ("job_id") REFERENCES jobs (id) ON DELETE CASCADE,
    FOREIGN KEY ("triggered_by") REFERENCES cx_users (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS "idx_job_runs_job_time" ON "job_runs" ("job_id" ASC, "queued_at" DESC);
CREATE INDEX IF NOT EXISTS "idx_job_runs_status_time" ON "job_runs" ("status" ASC, "queued_at" DESC);
CREATE INDEX IF NOT EXISTS "idx_job_runs_started_time" ON "job_runs" ("started_at" DESC);
CREATE INDEX IF NOT EXISTS "idx_job_runs_shell_conn" ON "job_runs" ("shell_connection_id" ASC);
CREATE INDEX IF NOT EXISTS "idx_job_runs_webtask_conn" ON "job_runs" ("webtask_connection_id" ASC);

-- ----------------------------
-- job_run_steps: one row per action step execution within a run
-- ----------------------------
DROP TABLE IF EXISTS "job_run_steps";
CREATE TABLE "job_run_steps" (
    "id"              INTEGER,
    "run_id"          INTEGER NOT NULL, -- FK -> job_runs.id
    "step_order"      INTEGER NOT NULL CHECK ("step_order" >= 0), -- from action_steps.step_order
    "step_name"       TEXT    NOT NULL, -- snapshot of action_steps.name

    -- lifecycle
    "started_at"      TIMESTAMP,
    "finished_at"     TIMESTAMP,
    "status"          TEXT    NOT NULL CHECK ("status" IN ('queued','running','success','error','skipped')), -- queued | running | success | error | skipped

    -- SQL execution context
    "sql_text"        TEXT,             -- snapshot of executed SQL (optional; can truncate if large)
    "timeout_seconds" INTEGER,          -- snapshot
    "command_text"    TEXT,
    "script_text"     TEXT,
    "shell_path"      TEXT,
    "working_dir"     TEXT,
    "exit_code"       INTEGER,
    -- result metrics
    "rows_count"      INTEGER,          -- rows returned (for SELECT)
    "rows_affected"   INTEGER,          -- rows affected (DML)

    -- expectation evaluation outcome
    "expectation"     JSONB,            -- snapshot of expectation config
    "expect_ok"       BOOL CHECK ("expect_ok" IN (0, 1)),             -- true/false if evaluated
    "expect_message"  TEXT,             -- details when expect_ok=false

    -- error info
    "error_code"      TEXT,
    "error_message"   TEXT,

    -- arbitrary per-step data (e.g., first/last row samples)
    "details"         JSONB,

    PRIMARY KEY ("id"),
    FOREIGN KEY ("run_id") REFERENCES job_runs (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_job_run_steps_run_order" ON "job_run_steps" ("run_id" ASC, "step_order" ASC);
CREATE INDEX IF NOT EXISTS "idx_job_run_steps_run_status" ON "job_run_steps" ("run_id" ASC, "status" ASC);

-- ----------------------------
-- job_run_events: append-only event stream per run (and optionally per step)
-- ----------------------------
DROP TABLE IF EXISTS "job_run_events";
CREATE TABLE "job_run_events" (
    "id"         INTEGER,
    "run_id"     INTEGER   NOT NULL, -- FK -> job_runs.id
    "step_id"    INTEGER,            -- nullable, FK -> job_run_steps.id
    "created_at" TIMESTAMP NOT NULL, -- event timestamp

    -- classification & message
    "kind"       TEXT      NOT NULL, -- e.g., "info", "debug", "warning", "error", "progress", "result-row"
    "message"    TEXT,               -- human-readable

    -- arbitrary machine-readable payload
    "data"       JSONB,

    PRIMARY KEY ("id"),
    FOREIGN KEY ("run_id") REFERENCES job_runs (id) ON DELETE CASCADE,
    FOREIGN KEY ("step_id") REFERENCES job_run_steps (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS "idx_job_run_events_run_time" ON "job_run_events" ("run_id" ASC, "created_at" ASC);
CREATE INDEX IF NOT EXISTS "idx_job_run_events_step_time" ON "job_run_events" ("step_id" ASC, "created_at" ASC);

PRAGMA foreign_keys = true;


-- ----------------------------
-- Table structure for agents
-- ----------------------------
DROP TABLE IF EXISTS "agents";
CREATE TABLE "agents" (
    "uuid"                TEXT NOT NULL,
    "name"                TEXT NOT NULL,
    "status"              TEXT NOT NULL CHECK ("status" IN ('active','disabled')), -- active | disabled
    "suspended"           BOOL NOT NULL CHECK ("suspended" IN (0, 1)),
    "public_key"          TEXT NOT NULL, -- Ed25519 public key (base64 or PEM)
    "approved_by_user_id" INTEGER,
    "approved_at"         TIMESTAMP,
    "last_seen_ip"        TEXT,
    "last_seen_at"        TIMESTAMP,
    "version"             TEXT,
    "os"                  TEXT,
    "arch"                TEXT,
    "os_version"          TEXT,
    "os_type"             TEXT,
    "running_user"        TEXT,
    "metadata_json"       JSONB,
    PRIMARY KEY ("uuid"),
    FOREIGN KEY ("approved_by_user_id") REFERENCES cx_users (id) ON DELETE SET NULL
);

-- Helpful indexes for querying agents
CREATE INDEX IF NOT EXISTS "idx_agents_status" ON "agents" ("status" ASC);
CREATE INDEX IF NOT EXISTS "idx_agents_last_seen" ON "agents" ("last_seen_at" DESC);

-- ----------------------------
-- Table structure for agent_registration_requests
-- ----------------------------
DROP TABLE IF EXISTS "agent_registration_requests";
CREATE TABLE "agent_registration_requests" (
    "request_id"          TEXT      NOT NULL,
    "uuid"                TEXT      NOT NULL,
    "name"                TEXT      NOT NULL,
    "public_key"          TEXT      NOT NULL, -- proposed agent Ed25519 public key
    "version"             TEXT,
    "os"                  TEXT,
    "arch"                TEXT,
    "os_version"          TEXT,
    "os_type"             TEXT,
    "running_user"        TEXT,
    "ip"                  TEXT,
    "metadata_json"       JSONB,
    "status"              TEXT      NOT NULL CHECK ("status" IN ('pending','approved','denied','expired','consumed')), -- pending | approved | denied | expired | consumed
    "created_at"          TIMESTAMP NOT NULL,
    "expires_at"          TIMESTAMP NOT NULL,
    "approved_by_user_id" INTEGER,
    "approved_at"         TIMESTAMP,
    "consumed_at"         TIMESTAMP,
    PRIMARY KEY ("request_id"),
    FOREIGN KEY ("approved_by_user_id") REFERENCES cx_users (id) ON DELETE SET NULL
);

-- Helpful indexes for registration requests
CREATE INDEX IF NOT EXISTS "idx_agent_reg_requests_status_created" ON "agent_registration_requests" ("status" ASC, "created_at" DESC);
CREATE INDEX IF NOT EXISTS "idx_agent_reg_requests_uuid_status" ON "agent_registration_requests" ("uuid" ASC, "status" ASC);

-- ----------------------------
-- Table structure for shell_connections
-- ----------------------------
DROP TABLE IF EXISTS "shell_connections";
CREATE TABLE "shell_connections" (
    "id"              INTEGER,
    "name"            TEXT      NOT NULL,
    "description"     TEXT,
    "agent_uuid"      TEXT,
    "mode"            TEXT      NOT NULL CHECK ("mode" IN ('localhost','ssh')), -- localhost | ssh
    "run_as_user"     TEXT,               -- optional (localhost or ssh)
    "sudo"            BOOL CHECK ("sudo" IN (0, 1)),               -- 0/1 boolean; app enforces usage
    "host"            TEXT,
    "port"            INTEGER CHECK ("port" > 0),
    "ssh_username"    TEXT,
    "auth_method"     TEXT CHECK ("auth_method" IN ('password','key')),               -- password | key
    "ssh_password"    TEXT,
    "ssh_private_key" TEXT,               -- PEM text
    "ssh_key_pass"    TEXT,               -- optional passphrase
    "sudo_password"   TEXT,               -- optional sudo password
    "auto_check_enabled"          INTEGER   NOT NULL CHECK ("auto_check_enabled" IN (0, 1)),
    "auto_check_interval_seconds" INTEGER   NOT NULL CHECK ("auto_check_interval_seconds" >= 0),
    "created_at"      TIMESTAMP NOT NULL,
    "updated_at"      TIMESTAMP NOT NULL,
    "last_status"     TEXT,
    "last_error"      TEXT,
    "last_checked_at" TIMESTAMP,
    "alert_emails"                TEXT,
    "alert_phones"                TEXT,
    "notify_on_failure"           BOOL CHECK ("notify_on_failure" IN (0, 1)),
    "enabled" BOOL NOT NULL CHECK ("enabled" IN (0, 1)),
    "suspended" BOOL NOT NULL CHECK ("suspended" IN (0, 1)),

    PRIMARY KEY ("id"),
    CONSTRAINT "uq_shell_connections_name" UNIQUE ("name"),
    FOREIGN KEY (agent_uuid) REFERENCES agents (uuid) ON DELETE SET NULL
);

-- removed index on enabled
CREATE INDEX IF NOT EXISTS "idx_shell_connections_agent"
    ON "shell_connections" ("agent_uuid" ASC);
CREATE INDEX IF NOT EXISTS "idx_shell_connections_mode"
    ON "shell_connections" ("mode" ASC);


-- ----------------------------
-- Table structure for webtask_connections
-- ----------------------------
DROP TABLE IF EXISTS "webtask_connections";
CREATE TABLE "webtask_connections" (
    "id"                          INTEGER,
    "name"                        TEXT      NOT NULL,
    "description"                 TEXT,
    "base_url"                    TEXT,
    "auth_type"                   TEXT      NOT NULL CHECK ("auth_type" IN ('none','basic','bearer','header')), -- none | basic | bearer | header
    "auth_config"                 JSONB,            -- credentials
    "agent_uuid"                  TEXT,
    "auto_check_enabled"          INTEGER   NOT NULL CHECK ("auto_check_enabled" IN (0, 1)),
    "auto_check_interval_seconds" INTEGER   NOT NULL CHECK ("auto_check_interval_seconds" >= 0),
    "created_at"                  TIMESTAMP NOT NULL,
    "updated_at"                  TIMESTAMP NOT NULL,
    "last_status"                 TEXT,
    "last_error"                  TEXT,
    "last_checked_at"             TIMESTAMP,
    "alert_emails"                TEXT,
    "alert_phones"                TEXT,
    "notify_on_failure"           BOOL CHECK ("notify_on_failure" IN (0, 1)),
    "enabled" BOOL NOT NULL CHECK ("enabled" IN (0, 1)),
    "suspended" BOOL NOT NULL CHECK ("suspended" IN (0, 1)),
    PRIMARY KEY ("id"),
    CONSTRAINT "uq_webtask_connections_name" UNIQUE ("name"),
    FOREIGN KEY (agent_uuid) REFERENCES agents (uuid) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS "idx_webtask_connections_agent" ON "webtask_connections" ("agent_uuid" ASC);


-- ----------------------------
-- Table structure for shell_action_steps
-- ----------------------------
DROP TABLE IF EXISTS "shell_action_steps";
CREATE TABLE "shell_action_steps" (
    "id"                       INTEGER,
    "action_id"                INTEGER NOT NULL, -- FK -> actions.id
    "step_order"               INTEGER NOT NULL CHECK ("step_order" >= 0),
    "name"                     TEXT    NOT NULL,

    -- shell execution config
    "run_mode"                 TEXT    NOT NULL CHECK ("run_mode" IN ('command','script')), -- command | script
    "command"                  TEXT,             -- when run_mode = command
    "script_text"              TEXT,             -- when run_mode = script
    "shell_path"               TEXT    NOT NULL, -- e.g., /bin/bash or /bin/sh
    "working_dir"              TEXT,
    "timeout_seconds"          INTEGER,
    "env_json"                 JSONB,            -- JSON object {name:value}

    -- output capture policy (per step)
    "output_capture_max_bytes" INTEGER NOT NULL CHECK ("output_capture_max_bytes" > 0 AND "output_capture_max_bytes" <= 1048576),
    "output_truncation"        TEXT    NOT NULL CHECK ("output_truncation" IN ('head','tail')), -- head | tail
    "expectation"              JSONB,
    "output_capture"           JSONB,
    "on_failure"               TEXT,

    -- Cross-field validation: ensure the appropriate text field is provided per run_mode
    CHECK (("run_mode" = 'command' AND "command" IS NOT NULL AND "script_text" IS NULL)
        OR ("run_mode" = 'script' AND "script_text" IS NOT NULL)) ,

    PRIMARY KEY ("id"),
    FOREIGN KEY ("action_id") REFERENCES actions (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_shell_action_steps_action_order"
    ON "shell_action_steps" ("action_id" ASC, "step_order" ASC);


-- ----------------------------
-- Table structure for webtask_action_steps
-- ----------------------------
DROP TABLE IF EXISTS "webtask_action_steps";
CREATE TABLE "webtask_action_steps" (
    "id"                       INTEGER,
    "action_id"                INTEGER NOT NULL, -- FK -> actions.id
    "step_order"               INTEGER NOT NULL CHECK ("step_order" >= 0),
    "name"                     TEXT    NOT NULL,

    "method"                   TEXT    NOT NULL CHECK ("method" IN ('GET','POST','PUT','DELETE','PATCH')),
    "url"                      TEXT    NOT NULL,
    "headers"                  JSONB,            -- JSON object {name:value}
    "body"                     TEXT,
    "timeout_seconds"          INTEGER,
    "expectation"              JSONB,
    "response_capture"         JSONB,            -- extraction rules
    "on_failure"               TEXT,

    PRIMARY KEY ("id"),
    FOREIGN KEY ("action_id") REFERENCES actions (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_webtask_action_steps_action_order"
    ON "webtask_action_steps" ("action_id" ASC, "step_order" ASC);


-- ----------------------------
-- Table structure for job_run_shell_io
-- ----------------------------
DROP TABLE IF EXISTS "job_run_shell_io";
CREATE TABLE "job_run_shell_io" (
    "id"                   INTEGER,
    "step_id"              INTEGER   NOT NULL,    -- FK -> job_run_steps.id (one row per shell step)

    "stdout_text"          TEXT,
    "stderr_text"          TEXT,
    "stdout_truncated"     BOOL      NOT NULL CHECK ("stdout_truncated" IN (0, 1)),
    "stderr_truncated"     BOOL      NOT NULL CHECK ("stderr_truncated" IN (0, 1)),
    "stdout_bytes"         INTEGER,               -- stored length
    "stderr_bytes"         INTEGER,
    "stdout_total_bytes"   INTEGER,               -- original before truncation (if known)
    "stderr_total_bytes"   INTEGER,

    PRIMARY KEY ("id"),
    CONSTRAINT "uq_job_run_shell_io_step" UNIQUE ("step_id"),
    FOREIGN KEY ("step_id") REFERENCES job_run_steps (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_job_run_shell_io_step" ON "job_run_shell_io" ("step_id" ASC);


-- ----------------------------
-- Table structure for job_run_webtask_io
-- ----------------------------
DROP TABLE IF EXISTS "job_run_webtask_io";
CREATE TABLE "job_run_webtask_io" (
    "id"                   INTEGER,
    "step_id"              INTEGER   NOT NULL,    -- FK -> job_run_steps.id

    "request_url"          TEXT,
    "request_method"       TEXT,
    "request_headers"      JSONB,
    "request_body"         TEXT,
    "response_status"      INTEGER,
    "response_headers"     JSONB,
    "response_body"        TEXT,
    "latency_ms"           INTEGER,

    PRIMARY KEY ("id"),
    CONSTRAINT "uq_job_run_webtask_io_step" UNIQUE ("step_id"),
    FOREIGN KEY ("step_id") REFERENCES job_run_steps (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_job_run_webtask_io_step" ON "job_run_webtask_io" ("step_id" ASC);

CREATE INDEX IF NOT EXISTS "idx_actions_type_name" ON "actions" ("action_type" ASC, "name" ASC);
CREATE INDEX IF NOT EXISTS "idx_shell_action_steps_action" ON "shell_action_steps" ("action_id" ASC);

-- ----------------------------
-- Table structure for webhooks
-- ----------------------------
DROP TABLE IF EXISTS "webhooks";
CREATE TABLE "webhooks" (
    "id"          INTEGER PRIMARY KEY,
    "name"        TEXT NOT NULL,
    "url"         TEXT NOT NULL,
    "secret"      TEXT,
    "events"      TEXT NOT NULL, -- comma-separated list of event categories (job, connection, system)
    "enabled"     BOOL NOT NULL CHECK (enabled IN (0, 1)),
    "created_at"  TIMESTAMP NOT NULL,
    "updated_at"  TIMESTAMP NOT NULL
);

-- ----------------------------
-- Table structure for bug_reports
-- ----------------------------
DROP TABLE IF EXISTS "bug_reports";
CREATE TABLE "bug_reports" (
    "id"          INTEGER PRIMARY KEY,
    "summary"     TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "user_id"     INTEGER NOT NULL,
    "created_at"  TIMESTAMP NOT NULL,
    "status"      TEXT NOT NULL CHECK ("status" IN ('open', 'closed', 'in-progress')),
    FOREIGN KEY ("user_id") REFERENCES cx_users ("id") ON DELETE CASCADE
);

-- ----------------------------
-- Table structure for feature_requests
-- ----------------------------
DROP TABLE IF EXISTS "feature_requests";
CREATE TABLE "feature_requests" (
    "id"          INTEGER PRIMARY KEY,
    "summary"     TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "user_id"     INTEGER NOT NULL,
    "created_at"  TIMESTAMP NOT NULL,
    "status"      TEXT NOT NULL CHECK ("status" IN ('open', 'closed', 'in-progress')),
    FOREIGN KEY ("user_id") REFERENCES cx_users ("id") ON DELETE CASCADE
);

-- ----------------------------
-- Table structure for feedback_attachments
-- ----------------------------
DROP TABLE IF EXISTS "feedback_attachments";
CREATE TABLE "feedback_attachments" (
    "id"                 INTEGER PRIMARY KEY,
    "bug_report_id"      INTEGER,
    "feature_request_id" INTEGER,
    "file_name"          TEXT NOT NULL,
    "file_path"          TEXT NOT NULL,
    "content_type"       TEXT NOT NULL,
    "file_size"          INTEGER NOT NULL,
    "created_at"         TIMESTAMP NOT NULL,
    FOREIGN KEY ("bug_report_id") REFERENCES bug_reports ("id") ON DELETE CASCADE,
    FOREIGN KEY ("feature_request_id") REFERENCES feature_requests ("id") ON DELETE CASCADE
);
