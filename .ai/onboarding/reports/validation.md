# Validation Pipeline

This document proposes an authoritative validation pipeline for Dockyard based on discovered commands and current gaps.

## Validation Layers

Dockyard validation should be treated as layered:

1. Source hygiene.
2. Unit tests.
3. Static analysis and vulnerability checks.
4. Package validation.
5. Compose/container validation.
6. Registry validation.
7. Release validation.

Each layer has different runtime requirements and should not be collapsed into one always-on command.

## Layer 1: Source Hygiene

Authoritative local command:

```bash
make verify
```

Expands to:

```text
go mod tidy -diff
go run ./tools/fmtcheck ./cmd ./internal ./tools
go test ./...
```

Purpose:

- Module tidiness check.
- Formatting check.
- Unit tests.

CI equivalent:

```bash
go mod tidy
gofmt -w ./cmd ./internal
git diff --exit-code
go test ./...
```

Known mismatch:

- CI's formatting scope excludes `tools`.
- CI uses mutating tidy/format commands plus diff.
- `make verify` uses non-mutating check targets.

## Layer 2: Static Analysis and Vulnerability Checks

Commands:

```bash
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

GitHub Actions also runs:

```bash
semgrep scan --config auto
```

Recommended use:

- Staticcheck before PRs.
- govulncheck for security-sensitive changes and releases.
- Semgrep in CI/security workflow.

## Layer 3: Package Quality

Command:

```bash
dockyard package lint PACKAGE_DIR --strict
```

Validates:

- Manifest loading and API version.
- Dependency metadata.
- Required package docs.
- Forbidden files and directories.
- Values loading.
- Schema validation.
- Schema quality.
- Default rendering.
- Policy lint findings.

Recommended baseline package:

```bash
dockyard package lint ./examples/nginx --strict
```

Dependency package baseline:

```bash
dockyard package lint ./examples/postgres --strict
```

## Layer 4: Package Test

Command:

```bash
dockyard package test PACKAGE_SOURCE --strict
```

Default package test validates:

- Source preparation.
- Package quality.
- Values/schema.
- Compose rendering.
- Policy checks.
- `docker compose config`.

This requires Docker Compose for the config validation step unless `--skip-compose-config` is passed.

Recommended baseline:

```bash
dockyard package test ./examples/nginx --strict
```

Dependency behavior baseline:

```bash
dockyard package test ./examples/postgres --strict
dockyard package deps ./examples/team-dashboard
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard
```

## Layer 5: Container Smoke Test

Command:

```bash
dockyard package test PACKAGE_SOURCE --smoke
```

Preflight:

```bash
dockyard doctor
```

Smoke test validates:

- Docker CLI availability.
- Docker Compose availability.
- Docker daemon reachability.
- Rendered Compose can start.
- `docker compose ps --all` works.
- Compose stack can be stopped.

Recommended smoke baseline:

```bash
dockyard doctor
dockyard package test ./examples/nginx --smoke
```

Use smoke tests only for examples that are safe to start and stop locally.

## Layer 6: Archive Validation

Commands:

```bash
dockyard lock ./examples/nginx
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
```

Validates:

- Lockfile matches rendered output when required.
- Archive structure.
- Unsafe paths.
- Forbidden files.
- Integrity hashes.
- Provenance when present.
- Manifest loading.
- Package build validation.

Recommended trigger:

- Archive code changes.
- Lockfile changes.
- Packaging docs changes.
- Release preparation.

## Layer 7: Registry Validation

Current repository has unit coverage for OCI reference validation and push argument construction, but no live registry CI gate.

Manual registry gate:

```bash
oras login ghcr.io
dockyard push ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz oci://ghcr.io/nandub/dockyard/nginx:0.1.0
dockyard pull oci://ghcr.io/nandub/dockyard/nginx:0.1.0 -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard package test oci://ghcr.io/nandub/dockyard/nginx:0.1.0 --strict
```

Recommended improvement:

- Add a non-production registry round-trip test path using a disposable tag or local registry.

## Layer 8: Release Validation

Pre-tag local validation:

```bash
go mod tidy
make verify
make dev-build
dockyard version
dockyard compat
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Example validation:

```bash
dockyard lock ./examples/nginx
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Optional smoke:

```bash
dockyard package test ./examples/nginx --smoke
```

GitHub release validation:

- `make verify`
- Staticcheck.
- Cross-platform builds.
- Linux AMD64 binary `version`.
- Linux AMD64 strict compat/package test for `examples/nginx`.
- SBOM.
- SHA256 checksums.

## Recommended Pipeline by Change Type

### Documentation-only Changes

```bash
make verify
```

If command behavior or package docs change, also run package gates.

### Internal Go Logic Changes

```bash
make verify
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

Add focused package/example gates when the logic affects rendering, values, policy, state, archive, OCI, catalog, or CLI behavior.

### Package or Example Changes

```bash
make verify
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Run the same gates on the changed example package.

### Dependency Install Behavior Changes

```bash
make verify
dockyard package lint ./examples/postgres --strict
dockyard package test ./examples/postgres --strict
dockyard package deps ./examples/team-dashboard
dockyard install-plan team-dashboard ./examples/team-dashboard
dockyard install --dry-run team-dashboard ./examples/team-dashboard
```

When Docker and registry dependencies are available, extend with `--with-dependencies` smoke testing in a controlled environment.

### Archive or Path-Safety Changes

```bash
make verify
dockyard lock ./examples/nginx
dockyard package ./examples/nginx --locked -o ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz
dockyard verify ../dockyard-artifacts/nginx-0.1.0.dockyard.tgz --require-lock
dockyard package test ./examples/nginx --strict
```

Recommended improvement:

- Add focused `internal/archive` unit tests for unsafe archive paths, symlinks, forbidden files, provenance, and checksum mismatches.

### Docker/Compose Runtime Changes

```bash
make verify
dockyard doctor
dockyard package test ./examples/nginx --strict
dockyard package test ./examples/nginx --smoke
```

If install/upgrade/uninstall behavior changes, run an out-of-repo smoke flow using generated packages and values.

### OCI/Catalog Changes

```bash
make verify
dockyard catalog list
dockyard package deps ./examples/team-dashboard
dockyard install-plan team-dashboard ./examples/team-dashboard
```

When ORAS and registry credentials are available, add push/pull round-trip validation.

### Release Candidate

```bash
go mod tidy
make verify
make dev-build
dockyard version
dockyard compat
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Then repeat non-smoke package gates for every public example and run smoke tests where Docker is available.

## Recommended CI Improvements

Consider aligning CI with local authority:

- Replace CI's direct tidy/gofmt sequence with `make verify`, or add `tools` to CI's formatting scope.
- Add package gates for `examples/nginx` to normal CI.
- Add matrix or loop validation for all public examples, at least for `compat`, `package lint`, and non-smoke `package test`.
- Add coverage reporting with `go test ./... -coverprofile=coverage.out`.
- Add focused archive tests under `internal/archive`.
- Add a Docker-enabled optional workflow for smoke tests.
- Add an ORAS/local-registry workflow for registry round trips.
- Add release workflow checks for more than Linux AMD64 where feasible.

## Validation Gaps to Track

- Benchmarks: none.
- Performance: none.
- Coverage gate: none.
- Live registry validation: manual only.
- Smoke tests: local/manual only.
- Archive test coverage: missing dedicated unit tests.
- Runner test coverage: missing dedicated unit tests.
- Upgrade/rollback end-to-end behavior: limited automated coverage.
- Full install/uninstall Docker lifecycle in CI: absent.

## Evidence

- `Makefile`
- `.github/workflows/ci.yml`
- `.github/workflows/security.yml`
- `.github/workflows/release.yml`
- `CONTRIBUTING.md`
- `README.md`
- `AGENTS.md`
- `docs/command-reference.md`
- `docs/packaging-and-distribution.md`
- `docs/release-candidate-checklist.md`
- `internal/cli/package.go`
- `internal/quality/quality.go`
- `internal/oci/oci_test.go`
- `internal/catalog/catalog_test.go`
- `internal/cli/install_plan_test.go`
- `internal/*/*_test.go`
