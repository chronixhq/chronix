# Chronix Tester

`chronix-tester` is the local fixture harness for Chronix development and release validation.

It gives Chronix a controlled environment to exercise:

- SQLite database actions against a known target database
- shell multi-step actions with variable capture
- web-task actions against local HTTP fixtures
- webhook delivery capture
- IMAP polling for notification verification
- a small results UI plus JSON snapshot endpoints for quick inspection and future automation

## Why It Exists

This tool is the bridge between ad-hoc manual testing and a real automated Chronix validation platform.
It is intentionally external to the main server so we can test Chronix as a consumer would use it: through real connections, actions, jobs, HTTP calls, shell commands, and notifications.

## Commands

```bash
go run ./dev/chronix_tester run
go run ./dev/chronix_tester bootstrap ./chronix.db localhost
go run ./dev/chronix_tester config imap --host imap.example.com --user bot@example.com --pass secret
go run ./dev/chronix_tester generate-token
go run ./dev/chronix_tester reset
go run ./dev/chronix_tester version
```

The tester stores its local state in `dev/chronix_tester` by default. Override that with `CHRONIX_TESTER_DATA_DIR` or `--data-dir`.

## Local Surfaces

When `run` is active:

- Results UI: `http://127.0.0.1:5180/`
- Results JSON snapshot: `http://127.0.0.1:5180/api/snapshot`
- API fixtures: `http://127.0.0.1:5181`
- Webhook capture: `http://127.0.0.1:5182`

Useful API fixture endpoints:

- `GET /json`
- `GET /html`
- `GET /get`
- `GET /response-headers?X-Test=ABC`
- `GET /token`
- `POST /echo`
- `GET /status/200`
- `GET /delay/250`

## Bootstrap Fixtures

`bootstrap` seeds a Chronix database with four tester connections, four tester actions, and four manual jobs.

Coverage currently includes:

- DB session persistence via temp-table state and `last_insert_rowid()`
- shell JSONPath capture, regex capture, line assertions, and exit-code assertions
- web-task JSONPath capture, header capture, regex capture, and piping
- webhook delivery capture

Bootstrap is now convergent for tester fixtures:

- named connections and actions are updated in place
- tester steps are replaced so the fixture definition stays current
- tester jobs are recreated cleanly so repeated bootstrap runs do not stack duplicates

## Notes

- The shell fixture is currently POSIX-shell oriented and assumes `/bin/sh`.
- IMAP support is intentionally lightweight right now and is focused on confirming message arrival.
- Generated local artifacts are no longer tracked in Git; they are created on demand.

## Next Platform Direction

This tool is now in a better place to become the base of a fuller automated Chronix test platform. The next natural steps are:

- add a first-class `verify` command that runs Chronix jobs end-to-end and returns a pass/fail summary
- drive Chronix through API calls instead of manual UI setup after bootstrap
- add fixture scenarios for schedules, cancellations, retries, notifications, and agent-proxied execution
- capture richer notification evidence, ideally with a local SMTP sink instead of IMAP polling
- add browser-based UI automation against the embedded React app, likely with Playwright, using this tester as the stable backend fixture system
