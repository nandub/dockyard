# Dockyard Architecture

This document records repository discovery findings for Dockyard. It is based on static inspection of the repository and does not assume unverified runtime behavior.

## Summary

Dockyard is a Go command-line application that adds package, render, validation, release-state, security-policy, dependency-planning, and distribution workflows on top of Docker Compose. Docker Compose remains the runtime source of truth.

Evidence:

- Go module: `github.com/nandub/dockyard` in `go.mod`.
- CLI entry point: `cmd/dockyard/main.go`.
- Root command wiring: `internal/cli/root.go`.
- Runtime delegation: `internal/runner/docker.go`.
- OCI delegation through ORAS: `internal/oci/oci.go` and `internal/catalog/catalog.go`.

## Runtime Model

Dockyard handles:

- Loading `Dockyard.yaml`.
- Loading and merging values.
- Validating values against `values.schema.json`.
- Rendering Docker Compose YAML.
- Optionally validating the rendered Compose file with `docker compose config`.
- Running package quality, compatibility, secrets, and policy checks.
- Creating and verifying `.dockyard.tgz` archives.
- Creating `dockyard.lock`.
- Pushing and pulling packages through OCI using `oras`.
- Recording local release state under Dockyard home.
- Delegating install, upgrade, rollback, status, and uninstall operations to Docker Compose.

Docker Compose handles:

- Compose file interpretation.
- Image pulls.
- Container, volume, and network lifecycle.
- Runtime status.

## Entry Points

Executable entry points:

- `cmd/dockyard/main.go`: production CLI binary.
- `tools/fmtcheck/main.go`: developer formatting check helper.

CLI root:

- `internal/cli/root.go` creates the `dockyard` Cobra root command and registers subcommands.

Registered command groups include:

- Core package operations: `init`, `lint`, `render`, `config`.
- Catalog operations: `catalog list`, `catalog info`.
- Release lifecycle: `install`, `install-plan`, `upgrade`, `rollback`, `uninstall`, `status`, `inspect`, `list`, `diff`, `prune`.
- Packaging and distribution: `lock`, `package`, `verify`, `push`, `pull`.
- Diagnostics and policy: `doctor`, `policy`, `secrets`, `compat`, `version`.
- Values and env helpers: `values`, `env`.

## Major Boundaries

### CLI Layer

Path: `internal/cli`

Purpose:

- Cobra command definitions.
- Command option parsing.
- User-facing output.
- Coordination of internal packages.

The CLI layer should stay thin; reusable logic belongs in internal packages.

### Package Manifest Layer

Path: `internal/dockpkg`

Purpose:

- Load and validate `Dockyard.yaml`.
- Enforce the stable manifest API version.
- Validate dependency metadata.
- Provide path-safety helpers.

Stable manifest API version:

- `dockyard.dev/v1alpha1`, defined in `internal/format/format.go`.

### Values Layer

Path: `internal/values`

Purpose:

- Load default and override values.
- Merge values.
- Validate values against JSON Schema.
- Generate values templates.

### Rendering Layer

Path: `internal/render`

Purpose:

- Render Compose YAML from package manifests and values.
- Produce diagnostics while masking sensitive-looking values.

### Policy and Quality Layers

Paths:

- `internal/policy`
- `internal/quality`
- `internal/cli/secrets.go`
- `internal/cli/compat.go`

Purpose:

- Security policy findings.
- Package quality gates.
- Secrets scanning.
- Format compatibility checks.

### State Layer

Path: `internal/state`

Purpose:

- Resolve Dockyard home.
- Validate release names.
- Read and write release state.
- Track current revisions.
- Store parent/dependency relationship metadata.

Dockyard home resolution order:

1. `--home`
2. `DOCKYARD_HOME`
3. `~/.dockyard`

### Docker Runner Layer

Path: `internal/runner`

Purpose:

- Shell out to `docker`.
- Check Docker CLI, Docker Compose, and daemon availability.
- Run `docker compose` lifecycle and config commands.

### Archive and Lock Layers

Paths:

- `internal/archive`
- `internal/lock`

Purpose:

- Create `.dockyard.tgz` archives.
- Add generated archive metadata such as `package.provenance.json` and `SHA256SUMS`.
- Verify archives.
- Reject unsafe archive paths and forbidden files.
- Create and verify `dockyard.lock`.

### OCI and Catalog Layers

Paths:

- `internal/oci`
- `internal/catalog`

Purpose:

- Normalize OCI references.
- Push and pull package archives using the external `oras` CLI.
- Load OCI-backed or file-backed catalog indexes.
- Resolve `catalog://NAME[:VERSION]` and bare catalog shorthand.
- Cache catalog metadata under `~/.dockyard/cache/catalogs`.

## Formats

Centralized in `internal/format/format.go`.

Stable:

- `Dockyard.yaml`: `dockyard.dev/v1alpha1`.

Experimental:

- `dockyard.lock`: `dockyard.dev/lockfile/v1alpha1`.
- `package.provenance.json`: `dockyard.dev/provenance/v1alpha1`.
- `release.json`: `dockyard.dev/release/v1alpha1`.

## External Requirements

Required for normal lifecycle/runtime workflows:

- Go toolchain for development.
- Docker CLI.
- Docker Compose plugin.
- Reachable Docker daemon.

Required only for OCI/catalog workflows:

- ORAS CLI.
- External registry authentication managed outside Dockyard.

## Repository Outputs

Generated or output locations:

- `bin/`: local binaries from `make build` and `make release-snapshot`.
- `dist/`: release workflow artifacts.
- `.dockyard/`: local Dockyard state if accidentally created in the repo.
- `*.dockyard.tgz`: package archives.
- `SHA256SUMS`: release/package checksum output.
- `coverage.out`: test coverage output.

The repository documentation recommends generated packages, deployment values, and artifacts live outside the repo:

- `../dockyard-work/`
- `../deploy-values/`
- `../dockyard-artifacts/`

## Unverified Items

The following were not verified by execution during discovery:

- `go test ./...`
- `make verify`
- `dockyard doctor`
- Docker daemon availability
- ORAS availability
- End-to-end package smoke tests
