# CI Workflow

This document maps GitHub Actions workflows and compares them with local documentation.

## Workflow Inventory

GitHub Actions workflows:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/workflows/security.yml`

GitHub templates:

- `.github/PULL_REQUEST_TEMPLATE.md`
- `.github/ISSUE_TEMPLATE/bug_report.md`
- `.github/ISSUE_TEMPLATE/feature_request.md`

## CI Workflow

File:

- `.github/workflows/ci.yml`

Triggers:

- Push to `main`.
- Pull request.

Permissions:

- `contents: read`

Runner:

- `ubuntu-latest`

Steps:

```bash
actions/checkout@v4
actions/setup-go@v5 with go-version-file: go.mod and cache: true
go mod tidy
gofmt -w ./cmd ./internal
git diff --exit-code
go test ./...
go build -o bin/dockyard ./cmd/dockyard
```

Purpose:

- Verify module tidiness and formatting through mutation plus diff.
- Run all Go tests.
- Build the Linux CI binary.

## Security Workflow

File:

- `.github/workflows/security.yml`

Triggers:

- Push to `main`.
- Pull request.
- Weekly schedule: `19 3 * * 1`.

Permissions:

- `contents: read`
- `security-events: write`

Jobs:

- `go-security`
- `semgrep`

Go security steps:

```bash
actions/checkout@v4
actions/setup-go@v5 with go-version-file: go.mod and cache: true
go mod tidy
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Semgrep steps:

```bash
container image: semgrep/semgrep:latest
actions/checkout@v4
semgrep scan --config auto
```

Purpose:

- Vulnerability scan.
- Staticcheck lint.
- Semgrep static analysis.

## Release Workflow

File:

- `.github/workflows/release.yml`

Trigger:

- Push tags matching `v*`.

Permissions:

- `contents: write`

Jobs:

- `build`
- `checksums-and-sbom`

Build matrix:

- `windows/amd64`
- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

Build job steps:

```bash
actions/checkout@v4
actions/setup-go@v5 with go-version-file: go.mod and cache: true
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go build -trimpath -ldflags "...Version=${github.ref_name} Commit=${github.sha} Date=${github.event.head_commit.timestamp}" -o dist/<asset> ./cmd/dockyard
```

Linux AMD64 binary verification:

```bash
./dist/dockyard-linux-amd64 version
./dist/dockyard-linux-amd64 compat ./examples/nginx --strict
./dist/dockyard-linux-amd64 package test ./examples/nginx --strict
```

Checksums and SBOM steps:

```bash
actions/download-artifact@v4
anchore/sbom-action/download-syft@v0
syft dir:. -o spdx-json=dist/dockyard-source.spdx.json
sha256sum ... > SHA256SUMS
softprops/action-gh-release@v2
```

Release assets:

- `dockyard-windows-amd64.exe`
- `dockyard-linux-amd64`
- `dockyard-linux-arm64`
- `dockyard-darwin-amd64`
- `dockyard-darwin-arm64`
- `SHA256SUMS`
- `dockyard-source.spdx.json`

## Local Documentation Comparison

README common development commands:

```bash
go mod tidy
make verify
make dev-build
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

CONTRIBUTING pre-PR commands:

```bash
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Package/example change commands:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
dockyard package test ./examples/nginx --smoke
```

Makefile verification:

```bash
make verify
```

Which expands to:

```text
tidy-check -> fmt-check -> test
```

## Contradictions and Drift

- CI uses `go mod tidy`, while `make verify` uses `go mod tidy -diff`.
- CI uses `gofmt -w ./cmd ./internal`, while `make fmt-check` uses `go run ./tools/fmtcheck ./cmd ./internal ./tools`.
- CI excludes `tools` from the direct `gofmt -w` step, but `make fmt-check` includes `tools`.
- CI build does not use `make build`, so it does not use Makefile `LDFLAGS` version metadata.
- Security workflow runs `go mod tidy` but does not follow with `git diff --exit-code`, so it may not fail on module drift unless later commands fail.
- Package/example gates from `CONTRIBUTING.md` are not run in the normal CI workflow.
- Release workflow runs package checks only for Linux AMD64, not for every matrix artifact.

## Command Authority

- Normal PR CI authority: `.github/workflows/ci.yml`.
- Hosted security authority: `.github/workflows/security.yml`.
- Official release authority: `.github/workflows/release.yml`.
- Local pre-PR authority: `make verify` plus Staticcheck from `CONTRIBUTING.md`.
- Package/example local authority: package gates in `CONTRIBUTING.md` and `docs/packaging-and-distribution.md`.

## Unverified

This document does not assert current pass/fail status for GitHub Actions. It records configured commands only.
