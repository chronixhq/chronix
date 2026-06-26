# Project Guidelines

## Project Overview

This repository contains the Chronix internal React application built with Vite and TypeScript. It provides the web UI for Chronix. The app is structured as a typical Vite + React + TS project and uses ESLint for linting.

- Tooling: Vite, React, TypeScript
- Package manager: npm (package-lock.json present)
- Entry HTML: index.html
- Dev server/build config: vite.config.ts
- TypeScript configs: tsconfig.json, tsconfig.app.json, tsconfig.node.json
- Lint config: eslint.config.js

## Directory Structure (top-level)

- src/ — application source code (components, modules, and app bootstrap)
- public/ — static assets copied as-is
- dev/ — development utilities/scripts (if any)
- dist/ — production build output (generated)
- .junie/ — Junie configuration and guidelines

Within src/ (partial, based on recent files):

- src/main/ — app bootstrap and layout (App.tsx, main.tsx, NavDrawer.tsx, TopAppBar.tsx, MainContent.tsx)
- src/modules/ — feature modules (e.g., User/ForgotPassword.tsx, Testing/Testing.tsx)

## How to Run

- Install dependencies: npm install
- Start dev server: npm run dev
- Open the URL printed by Vite (usually http://192.168.0.102:5173)

## How to Build

- Production build: npm run build
- Preview build locally: npm run preview

## Testing

- No explicit test runner is configured in the provided structure. If tests are added later, prefer npm scripts (e.g., npm test) and document any non-standard steps here.

## Code Style

- **Imports & Linting:** Whenever you make ANY changes to React files, you MUST check the imports to ensure they are correct and complete. After making changes, run `npm run lint` within this directory to verify.
- Run lint: npm run lint (if defined) or configure ESLint via eslint.config.js
- Use TypeScript strictness defined in tsconfig files
- Prefer functional React components and hooks
- Do not call hooks off the React namespace (e.g., avoid React.useEffect). Import hooks by name and call them directly (e.g., import { useEffect, useState } from 'react' then useEffect(...)).
- React components: Do not annotate components with React.FC. Rely on inference for function components and type props explicitly when needed (e.g., const MyComp = ({id}: {id: string}) => { ... }). This avoids unnecessary clutter and default children props.
- Keep module organization under src/modules and shared layout in src/main
- JSON and TypeScript property naming: prefer camelCase in the frontend. If backend responses use snake_case, normalize them to camelCase at the data/context layer before passing into components and types. Avoid casting to any; prefer typed normalization helpers.

## HTTP/API

- Use @dsherwin/react-api-interface for all HTTP calls (apiGet, apiPost, apiPut, apiPatch, apiDelete). Do not use window.fetch directly in app code.
- Base URL is configured once in AuthContext.tsx via setAPIBaseURL; development points to http://192.168.0.102:6060, production uses window.location.origin.
- Example:
    - const res = await apiPost('/api/admin/initialize', payload)
    - const data = await apiGet('/server/status')
- Error handling: apiGet returns parsed JSON; apiPost/apiPut return a Response-like object. Check res.ok or use helpers from the package as they are introduced.

## UI Components (MUI)

- MUI version: We standardize on Material UI v7.3.4 for this app. Always consult v7.3 docs (https://mui.com/material-ui/react-button/) and APIs for guidance, not older v5/v6 content. Prefer the API search scoped to v7.3.
  - Do: use `slotProps` APIs and `sx` styling; follow v7 component prop names and deprecations.
  - Do not: use legacy `@mui/styles`, `makeStyles`, or v5-only patterns; avoid examples that rely on deprecated props.
  - Notes specific to v7.3.4:
    - `Typography` `paragraph` is deprecated in v7; when you need a `<p>`, use the `component` prop.
    - `TextField` native input attributes go under `slotProps.htmlInput` (not `inputProps`).
- TextField: `inputProps` is deprecated. Use `slotProps.htmlInput` for native input attributes.
- For numeric-only input, prefer the following pattern:

```tsx
<TextField
  label="Value"
  placeholder="1"
  value={(step.expectation as ExpectRowsAffected).value || ''}
  onChange={(e) => onChange({
    kind: 'rowsAffected',
    op: (step.expectation as ExpectRowsAffected).op || '>=',
    value: e.target.value.replace(/[^0-9]/g, ''),
  })}
  sx={{ minWidth: { xs: '100%', md: 160 } }}
  slotProps={{
    htmlInput: {
      inputMode: 'numeric',
      pattern: '[0-9]*',
    },
  }}
  error={!!fieldErrors?.expectationValueNum}
  helperText={fieldErrors?.expectationValueNum}
/>
```

## Junie Guidance

- When making changes, update this guidelines file if build/run/test steps change.
- Keep edits minimal to satisfy issues; run build locally when changing TypeScript or Vite config.
- If a requested task appears egregiously wrong or there’s an obviously better approach, pause and ask clarifying questions before proceeding.
