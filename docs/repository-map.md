# Repository Map

This map summarizes observed repository structure. Historical discovery detail is preserved under `.ai/onboarding/reports/`.

## Root

- `AGENTS.md` - primary AI entry point.
- `PLANS.md` - execution-plan requirements.
- `README.md` - user-facing overview and quick start.
- `CHANGELOG.md` - release history.
- `SECURITY.md` - vulnerability reporting.
- `CONTRIBUTING.md` - contribution workflow.
- `go.mod`, `go.sum` - Go module and dependency checksums.
- `Makefile` - local development targets.

## Source

- `cmd/dockyard/main.go` - CLI process entry point.
- `internal/cli/` - Cobra command wiring and CLI behavior.
- `internal/dockpkg/` - package manifest loading and package path safety.
- `internal/values/` - values loading, merge, defaults, and JSON Schema validation.
- `internal/render/` - Compose template rendering and diagnostics.
- `internal/policy/` - selected Compose security and compatibility policy checks.
- `internal/state/` - Dockyard home, releases, revisions, and metadata.
- `internal/runner/` - Docker and Docker Compose subprocess integration.
- `internal/archive/` - `.dockyard.tgz` archive creation and verification.
- `internal/lock/` - `dockyard.lock` creation and verification.
- `internal/oci/` - ORAS-based OCI push and pull.
- `internal/catalog/` - catalog source resolution and metadata loading.
- `internal/envfile/` - env-file template and validation helpers.
- `internal/quality/` - package quality validation.
- `internal/format/` - format/API version constants.
- `internal/version/` - version metadata.

## Tools, Tests, and Examples

- `tools/fmtcheck/` - formatting check helper used by `make fmt-check`.
- `internal/**/*_test.go` - Go tests.
- `examples/` - Dockyard package examples.

## Documentation and AI Material

- `docs/` - durable user and technical documentation.
- `.ai/` - AI operating docs, playbooks, prompts, templates, checklists, and onboarding history.

## Workflows

- `.github/workflows/ci.yml` - CI build/test workflow.
- `.github/workflows/security.yml` - security/static analysis workflow.
- `.github/workflows/release.yml` - release artifact workflow.

## Generated, Artifact, or Local-Only Paths

Do not edit or commit these unless explicitly requested:

- `bin/`
- `dist/`
- `coverage.out`
- `*.dockyard.tgz`
- `.dockyard/`
- local smoke-test directories such as `../dockyard-work/`, `../deploy-values/`, and `../dockyard-artifacts/`.
