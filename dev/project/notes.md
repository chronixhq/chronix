# Next Session – Reminders

Date captured: 2025-10-30 06:02

Context
- Area: React UI
- Where I left off: I want to move the RunNow handler and progress handler into its own component and possibly its own context provider so we can have a single, replicable progress panel that persists across navigation.

Must-remember items
- [ ] Extract the "Run Now" trigger logic (handler) into a shared component/provider.
- [ ] Extract the live run progress state/logic into a shared context provider (or hook with provider) so that any page can render a consistent progress panel.
- [ ] Ensure the progress panel can render anywhere in the app and always has up-to-date data for the active run, even while navigating between routes.
- [ ] Keep using @dsherwin/react-api-interface for REST; use SSE (job_progress, job_finished) via SseContext. Consider centralizing buffering/deduping in the provider.
- [ ] Maintain global finish snack (success/error) behavior; start snack remains where appropriate.

Open questions / decisions
- Should the provider track multiple concurrent runs (queue) or just the most recent/active one per job?
- Where to mount the provider? Likely at App-level under SseProvider so it spans the entire UI.
- API for panel consumers: minimal props (maybe just runId?), or zero-prop panel that attaches to the provider’s current active run?

Quick test checklist
- [ ] Start a long-running job from Jobs page; navigate to Runs list and Run Detail — the panel continues to show live updates.
- [ ] Start a very fast job; panel appears and remains with final state until closed (if closable).
- [ ] Start a new run after closing the panel; it reopens for the new run.
- [ ] Verify no duplicate snacks are shown on finish.

Notes usage guideline
- Any time I say “remember this” or “take a note,” append it here with a timestamp under a new subheading.
- Keep entries concise and actionable; link to relevant files/routes when helpful.


## Run Now UX — Phase 2 (2025-11-02)

- Notification policy: dialog-only. Removed global snackbar in `App.tsx`; completion is communicated via `GlobalRunFinishedDialog`.
- Multi-run global panel: `GlobalRunProgressPanel` renders one `LiveRunNowProgressPanel` per active run.
- Start flow: `runNow(jobId, { jobName? })` returns `runId` from parsed JSON (`{ status: "queued", runId }`). On success, run is added to active panels immediately.
- Titles: panel titles include job name when provided (e.g., `Run now — <Job Name>`).
- Cancel action: each panel has a `Cancel run` button wired to `POST /runs/:runId/cancel`, showing intermediate `cancellationRequested` state until `job_finished`.
- Completion: on `job_finished`, the panel is removed automatically and a completion dialog is queued with status/message and a link to `/runs/:id`.
- SSE: UI consumes `job_progress` and `job_finished` events; duplicate `job_finished` events are deduped for 60s per runId.
- Canceled semantics: user cancellations now surface as `canceled` (not `error`) end-state for both runs and jobs. This is reflected in SSE (`job_finished.status = "canceled"`) and in the `JobStatusChip`.

Verification checklist:
- Start two runs; both panels appear and stream progress; dialogs queue on finish.
- Cancel one run; panel shows `cancellationRequested`, then disappears when finished; dialog indicates canceled.
- Jobs and Runs views update status via SSE without a full reload.

Notes:
- If a caller omits `jobName` when invoking `runNow`, panel title falls back to `Live run progress`.
- Consider adding light toasts for cancel errors via `NotificationsProvider` if desired.


## Guidelines update — Type-only imports (2025-11-02)

- Rule: Always import types using the `type` modifier (TypeScript 5+).
  - Correct: `import { DataGrid, type GridColDef } from '@mui/x-data-grid'`
  - Incorrect: `import { DataGrid, GridColDef } from '@mui/x-data-grid'`
- Enforcement:
  - TS config: `verbatimModuleSyntax: true` (already set in `internal/chronix_react_app/tsconfig.app.json`).
  - ESLint: `@typescript-eslint/consistent-type-imports` set to `error` with autofix; see `internal/chronix_react_app/eslint.config.js`.
  - NPM script: run `npm run lint:fix` in `internal/chronix_react_app` to auto-convert imports.
- Rationale: prevents runtime import issues when types are imported as values, improves tree-shaking, aligns with TS 5 semantics.
- Scope: Apply this rule across all TS/JS packages in the repo. Add the ESLint rule to additional packages if/when they are introduced.
