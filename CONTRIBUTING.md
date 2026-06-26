# Contributing to Chronix

Chronix is open source and community contributions are welcome.

## Development

- Use Go for the server and React/TypeScript for the embedded UI.
- Keep changes focused and consistent with the existing package boundaries.
- Run `go test ./...` before submitting backend changes.
- Run `npm run build` from `internal/chronix_react_app` before submitting UI changes.
- Update docs when behavior, commands, settings, or visible workflows change.

## Pull Requests

Open a pull request with:

- a short summary of what changed,
- why the change is useful,
- validation performed,
- screenshots for meaningful UI changes.

Please avoid committing local databases, logs, private keys, IDE metadata, or generated release artifacts.
