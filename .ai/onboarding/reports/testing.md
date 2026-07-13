# Testing Strategy

This document records Dockyard's discovered testing and validation strategy.

## Summary

Dockyard uses several validation layers:

- Go unit tests with `go test ./...`.
- Local verification through `make verify`.
- Static analysis through Staticcheck, govulncheck, and Semgrep.
- Package quality checks through `dockyard package lint`.
- Package validation through `dockyard package test`.
- Docker Compose validation through `docker compose config`.
- Optional Docker smoke tests through `dockyard package test --smoke`.
- Release validation through GitHub Actions.

No benchmark tests or dedicated performance tests were found.

## Unit Tests

Unit tests are Go tests under `internal`.

Discovered test files:

- `internal/catalog/catalog_test.go`
- `internal/cli/catalog_source_test.go`
- `internal/cli/install_plan_test.go`
- `internal/cli/list_test.go`
- `internal/cli/uninstall_test.go`
- `internal/dockpkg/package_test.go`
- `internal/envfile/envfile_test.go`
- `internal/lock/lock_test.go`
- `internal/oci/oci_test.go`
- `internal/policy/policy_test.go`
- `internal/quality/quality_test.go`
- `internal/render/render_test.go`
- `internal/state/state_test.go`
- `internal/values/template_test.go`

Unit test command:

```bash
go test ./...
```

Make target:

```bash
make test
```

CI command:

```bash
go test ./...
```

## Unit Test Coverage Areas

### Catalog

Tests cover:

- `catalog://` resolution with default versions.
- Bare catalog name resolution.
- Unknown bare names remaining unchanged.
- Unknown catalog names failing.
- Unknown versions failing.
- Sorted listing from a file-backed catalog.

Files:

- `internal/catalog/catalog_test.go`
- `internal/cli/catalog_source_test.go`

### OCI

Tests cover:

- OCI reference scheme requirement.
- Tag or digest requirement.
- Tag and digest acceptance.
- ORAS push argument construction.
- Dockyard artifact and layer media type usage.

File:

- `internal/oci/oci_test.go`

These are unit tests of normalization and argument construction. They do not contact a registry.

### Package Manifest and Path Safety

Tests cover:

- Manifest name validation.
- Missing API version.
- Safe path join accepting nested paths.
- Safe path join rejecting path escape.
- Dependency metadata acceptance.
- Rejection of unpinned OCI dependency sources.
- Duplicate dependency alias rejection.

File:

- `internal/dockpkg/package_test.go`

### Dependency Planning

Tests cover:

- Dependency-aware install plan construction.
- Existing release detection.
- Dry-run plan matching install-plan behavior.
- Failed dependencies blocking automatic install.
- Dependency release naming.
- Dependency install options not reusing root values or overlay.
- Inline dependency values temporary-file behavior.
- Rejection of inline values with existing values file.

File:

- `internal/cli/install_plan_test.go`

### Release List and Uninstall Safety

Tests cover:

- Hiding uninstalled releases by default.
- Including uninstalled releases with `--all`.
- Status filtering.
- Relationship summaries.
- Active dependent release detection.
- Ignoring uninstalled parents.
- Dependency uninstall block error text.

Files:

- `internal/cli/list_test.go`
- `internal/cli/uninstall_test.go`

### Env Files

Tests cover:

- Env var name generation.
- Template generation and sensitive masking.
- Duplicate key detection.
- Populated secret-like value detection.
- Quoted values.
- Duplicate rejection for process env loading.

File:

- `internal/envfile/envfile_test.go`

### Lock, Render, Policy, State, Values, Quality

Tests cover:

- Image extraction and dependency sorting in lockfiles.
- Value flattening and sensitive key detection.
- Privileged service detection and secure service pass cases.
- Release name validation.
- Values template generation using schema descriptions and sensitive handling.
- Schema quality inspection and strict/advisory blocking behavior.

Files:

- `internal/lock/lock_test.go`
- `internal/render/render_test.go`
- `internal/policy/policy_test.go`
- `internal/state/state_test.go`
- `internal/values/template_test.go`
- `internal/quality/quality_test.go`

## Integration Tests

No separate integration test project or integration test directory was found.

Some Go tests exercise multiple internal components together using temporary directories, especially:

- Install-plan tests.
- Catalog source tests.
- Env-file tests.

However, these still run as normal Go unit tests and do not start Docker, call ORAS, or contact a registry.

## Smoke Tests

Smoke testing is documented as Docker-backed and optional.

Command:

```bash
dockyard package test ./examples/nginx --smoke
```

Behavior from code:

1. Preflight Docker CLI availability.
2. Check Docker CLI usability.
3. Check Docker Compose availability.
4. Check Docker daemon reachability.
5. Render package Compose.
6. Run `docker compose up -d` with a temporary project name.
7. Run `docker compose ps --all`.
8. Defer `docker compose down`.
9. Do not write Dockyard release state.

Implementation:

- `internal/cli/package.go`

Documentation says smoke tests require Docker Desktop or a reachable Docker daemon and remain local/manual for release workflows.

## Benchmark Tests

No Go benchmark functions were found.

Evidence:

- Search for `func Benchmark` in `*_test.go` returned no matches.

## Performance Tests

No dedicated performance test project, performance test command, or benchmark workflow was found.

## Package Validation

Package validation exists in two main commands:

```bash
dockyard package lint PACKAGE_DIR [--strict] [--allow-advisory] [--json]
dockyard package test PACKAGE_SOURCE [--strict] [--allow-advisory] [--smoke] [--env-file file]
```

`dockyard package lint` checks:

- `Dockyard.yaml` validity.
- Dependency metadata.
- Required/recommended metadata files: `README.md`, `SECURITY.md`, `LICENSE`.
- Forbidden local artifacts and secret-like files.
- `values.yaml` loading.
- `values.schema.json` validation against default values.
- Schema descriptions for public values.
- Sensitive schema markers.
- Default Compose render.
- Policy findings on default render.

`dockyard package test` checks:

- Source preparation from directory, archive, or OCI.
- Package quality checks.
- Values and schema validation through package build.
- Compose rendering.
- Dockyard policy checks.
- `docker compose config` unless skipped.
- Optional smoke up/down with `--smoke`.

## Container Validation

Container validation is Compose-based.

Non-destructive container/Compose validation:

```bash
dockyard config PACKAGE_SOURCE -f values.yaml
dockyard render PACKAGE_DIR -f values.yaml --validate-compose
dockyard package test PACKAGE_SOURCE
```

These validate rendered Compose with:

```bash
docker compose config
```

Runtime smoke validation:

```bash
dockyard package test PACKAGE_SOURCE --smoke
```

This starts and stops containers with Docker Compose using a temporary project name.

## Registry Validation

Registry behavior exists through ORAS and catalog code.

Current automated/unit coverage:

- OCI reference normalization.
- Tag/digest requirement.
- ORAS push argument construction.
- Catalog resolution and file-backed catalog loading.

Registry validation commands documented:

```bash
oras login ghcr.io
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz oci://ghcr.io/nandub/dockyard/nginx:0.1.0
dockyard pull oci://ghcr.io/nandub/dockyard/nginx:0.1.0 -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard package test oci://ghcr.io/nandub/dockyard/nginx:0.1.0 --smoke
```

No CI workflow was found that logs into a registry, pushes a test package, pulls it back, or validates against a live registry.

## Coverage

Coverage output is recognized by `.gitignore`:

```text
/coverage.out
```

No coverage command or coverage reporting workflow was found.

Possible manual command:

```bash
go test ./... -coverprofile=coverage.out
```

This is not documented as an authoritative repository command.

## Current CI Validation

Normal CI:

```bash
go mod tidy
gofmt -w ./cmd ./internal
git diff --exit-code
go test ./...
go build -o bin/dockyard ./cmd/dockyard
```

Security workflow:

```bash
go mod tidy
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
semgrep scan --config auto
```

Release workflow:

```bash
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go build ...
./dist/dockyard-linux-amd64 version
./dist/dockyard-linux-amd64 compat ./examples/nginx --strict
./dist/dockyard-linux-amd64 package test ./examples/nginx --strict
```

The release binary package checks run only for Linux AMD64.

## Missing Areas

Missing or limited validation areas:

- No benchmark tests.
- No dedicated performance tests.
- No coverage reporting gate.
- No dedicated integration test suite naming or separation.
- No CI Docker smoke tests.
- No CI package validation for every example package.
- No live OCI registry round-trip test in CI.
- No ORAS command execution test with a fake or local registry.
- No Dockerfile build validation, because no Dockerfiles exist.
- No end-to-end install/upgrade/rollback/uninstall test that runs against Docker in CI.
- No dedicated tests found for archive packaging in `internal/archive`, despite archive handling being security-sensitive.
- No dedicated tests found for `internal/runner`.
- No dedicated tests found for release upgrade/rollback command behavior beyond install-plan/list/uninstall helper tests.
- No documented mutation-free coverage command.

## Recommended Validation Pipeline

### Fast Local Gate

Run before committing ordinary code changes:

```bash
go mod tidy
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Rationale:

- Aligns with `README.md`, `CONTRIBUTING.md`, and the Makefile.
- Covers tidy, formatting, and unit tests.
- Adds Staticcheck, which is documented but not part of `make verify`.

### Security Gate

Run for security-sensitive changes or before release:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Semgrep is currently run in GitHub Actions through the Semgrep container.

### Package Gate

Run when changing package behavior, examples, rendering, values/schema logic, archives, policy, or docs that affect package authors:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

For dependency behavior changes:

```bash
dockyard package lint ./examples/postgres --strict
dockyard package test ./examples/postgres --strict
dockyard package deps ./examples/team-dashboard
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard
```

### Public Example Gate

Before release or broad packaging changes, repeat non-smoke package gates for every public example:

```bash
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Repeat for:

- `examples/postgres`
- `examples/postgres-app`
- `examples/team-dashboard`
- `examples/caddy-letsencrypt`
- `examples/nginx-tls-mounted-certs`
- `examples/traefik-letsencrypt`

### Docker Smoke Gate

Run only when Docker Desktop or Docker Engine is available:

```bash
dockyard doctor
dockyard package test ./examples/nginx --smoke
```

For affected runnable examples, run `--smoke` selectively.

### Registry Gate

Run only when ORAS and registry credentials are available:

```bash
oras login ghcr.io
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz oci://ghcr.io/nandub/dockyard/nginx:0.1.0
dockyard pull oci://ghcr.io/nandub/dockyard/nginx:0.1.0 -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard package test oci://ghcr.io/nandub/dockyard/nginx:0.1.0 --strict
```

Use a disposable repository/tag for test pushes.

### Release Gate

Before tagging:

```bash
go mod tidy
make verify
make dev-build
dockyard version
dockyard compat
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Then run package gates and smoke gates as applicable.

Official release publishing remains the `v*` tag-triggered GitHub Actions workflow.

## Evidence

- `Makefile`
- `.github/workflows/ci.yml`
- `.github/workflows/security.yml`
- `.github/workflows/release.yml`
- `CONTRIBUTING.md`
- `README.md`
- `docs/command-reference.md`
- `docs/packaging-and-distribution.md`
- `docs/release-candidate-checklist.md`
- `internal/cli/package.go`
- `internal/quality/quality.go`
- `internal/*/*_test.go`
