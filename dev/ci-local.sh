#!/usr/bin/env bash
set -euo pipefail
trap "rm -f coverage.out" EXIT

# Keep the local CI gate on the patched Go toolchain so stdlib vuln checks
# and builds don't depend on whatever Homebrew currently has linked.
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.26.2}"

# 1) Tidy and ensure no changes
go mod tidy

# 2) Build, Vet, Test (race)
go build ./...
go vet $(go list ./... | grep -v "/internal/sqlsyntax" | grep -v "/internal/db")
go test $(go list ./... | grep -v "/internal/sqlsyntax" | grep -v "/internal/db") -race -count=1 -covermode=atomic -coverprofile=coverage.out
go tool cover -func=coverage.out \
| awk -v thr="${COVER_THRESH:-12}" '
/^total:/ {
  gsub(/%/, "", $3);    # strip percent sign
  cov = $3 + 0;
  if (cov < thr) {
    printf "FAIL: coverage %.1f%% < %d%%\n", cov, thr;
    exit 1
  } else {
    exit 0
  }
}
END {
  # If we never saw a total line, treat as failure
  if (NR == 0) { print "ERROR: no coverage data."; exit 2 }
}'
rm -f coverage.out

# 3) Lint
GOGC=off golangci-lint config verify
GOGC=off golangci-lint run --timeout 5m

# 4) Vulnerabilities
govulncheck -test ./...

#  5) gofmt test
test -z "$(gofmt -s -l .)" || { echo "gofmt needed"; exit 1; }


# 6) React NPM linter

# Ensure the React app lints cleanly
cd internal/chronix_react_app || exit 1
# Install dependencies in a CI-friendly way and run lint
npm ci --no-audit --no-fund
npm run lint
cd ../..
