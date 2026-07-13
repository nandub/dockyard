# AGENTS.md

This is the primary entry point for AI agents working in Dockyard.

## Repository Purpose

Dockyard is a Go CLI that adds package, render, validation, release-state, policy, dependency-planning, catalog, and OCI distribution workflows on top of Docker Compose. Docker Compose remains the runtime source of truth.

Observed evidence:

- Module: `github.com/nandub/dockyard` in `go.mod`.
- CLI entry point: `cmd/dockyard/main.go`.
- Cobra command wiring: `internal/cli/`.
- Docker Compose delegation: `internal/runner/`.
- OCI operations through the embedded ORAS Go client: `internal/oci/` and `internal/catalog/`.

Dockyard is a package/deployment manager for Docker Compose workflows.

## Sources of Truth

Use this order when sources conflict:

1. Current source code.
2. Automated tests.
3. CI workflows.
4. Executable scripts and Makefile targets.
5. Current technical documentation.
6. Historical onboarding material under `.ai/onboarding/`.

Conflicts must be reported instead of silently resolved.

## Evidence Policy

Classify claims explicitly when accuracy matters:

- **Verified** - confirmed by executing the relevant behavior or command in the current work.
- **Observed** - directly supported by source, configuration, tests, CI, or current docs.
- **Inferred** - derived from indirect evidence and not yet verified.
- **Unknown** - insufficient evidence.

Do not present inferred behavior as verified.

## Required Initial Actions

Before changing code or behavior:

1. Inspect `git status --short`.
2. Read this root `AGENTS.md`.
3. Read any more specific nested `AGENTS.md` in the affected path.
4. Read the relevant playbook under `.ai/playbooks/`.
5. Inspect the affected implementation and tests.
6. Determine whether an execution plan is required by `PLANS.md`.
7. Identify the smallest relevant validation commands.

## Playbook Routing

Use only playbooks that exist under `.ai/playbooks/`:

- Onboarding and repository discovery: `.ai/playbooks/onboarding.md`
- Feature work: `.ai/playbooks/feature.md`
- Bug fixes: `.ai/playbooks/bugfix.md`
- Refactoring: `.ai/playbooks/refactor.md`
- Testing and validation: `.ai/playbooks/testing.md`
- Dependency upgrades: `.ai/playbooks/dependency-upgrade.md`
- Security review: `.ai/playbooks/security-review.md`
- Performance work: `.ai/playbooks/performance.md`
- CLI changes: `.ai/playbooks/cli-change.md`
- OCI package changes: `.ai/playbooks/oci-package.md`
- Catalog changes: `.ai/playbooks/catalog-change.md`
- Releases: `.ai/playbooks/release.md`
- Documentation updates: `.ai/playbooks/documentation.md`
- Repository cleanup: `.ai/playbooks/repository-cleanup.md`

The `.ai/` directory is repository documentation for humans and AI tools. It is not a Codex plugin, skill, or automatically loaded workflow system.

## Engineering Rules

- Preserve public behavior unless the user explicitly authorizes a change.
- Avoid unrelated refactoring.
- Reuse existing abstractions and package boundaries.
- Keep CLI code thin; prefer testable logic in `internal/*`.
- Add or update tests for behavior changes.
- Do not weaken validation, policy checks, path safety, archive safety, or secret handling.
- Do not manually modify generated files, vendored files, package archives, binaries, local state, or local smoke-test artifacts.
- Do not expose credentials, tokens, secret values, or private key material in output, logs, docs examples, or errors.
- Keep OCI credentials outside Dockyard-owned state; current behavior reads Docker-compatible registry credentials through the embedded ORAS Go client.
- Treat package archives and package paths as untrusted input.
- Preserve Windows path handling; use `filepath` for filesystem paths and `path` only for archive-internal slash paths.
- Report commands that could not be run.
- Document remaining uncertainty.

## Validation Commands

Use the smallest validation set that matches the change.

Common commands observed in this repository:

```sh
go mod tidy
go fmt ./...
go test ./...
go build -o bin/dockyard$(go env GOEXE) ./cmd/dockyard
```

Makefile targets:

```sh
make tidy
make fmt
make fmt-check
make tidy-check
make test
make verify
make build
make clean
```

Package validation commands, when a built `dockyard` binary is available:

```sh
dockyard compat ./examples/nginx --strict
dockyard package lint ./examples/nginx --strict
dockyard package test ./examples/nginx --strict
```

Runtime smoke testing requires Docker Desktop or Docker Engine and should keep generated packages, values, archives, and rendered files outside this repository.

Do not invent a single universal validation command. Choose based on the affected files and report anything not run.

## Completion Report

Every completed task should report:

- Summary.
- Files changed.
- Design decisions.
- Tests changed.
- Validation commands and results.
- Unverified items.
- Remaining risks.

## Documentation Map

- Durable documentation: `docs/index.md`.
- AI operating docs: `.ai/README.md`.
- Execution-plan rules: `PLANS.md`.
- Historical onboarding reports: `.ai/onboarding/`.
