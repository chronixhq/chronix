# Milestone: Develop a Plan

Goal: Chronix can schedule and execute SQL jobs against any GORM-supported database (e.g., PostgreSQL), using a local SQLite database for internal storage of schedules, logs, and settings.

This includes:
- Defining how users create and manage schedules via a web-based UI
- Executing raw SQL commands on external databases (PostgreSQL, etc.)
- Logging job runs, including execution status, timing, and errors
- Providing a web interface to view schedules and job execution history

## Open Questions for v0.1

- What scheduling formats are supported in v0.1? (cron only? presets?)
- How do users define start/end times or blackout windows?
- How is database connection info stored and managed?
- Will users enter raw SQL strings only, or can we validate schema?
- What happens after a job runs? (log only? notify?)
- How are job results and history displayed in the CLI?
- Do we support log retention limits or cleanup?
- What basic safeguards are needed to avoid double execution?

## Tasks

- [ ] Answer open questions to develop a plan for v0.1.0:
  - [ ] Decide on supported scheduling formats (cron, presets)
  - [ ] Define how users configure start/end times or blackout periods
  - [ ] Determine how database connection info is collected and stored
  - [ ] Decide on job definition input method (raw SQL only vs schema-aware)
  - [ ] Define behavior for job result handling (logging, notifications)
  - [ ] Establish how job history and logs are viewed in CLI
  - [ ] Decide on log retention and cleanup strategy
  - [ ] Determine approach to avoid duplicate execution

- [ ] Based on answers above, define GORM models for jobs and logs
- [ ] Begin development of a React-based web interface for managing schedules and job results

## Status

In Progress

## Postmortem (to be filled out after release)

- What worked:
- What needs improvement:
- What surprised me:
- 