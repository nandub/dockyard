# Dockyard Component Map

This document maps repository components by purpose.

## Root Files

Repository instructions and docs:

- `AGENTS.md`: coding-agent repository guidance.
- `README.md`: user-facing overview, local development, quick start, packaging example.
- `CONTRIBUTING.md`: contribution workflow and pre-PR checks.
- `SECURITY.md`: vulnerability and security posture.
- `CHANGELOG.md`: release history.
- `CODE_OF_CONDUCT.md`: community conduct.

Build and dependency configuration:

- `go.mod`: Go module and dependency declarations.
- `go.sum`: Go dependency checksums.
- `Makefile`: build, test, format, verify, release-snapshot, and clean targets.
- `.editorconfig`: editor formatting conventions.
- `.gitignore`: ignored generated, sensitive, and output files.

## Executables

### Dockyard CLI

Path:

- `cmd/dockyard/main.go`

Purpose:

- Starts the Cobra CLI by calling `cli.NewRootCommand().Execute()`.

### fmtcheck Tool

Path:

- `tools/fmtcheck/main.go`

Purpose:

- Walks Go source roots and exits non-zero if files are not gofmt-formatted.
- Used by `make fmt-check`.

## CLI Components

Path:

- `internal/cli`

Files by command or concern:

- `root.go`: root command and command registration.
- `init.go`: starter package creation.
- `lint.go`: package linting.
- `render.go`: render Compose YAML.
- `config.go`: render and run `docker compose config`.
- `catalog.go`: catalog list/info commands.
- `install.go`: release install and dependency install flow.
- `install_plan.go`: read-only dependency-aware install planning.
- `uninstall.go`: release uninstall.
- `list.go`: release listing.
- `status.go`: release status and optional Compose status.
- `inspect.go`: release inspection.
- `diff.go`, `diffutil.go`: release/package diffing.
- `upgrade.go`: release upgrade.
- `rollback.go`: release rollback.
- `doctor.go`: prerequisite checks.
- `lock.go`: lockfile generation.
- `values.go`: values template, validate, schema commands.
- `package.go`: package archive, dependency, lint, and test commands.
- `verify.go`: package archive verification.
- `push.go`: OCI push.
- `pull.go`: OCI pull.
- `policy.go`: policy list/check commands.
- `secrets.go`: secrets scan command.
- `env.go`: env template/check commands.
- `prune.go`: release revision pruning.
- `version.go`: version command.
- `compat.go`: compatibility command.
- `common.go`: shared CLI helpers.

CLI tests:

- `catalog_source_test.go`
- `install_plan_test.go`
- `list_test.go`
- `uninstall_test.go`

## Internal Libraries

### archive

Path:

- `internal/archive`

Purpose:

- Create `.dockyard.tgz` packages.
- Generate package provenance and checksum metadata.
- Verify archives.
- Extract archives safely.

Tests:

- No `archive_test.go` found.

### catalog

Path:

- `internal/catalog`

Purpose:

- Load catalog indexes from file paths, `file://`, or OCI.
- Resolve catalog package shorthand.
- Cache pulled catalog metadata.

Tests:

- `catalog_test.go`

### dockpkg

Path:

- `internal/dockpkg`

Purpose:

- Load and validate `Dockyard.yaml`.
- Validate dependency metadata.
- Enforce path containment.

Tests:

- `package_test.go`

### envfile

Path:

- `internal/envfile`

Purpose:

- Generate and validate dotenv templates/files.

Tests:

- `envfile_test.go`

### format

Path:

- `internal/format`

Purpose:

- Central format API versions and stability metadata.

Tests:

- No `format_test.go` found.

### lock

Path:

- `internal/lock`

Purpose:

- Create and verify `dockyard.lock`.

Tests:

- `lock_test.go`

### oci

Path:

- `internal/oci`

Purpose:

- Normalize OCI refs.
- Push and pull archives through `oras`.

Tests:

- `oci_test.go`

### policy

Path:

- `internal/policy`

Purpose:

- Compose security policy checks.

Tests:

- `policy_test.go`

### quality

Path:

- `internal/quality`

Purpose:

- Package quality checks for publication readiness.

Tests:

- `quality_test.go`

### render

Path:

- `internal/render`

Purpose:

- Render Compose YAML and collect diagnostics.

Tests:

- `render_test.go`

### runner

Path:

- `internal/runner`

Purpose:

- Docker and Docker Compose subprocess wrapper.

Tests:

- No `runner_test.go` found.

### state

Path:

- `internal/state`

Purpose:

- Dockyard home resolution.
- Release state read/write.
- Current revision tracking.

Tests:

- `state_test.go`

### values

Path:

- `internal/values`

Purpose:

- Values loading, merging, JSON Schema validation, and template generation.

Tests:

- `template_test.go`

### version

Path:

- `internal/version`

Purpose:

- Build-time version metadata.

Tests:

- No `version_test.go` found.

## Example Packages

Each example package contains:

- `Dockyard.yaml`
- `compose.yaml`
- `values.yaml`
- `values.schema.json`
- `README.md`
- `SECURITY.md`
- `LICENSE`

Examples:

- `examples/nginx`: runnable nginx starter package; also contains tracked `dockyard.lock`.
- `examples/postgres`: reusable PostgreSQL dependency package.
- `examples/postgres-app`: app plus PostgreSQL package using env-backed secret values.
- `examples/team-dashboard`: dependency-metadata example depending on PostgreSQL.
- `examples/caddy-letsencrypt`: Caddy automatic HTTPS reverse proxy example.
- `examples/nginx-tls-mounted-certs`: nginx TLS with mounted certificate/key paths.
- `examples/traefik-letsencrypt`: Traefik Let's Encrypt example with Docker provider labels and whoami service.

## Documentation Folder

Path:

- `docs`

Files:

- `command-reference.md`
- `compose-compatibility.md`
- `getting-started.md`
- `operator-guide.md`
- `packaging-and-distribution.md`
- `real-world-example.md`
- `release-candidate-checklist.md`
- `release-engineering.md`
- `security.md`
- `support-policy.md`
- `upgrade-policy.md`
- `v1-readiness.md`

## GitHub Configuration

Templates:

- `.github/PULL_REQUEST_TEMPLATE.md`
- `.github/ISSUE_TEMPLATE/bug_report.md`
- `.github/ISSUE_TEMPLATE/feature_request.md`

Workflows:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/workflows/security.yml`

## Local Non-Source Directories

Present locally:

- `.agents/`: no files found during discovery.
- `.codex/`: no files found during discovery.
- `bin/`: ignored local build outputs; contains built Dockyard binaries.

Ignored by `.gitignore` but not found in the repository scan:

- `dist/`
- `.dockyard/`
- `deploy-values/`
- `dockyard-work/`
- `dockyard-artifacts/`
- `example-app/`
- `team-dashboard/`
