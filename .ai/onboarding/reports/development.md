# Development Workflow

This document records how developers are expected to work in this repository, based on repository docs, the Makefile, and GitHub Actions.

## Authoritative Sources

Use these sources in this order:

1. `Makefile` for local build, format, tidy, test, and verify targets.
2. `.github/workflows/*.yml` for hosted CI, security, and release behavior.
3. `CONTRIBUTING.md` for pull request expectations.
4. `README.md` and `docs/getting-started.md` for onboarding and smoke-test workflow.
5. `docs/release-engineering.md` and `docs/release-candidate-checklist.md` for release-specific workflow.

## Bootstrap

Minimum local setup:

```bash
go mod tidy
make dev-build
```

Evidence:

- `CONTRIBUTING.md` lists `go mod tidy` and `make dev-build` under local setup.
- `README.md` lists `go mod tidy`, `make verify`, `make dev-build`, and Staticcheck as common development commands.
- `docs/getting-started.md` recommends `make dev-build` and `make test` for first-time local development.

Recommended local working directories:

```text
../dockyard-work/
../deploy-values/
../dockyard-artifacts/
```

These are for generated test packages, operator values, env files, rendered files, and package archives. They should stay outside the repository.

## Restore

Restore is Go module based:

```bash
go mod tidy
```

For verification without accepting dependency changes:

```bash
make tidy-check
```

`make tidy-check` runs:

```bash
go mod tidy -diff
```

Authority:

- Local mutation path: `go mod tidy`.
- Local verification path: `make tidy-check` or `make verify`.
- CI path: `.github/workflows/ci.yml` and `.github/workflows/security.yml` run `go mod tidy`, then CI relies on `git diff --exit-code` only in the CI workflow.

## Build

Local build:

```bash
make build
```

`make build` emits:

- `bin/dockyard.exe` on Windows.
- `bin/dockyard` on Linux and macOS.

Development build:

```bash
make dev-build
```

`make dev-build` runs:

```text
tidy -> fmt -> build
```

Manual build forms are documented:

```powershell
go mod tidy
go test ./...
go build -o bin/dockyard.exe ./cmd/dockyard
```

```bash
go mod tidy
go test ./...
go build -o bin/dockyard ./cmd/dockyard
```

Authority:

- Prefer `make build` for non-mutating local builds.
- Prefer `make dev-build` for local convenience when tidy and formatting changes are acceptable.
- CI uses direct `go build -o bin/dockyard ./cmd/dockyard`, not `make build`.

## Format

Mutation command:

```bash
make fmt
```

`make fmt` runs:

```bash
go fmt ./...
```

Verification command:

```bash
make fmt-check
```

`make fmt-check` runs:

```bash
go run ./tools/fmtcheck ./cmd ./internal ./tools
```

CI formatting:

```bash
gofmt -w ./cmd ./internal
git diff --exit-code
```

Authority:

- Local check authority is `make fmt-check`.
- Local mutation authority is `make fmt`.
- CI's formatting command is narrower than `make fmt`; see contradictions.

## Lint

Staticcheck is the documented Go lint command:

```bash
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Security lint/scanning in GitHub Actions:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
semgrep scan --config auto
```

Package/application quality lint:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Authority:

- Staticcheck is required before opening a PR according to `CONTRIBUTING.md`.
- Security workflow is authoritative for hosted vulnerability and Semgrep scanning.
- Package lint commands are authoritative for package/example changes.

## Test

Unit test command:

```bash
make test
```

`make test` runs:

```bash
go test ./...
```

Pre-commit verification:

```bash
make verify
```

`make verify` runs:

```text
tidy-check -> fmt-check -> test
```

Package tests:

```bash
dockyard package test ./examples/nginx --strict
```

Optional smoke test when Docker is available:

```bash
dockyard package test ./examples/nginx --smoke
```

Authority:

- `make verify` is the local pre-commit authority.
- `go test ./...` is the core test command in Makefile and CI.
- `dockyard package test` is the package-author verification command.

## Package

Package validation and publishing preparation:

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
```

Authority:

- `docs/packaging-and-distribution.md` has the most complete package workflow.
- `README.md` has the short package publishing example.
- `docs/command-reference.md` lists the command surface and flags.

## Publish

Publish package archive:

```bash
oras login ghcr.io
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz oci://ghcr.io/nandub/dockyard/nginx:0.1.0
```

Publish catalog index:

```bash
oras push --artifact-type application/vnd.dockyard.catalog.v1+yaml ghcr.io/nandub/dockyard-packages/catalog:latest catalog.yaml:application/vnd.dockyard.catalog.index.v1+yaml
```

Authority:

- Package publishing is documented in `README.md` and `docs/packaging-and-distribution.md`.
- Registry authentication is delegated to `oras`; Dockyard should not receive credentials directly.

## Release

Local release snapshot:

```bash
make release-snapshot VERSION=v1.0.0
```

GitHub release:

```bash
git tag v1.0.0
git push origin v1.0.0
```

Authority:

- `.github/workflows/release.yml` is authoritative for official release builds.
- `docs/release-engineering.md` and `docs/release-candidate-checklist.md` describe the human process.

## Contradictions and Drift

- CI formats only `./cmd ./internal`, while `make fmt` formats `./...` and `make fmt-check` checks `./cmd ./internal ./tools`.
- CI runs mutating `go mod tidy` and `gofmt -w`, then checks `git diff --exit-code`; local `make verify` uses non-mutating `go mod tidy -diff` and `fmtcheck`.
- CI builds with direct `go build -o bin/dockyard ./cmd/dockyard`; local docs prefer `make build`, which injects version metadata through Makefile `LDFLAGS`.
- `docs/packaging-and-distribution.md` says the default catalog registry is `ghcr.io/nandub/dockyard-packages`, while `README.md`, `docs/command-reference.md`, `AGENTS.md`, and code use `oci://ghcr.io/nandub/dockyard-packages/catalog:latest`.
- `docs/release-engineering.md` says the release workflow runs `dockyard version`, `dockyard compat`, and `dockyard package test`; in the workflow, those run only for the Linux AMD64 matrix entry.

## Unverified

These commands were not run during this documentation pass:

- `make verify`
- `make dev-build`
- Staticcheck
- govulncheck
- package tests
- Docker smoke tests
