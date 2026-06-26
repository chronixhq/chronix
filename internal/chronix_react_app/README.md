# Chronix React App

The Chronix dashboard lives here and is built into the main Chronix server binary.

## Current Stack

- React 19
- TypeScript 6
- Vite 8
- Material UI 9
- MUI X Data Grid / Date Pickers
- Vitest
- ESLint flat config with `typescript-eslint`, `react-hooks`, `react-refresh`, and `unused-imports`

## Key Commands

- `npm install`
- `npm run dev`
- `npm run build`
- `npm run lint`
- `npm run lint:fix`
- `npm run test:run`

## Architecture Notes

- `src/main/` contains the app shell, routing, navigation, and top-level boot flow.
- `src/data/` contains providers and shared app state.
- `src/modules/` contains feature areas such as Actions, Jobs, Runs, Connections, Agents, Settings, and User flows.
- `src/lib/` contains shared helpers, formatting, utility hooks, and cross-feature UI helpers.

## Important Boundaries

- `@dsherwin/mui-kit` is a real dependency boundary. If a React or Material UI upgrade starts breaking multiple screens in similar ways, check that package early instead of patching every use site locally.
- The shell route/navigation metadata is centralized in `src/main/appShellManifest.ts`.
- Route-level lazy loading is defined from the shell layer and should stay coordinated with navigation and boot-state behavior.

## Build And Embed Flow

- `npm run build` outputs the dashboard to `dist/`.
- The built assets are embedded into the Go server from this project directory.
- Changes to frontend routing, help anchors, or shell structure should be considered alongside the embedded docs under `../docs/`.

## Tooling Notes

- The project intentionally does not use `eslint-plugin-react`.
- `package.json` pins top-level `eslint` and `typescript` through `overrides` so clean installs stay deterministic.
- `npm-check-updates` is configured to reject automatic `eslint` major bumps by default.
